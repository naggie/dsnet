package dsnet

import (
	"net"
)

func Add(hostname string, owner string, description string) { //, publicKey string) {
	conf := MustLoadDsnetConfig()

	privateKey := GenerateJSONPrivateKey()
	publicKey := privateKey.PublicKey()

	IP := conf.MustAllocateIP()

	peer := PeerConfig{
		Owner:        owner,
		Hostname:     hostname,
		Description:  description,
		PublicKey:    publicKey,
		PrivateKey:   privateKey, // omitted from server config JSON!
		PresharedKey: GenerateJSONKey(),
		AllowedIPs: []JSONIPNet{
			JSONIPNet{
				IPNet: net.IPNet{
					IP:   IP,
					Mask: net.CIDRMask(32, 32),
				},
			},
		},
	}

	conf.MustAddPeer(peer)
	conf.MustSave()
}

func GetPeerWgQuickConf(peer PeerConfig) string {
	return `[Interface]
Address = 10.50.60.2/24
PrivateKey={{
DNS = 8.8.8.8

[Peer]
PublicKey=cAR+SMd+yvGw2TVzVSRoLtxF5TLA2Y/ceebO8ZAyITw=
Endpoint=3.9.82.135:51820
AllowedIPs=0.0.0.0/0
PersistentKeepalive=21
`
}
