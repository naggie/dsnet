package dsnet

import (
	"fmt"
	"os"
	"text/template"
)

func Add() {
	// TODO maybe accept flags to avoid prompt and allow programmatic use
	// TODO accept existing pubkey
	conf := MustLoadDsnetConfig()

	hostname := MustPromptString("Hostname (unique)", true)
	owner := MustPromptString("owner", true)
	description := MustPromptString("Description", true)
	//publicKey := MustPromptString("PublicKey (optional)", false)
	ConfirmOrAbort("\nDo you want to add the above configuration?")

	// newline (not on stdout) to separate config
	fmt.Fprintln(os.Stderr)

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
		Networks:     []JSONIPNet{},
	}

	conf.MustAddPeer(peer)
	PrintPeerCfg(peer, conf)
	conf.MustSave()
	ConfigureDevice(conf)
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
