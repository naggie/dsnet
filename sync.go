package dsnet

import (
	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

func Sync() {
	// TODO check device settings first
	conf := MustLoadDsnetConfig()
	ConfigureDevice(conf)
}

func ConfigureDevice(conf *DsnetConfig) {
	peers := conf.GetWgPeerConfigs()

	wgConfig := wgtypes.Config{
		PrivateKey:   &conf.PrivateKey.Key,
		ListenPort:   &conf.ListenPort,
		ReplacePeers: true,
		Peers:        peers,
	}

	wg, err := wgctrl.New()
	check(err)
	defer wg.Close()

	err = wg.ConfigureDevice(conf.InterfaceName, wgConfig)

	if err != nil {
		ExitFail("Could not configure device '%s' (%v)", conf.InterfaceName, err)
	}
}
