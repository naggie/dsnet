package dsnet

import (
	"github.com/vishvananda/netlink"
)

func Down() {
	conf := MustLoadDsnetConfig()
	DelLink(conf)
	RunPostDown(conf)
}

func RunPostDown(conf *DsnetConfig) {
	ShellOut(conf.PostDown, "PostDown")
}

func DelLink(conf *DsnetConfig) {
	linkAttrs := netlink.NewLinkAttrs()
	linkAttrs.Name = conf.InterfaceName

	link := &netlink.GenericLink{
		LinkAttrs: linkAttrs,
	}

	err := netlink.LinkDel(link)
	if err != nil {
		ExitFail("Could not delete interface '%s' (%v)", conf.InterfaceName, err)
	}
}
