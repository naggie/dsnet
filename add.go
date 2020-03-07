package dsnet

import (
	"fmt"
	"os"
	"text/template"
	"time"
)

func Add() {
	if len(os.Args) <= 2 {
		// TODO non-red
		ExitFail("Hostname argument required: dsnet add <hostname>")
	}

	// TODO maybe accept flags to avoid prompt and allow programmatic use?
	// TODO accept existing pubkey
	conf := MustLoadDsnetConfig()

	hostname := os.Args[2]
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
DNS = {{ .DsnetConfig.DNS }}

[Peer]
PublicKey={{ .DsnetConfig.PrivateKey.PublicKey.Key }}
PresharedKey={{ .Peer.PresharedKey.Key }}
Endpoint={{ .DsnetConfig.ExternalIP }}:{{ .DsnetConfig.ListenPort }}
AllowedIPs={{ .DsnetConfig.Network }}
PersistentKeepalive={{ .Keepalive }}
`

	t := template.Must(template.New("peerConf").Parse(peerConf))
	err := t.Execute(os.Stdout, map[string]interface{}{
		"Peer":        peer,
		"DsnetConfig": conf,
		"Keepalive":   time.Duration(KEEPALIVE).Seconds(),
	})
	check(err)
}
