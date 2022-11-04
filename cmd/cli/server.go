package cli

import (
	"github.com/naggie/dsnet/lib"
	"github.com/spf13/viper"
)

func GetServer(config *DsnetConfig) *lib.Server {
	fallbackWGBin := viper.GetString("fallback_wg_bin")
	return &lib.Server{
		ExternalHostname:    config.ExternalHostname,
		ExternalIP:          config.ExternalIP,
		ExternalIP6:         config.ExternalIP6,
		ListenPort:          config.ListenPort,
		Domain:              config.Domain,
		InterfaceName:       config.InterfaceName,
		Network:             config.Network,
		Network6:            config.Network6,
		IP:                  config.IP,
		IP6:                 config.IP6,
		DNS:                 config.DNS,
		PrivateKey:          config.PrivateKey,
		PostUp:              config.PostUp,
		PostDown:            config.PostDown,
		FallbackWGBin:       fallbackWGBin,
		Peers:               jsonPeerToDsnetPeer(config.Peers),
		Networks:            config.Networks,
		PersistentKeepalive: config.PersistentKeepalive,
	}
}
