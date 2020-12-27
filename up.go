package dsnet

import (
	"net"

	"github.com/vishvananda/netlink"
)

func Up() {
	conf := MustLoadDsnetConfig()
	CreateLink(conf)
	ConfigureDevice(conf)
	RunPostUp(conf)
}

func RunPostUp(conf *DsnetConfig) {
	ShellOut(conf.PostUp, "PostUp")
}

// CreateLink sets up the WG interface and link with the correct
// address
func CreateLink(conf *DsnetConfig) {
	linkAttrs := netlink.NewLinkAttrs()
	linkAttrs.Name = conf.InterfaceName

	link := &netlink.GenericLink{
		LinkAttrs: linkAttrs,
		LinkType:  "wireguard",
	}

	err := netlink.LinkAdd(link)
	if err != nil {
		ExitFail("Could not add interface '%s' (%v)", conf.InterfaceName, err)
	}

	if len(conf.IP) == 0 && len(conf.IP6) == 0 {
		ExitFail("No IPv4 or IPv6 network defined in config")
	}

	if len(conf.IP) != 0 {
		addr := &netlink.Addr{
			IPNet: &net.IPNet{
				IP:   conf.IP,
				Mask: conf.Network.IPNet.Mask,
			},
		}

		err = netlink.AddrAdd(link, addr)
		if err != nil {
			ExitFail("Could not add ipv4 addr %s to interface %s", addr.IP, err)
		}
	}

	if len(conf.IP6) != 0 {
		addr6 := &netlink.Addr{
			IPNet: &net.IPNet{
				IP:   conf.IP6,
				Mask: conf.Network6.IPNet.Mask,
			},
		}

		err = netlink.AddrAdd(link, addr6)
		if err != nil {
			ExitFail("Could not add ipv6 addr %s to interface %s", addr6.IP, err)
		}
	}

	// bring up interface (UNKNOWN state instead of UP, a wireguard quirk)
	err = netlink.LinkSetUp(link)

	if err != nil {
		ExitFail("Could not bring up device '%s' (%v)", conf.InterfaceName, err)
	}
}
