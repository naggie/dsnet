package dsnet

import (
	"net"
)

func Add(hostname string, owner string, description string) { //, publicKey string) {
	conf := MustLoadDsnetConfig()

	privateKey := GenerateJSONPrivateKey()
	presharedKey := GenerateJSONKey()
	publicKey := privateKey.PublicKey()

	IP := conf.MustChooseIP()
	check(err)

	peer := PeerConfig{
		Owner:        owner,
		Hostname:     hostname,
		Description:  description,
		PublicKey:    publicKey,
		PresharedKey: presharedKey,
		// TODO Endpoint:
		AllowedIPs: []JSONIPNet{
			JSONIPNet{
				IPNet: net.IPNet{
					IP:   IP,
					Mask: conf.Network.IPNet.Mask,
				},
			},
		},
	}

	conf.MustAddPeer(peer)
	conf.MustSave()
}
