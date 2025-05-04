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

		if ! exists {
			err = netlink.AddrAdd(link, addr)
			if err != nil {
				return fmt.Errorf("could not add ipv4 addr %s to interface %s: %v", addr.IP, s.InterfaceName, err)
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
	}

	// bring up interface (UNKNOWN state instead of UP, a wireguard quirk)
	err = netlink.LinkSetUp(link)

	if err != nil {
		return fmt.Errorf("could not bring up device '%s' (%v)", s.InterfaceName, err)
	}
	return nil
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
