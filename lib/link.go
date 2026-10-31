package lib

import (
	"errors"
	"fmt"
	"net"

	"github.com/naggie/dsnet/utils"
	"github.com/vishvananda/netlink"
)

// IPExists checks if the given IP address already exists on the specified link
func IPExists(link netlink.Link, ipNet *net.IPNet, family int) (bool, error) {
	addrs, err := netlink.AddrList(link, family)
	if err != nil {
		return false, fmt.Errorf("failed to list addresses for interface: %v", err)
	}
	for _, addr := range addrs {
		if addr.IPNet.String() == ipNet.String() {
			return true, nil
		}
	}
	return false, nil
}

// CreateLink sets up the WG interface and link with the correct
// address
func (s *Server) CreateLink() error {
	if len(s.IP) == 0 && len(s.IP6) == 0 {
		return errors.New("no IPv4 or IPv6 ip defined in config")
	}

	linkAttrs := netlink.NewLinkAttrs()
	linkAttrs.Name = s.InterfaceName

	link := &netlink.GenericLink{
		LinkAttrs: linkAttrs,
		LinkType:  "wireguard",
	}

	err := netlink.LinkAdd(link)
	if err != nil && s.FallbackWGBin != "" {
		// return fmt.Errorf("could not add interface '%s' (%v), falling back to the userspace implementation", s.InterfaceName, err)
		cmdStr := fmt.Sprintf("%s %s", s.FallbackWGBin, s.InterfaceName)
		if err = utils.ShellOut(cmdStr, "Userspace implementation"); err != nil {
			return fmt.Errorf("failed to start userspace wireguard: %s", err)
		}
	}

	if len(s.IP) != 0 {
		addr := &netlink.Addr{
			IPNet: &net.IPNet{
				IP:   s.IP,
				Mask: s.Network.IPNet.Mask,
			},
		}

		exists, err := IPExists(link, addr.IPNet, netlink.FAMILY_V4)
		if err != nil {
			return err
		}

		if !exists {
			err = netlink.AddrAdd(link, addr)
			if err != nil {
				return fmt.Errorf("could not add ipv4 addr %s to interface %s: %v", addr.IP, s.InterfaceName, err)
			}
		}

		// remove any other IPs on the interface
		addrs, err := netlink.AddrList(link, netlink.FAMILY_V4)
		if err != nil {
			return fmt.Errorf("failed to list addresses for interface: %v", err)
		}
		for _, existing := range addrs {
			if existing.IPNet.String() != addr.IPNet.String() {
				err := netlink.AddrDel(link, &existing)
				if err != nil {
					return fmt.Errorf("failed to delete address %s from interface %s: %v", existing.IP, s.InterfaceName, err)
				}
			}
		}
	}

	if len(s.IP6) != 0 {
		addr6 := &netlink.Addr{
			IPNet: &net.IPNet{
				IP:   s.IP6,
				Mask: s.Network6.IPNet.Mask,
			},
		}

		exists, err := IPExists(link, addr6.IPNet, netlink.FAMILY_V6)
		if err != nil {
			return err
		}

		if !exists {
			err = netlink.AddrAdd(link, addr6)
			if err != nil {
				return fmt.Errorf("could not add ipv6 addr %s to interface %s: %v. Do you have IPv6 enabled?", addr6.IP, s.InterfaceName, err)
			}
		}

		// remove any other IPs on the interface
		addrs, err := netlink.AddrList(link, netlink.FAMILY_V6)
		if err != nil {
			return fmt.Errorf("failed to list v6 addresses for interface: %v", err)
		}

		for _, addr := range addrs {
			if addr.IPNet.String() != addr6.IPNet.String() {
				err := netlink.AddrDel(link, &addr)
				if err != nil {
					return fmt.Errorf("failed to delete address %s from interface %s: %v", addr.IP, s.InterfaceName, err)
				}
			}
		}
	}

	// set MTU
	err = netlink.LinkSetMTU(link, s.MTU)
	if err != nil {
		return fmt.Errorf("failed to set MTU %d on interface %s: %v", s.MTU, s.InterfaceName, err)
	}

	// bring up interface (UNKNOWN state instead of UP, a wireguard quirk)
	err = netlink.LinkSetUp(link)

	if err != nil {
		return fmt.Errorf("could not bring up device '%s' (%v)", s.InterfaceName, err)
	}
	return nil
}

// dsnetRouteProtocol tags the routes dsnet adds for peer Networks, so a sync
// only removes its own stale routes, never ones added by the kernel, PostUp or
// an admin. 213 is unassigned in rtnetlink.h and iproute2's rt_protos.
const dsnetRouteProtocol = 213

// routedNetworks returns the Networks behind peers, keyed by CIDR. Peer IPs
// need no route: the interface address already has a connected route.
// Paused peers keep their routes so their traffic is dropped by WireGuard
// instead of leaking out via another route.
func (s *Server) routedNetworks() map[string]net.IPNet {
	networks := make(map[string]net.IPNet)
	for _, peer := range s.Peers {
		for _, network := range peer.Networks {
			if len(network.IPNet.IP) == 0 {
				continue
			}
			// the kernel rejects a route with host bits set, and config
			// parsing keeps them (10.1.2.3/16)
			dst := net.IPNet{
				IP:   network.IPNet.IP.Mask(network.IPNet.Mask),
				Mask: network.IPNet.Mask,
			}
			networks[dst.String()] = dst
		}
	}
	return networks
}

// SyncRoutes routes peer Networks via the interface and removes routes for
// Networks no longer configured
func (s *Server) SyncRoutes() error {
	link, err := netlink.LinkByName(s.InterfaceName)
	if err != nil {
		return fmt.Errorf("failed to get interface(%s): %v", s.InterfaceName, err)
	}

	// every route on the interface, not just dsnet's own, so routes already
	// added by hand or by PostUp are adopted
	existing, err := netlink.RouteListFiltered(
		netlink.FAMILY_ALL,
		&netlink.Route{LinkIndex: link.Attrs().Index},
		netlink.RT_FILTER_OIF,
	)
	if err != nil {
		return fmt.Errorf("failed to list routes for interface %s: %v", s.InterfaceName, err)
	}

	toAdd, toDelete := planRoutes(existing, s.routedNetworks())

	for _, route := range toDelete {
		route := route
		if err := netlink.RouteDel(&route); err != nil {
			return fmt.Errorf("failed to delete route %s from interface %s: %v", route.Dst, s.InterfaceName, err)
		}
	}

	// RouteAdd, not RouteReplace: a peer network clashing with a route on
	// another interface (a LAN, the default route) must fail loudly, not
	// hijack it
	for _, dst := range toAdd {
		dst := dst
		route := &netlink.Route{
			LinkIndex: link.Attrs().Index,
			Dst:       &dst,
			Scope:     netlink.SCOPE_LINK,
			Protocol:  dsnetRouteProtocol,
		}
		if err := netlink.RouteAdd(route); err != nil {
			return fmt.Errorf("could not add route %s to interface %s: %v", dst.String(), s.InterfaceName, err)
		}
	}
	return nil
}

// planRoutes compares the routes on the interface with the wanted Networks.
// Any existing route to a wanted Network satisfies it, whoever added it. Only
// routes dsnet added are deleted once unwanted; others are left alone.
func planRoutes(existing []netlink.Route, wanted map[string]net.IPNet) ([]net.IPNet, []netlink.Route) {
	present := make(map[string]bool)
	toDelete := make([]netlink.Route, 0)

	for _, route := range existing {
		if _, ok := wanted[route.Dst.String()]; ok {
			present[route.Dst.String()] = true
		} else if route.Protocol == dsnetRouteProtocol {
			toDelete = append(toDelete, route)
		}
	}

	toAdd := make([]net.IPNet, 0)
	for cidr, dst := range wanted {
		if !present[cidr] {
			toAdd = append(toAdd, dst)
		}
	}

	return toAdd, toDelete
}

// DeleteLink removes the Netlink interface if it exists
func (s *Server) DeleteLink() error {
	link, err := netlink.LinkByName(s.InterfaceName)
	if err != nil {
		// If the error is because the link doesn't exist, just return nil
		if _, ok := err.(netlink.LinkNotFoundError); ok {
			return nil
		}
		return fmt.Errorf("failed to get interface(%s): %v", s.InterfaceName, err)
	}

	err = netlink.LinkDel(link)
	if err != nil {
		return fmt.Errorf("failed to delete interface(%s): %v", s.InterfaceName, err)
	}
	return nil
}
