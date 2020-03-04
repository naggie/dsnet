package dsnet

import (
	"net"

	"github.com/vishvananda/netlink"
	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

func Up() {
	conf := MustLoadDsnetConfig()
	CreateInterface(conf)
}

func CreateInterface(conf *DsnetConfig) {
	linkAttrs := netlink.NewLinkAttrs()
	linkAttrs.Name = conf.InterfaceName

	iface := &netlink.GenericLink{
		LinkAttrs: linkAttrs,
		LinkType:  "wireguard",
	}

	err := netlink.LinkAdd(iface)
	if err != nil {
		ExitFail("Could not add '%s' (%v)", linkAttrs.Name, err)
	}

	addr := &netlink.Addr{
		IPNet: &net.IPNet{
			IP:   conf.IP,
			Mask: conf.Network.IPNet.Mask,
		},
	}

	err = netlink.AddrAdd(iface, addr)
	if err != nil {
		ExitFail("Could not add addr %s to interface %s", addr.IP, err)
	}

	deviceConfig := wgtypes.Config{
		PrivateKey: &conf.PrivateKey.Key,
		ListenPort: &conf.ListenPort,
	}

	wg, err := wgctrl.New()
	check(err)

	err = wg.ConfigureDevice(linkAttrs.Name, deviceConfig)

	if err != nil {
		ExitFail("Could not configure device '%s' (%v)", linkAttrs.Name, err)
	}
}
