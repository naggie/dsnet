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

	peer := PeerConfig{
		Owner:        owner,
		Hostname:     hostname,
		Description:  description,
		PublicKey:    publicKey,
		PresharedKey: presharedKey,
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


func GetPeerWgQuickConf(peer PeerConfig, privKey JSONKey) string {
	return `[Interface]
Address = 10.50.60.2/24
PrivateKey=REDACTED
DNS = 8.8.8.8

[Peer]
PublicKey=cAR+SMd+yvGw2TVzVSRoLtxF5TLA2Y/ceebO8ZAyITw=
Endpoint=3.9.82.135:51820
AllowedIPs=0.0.0.0/0
PersistentKeepalive=21
`
}
