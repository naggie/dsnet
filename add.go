package dsnet

import (
	"net"
	"os"
	"text/template"
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
	PrintPeerCfg(peer, conf)
	conf.MustSave()
}

func PrintPeerCfg(peer PeerConfig, conf *DsnetConfig) {
	const peerConf = `[Interface]
Address = {{ index .Peer.AllowedIPs 0 }}
PrivateKey={{ .Peer.PrivateKey.Key }}
PresharedKey={{ .Peer.PresharedKey.Key }}
DNS = {{ .DsnetConfig.InternalDNS }}

[Peer]
PublicKey={{ .DsnetConfig.PrivateKey.PublicKey.Key }}
PresharedKey={{ .DsnetConfig.PresharedKey.Key }}
Endpoint={{ .DsnetConfig.ExternalIP }}:{{ .DsnetConfig.ListenPort }}
#AllowedIPs=0.0.0.0/0
AllowedIPs={{ .DsnetConfig.Network }}
PersistentKeepalive=21
`

	t := template.Must(template.New("peerConf").Parse(peerConf))
	err := t.Execute(os.Stdout, map[string]interface{}{
		"Peer":        peer,
		"DsnetConfig": conf,
	})
	check(err)
}
