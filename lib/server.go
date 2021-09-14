package lib

import (
	"net"

	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

type Server struct {
	ExternalHostname string
	ExternalIP       net.IP
	ExternalIP6      net.IP
	ListenPort       int
	Domain           string
	InterfaceName    string
	Network          JSONIPNet
	Network6         JSONIPNet
	IP               net.IP
	IP6              net.IP
	DNS              net.IP
	PrivateKey       JSONKey
	PostUp           string
	PostDown         string
	FallbackWGBin    string
	Peers            []Peer
}

func (s *Server) GetPeers() []wgtypes.PeerConfig {
	wgPeers := make([]wgtypes.PeerConfig, 0, len(s.Peers))

	for _, peer := range s.Peers {
		// create a new PSK in memory to avoid passing the same value by
		// pointer to each peer (d'oh)
		presharedKey := peer.PresharedKey.Key

		// AllowedIPs = private IP + defined networks
		allowedIPs := make([]net.IPNet, 0, len(peer.Networks)+2)

		if len(peer.IP) > 0 {
			allowedIPs = append(
				allowedIPs,
				net.IPNet{
					IP:   peer.IP,
					Mask: net.IPMask{255, 255, 255, 255},
				},
			)
		}

		if len(peer.IP6) > 0 {
			allowedIPs = append(
				allowedIPs,
				net.IPNet{
					IP:   peer.IP6,
					Mask: net.IPMask{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
				},
			)
		}

		for _, net := range peer.Networks {
			allowedIPs = append(allowedIPs, net.IPNet)
		}

		wgPeers = append(wgPeers, wgtypes.PeerConfig{
			PublicKey:         peer.PublicKey.Key,
			Remove:            false,
			UpdateOnly:        false,
			PresharedKey:      &presharedKey,
			Endpoint:          nil,
			ReplaceAllowedIPs: true,
			AllowedIPs:        allowedIPs,
		})
	}

	return wgPeers
}
