package dsnet

import (
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
		IP:           IP,
	}

	conf.MustAddPeer(peer)
	PrintPeerCfg(peer, conf)
	conf.MustSave()
}

func PrintPeerCfg(peer PeerConfig, conf *DsnetConfig) {
	const peerConf = `[Interface]
Address = {{ .Peer.IP }}
PrivateKey={{ .Peer.PrivateKey.Key }}
PresharedKey={{ .Peer.PresharedKey.Key }}
DNS = {{ .DsnetConfig.DNS }}

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
