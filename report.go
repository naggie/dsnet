package dsnet

import (
	"net"

	"github.com/vishvananda/netlink"
	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

func Up() {
	conf := MustLoadDsnetConfig()

	dev, err := wgctrl.Device(conf.InterfaceName)
	check(err)

	report := Report(dev, conf)
	report.MustSave()
}
