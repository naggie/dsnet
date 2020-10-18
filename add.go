package dsnet

import (
	"fmt"
	"os"
	"strings"
	"text/template"
	"time"
)

const wgQuickPeerConf = `[Interface]
Address = {{ .Peer.IP }}
PrivateKey={{ .Peer.PrivateKey.Key }}
{{- if .DsnetConfig.DNS }}
DNS = {{ .DsnetConfig.DNS }}
{{ end }}

[Peer]
PublicKey={{ .DsnetConfig.PrivateKey.PublicKey.Key }}
PresharedKey={{ .Peer.PresharedKey.Key }}
Endpoint={{ .DsnetConfig.ExternalIP }}:{{ .DsnetConfig.ListenPort }}
AllowedIPs={{ .AllowedIPs }}
PersistentKeepalive={{ .Keepalive }}
`

const vyattaPeerConf = `[Interface]
configure

set interfaces wireguard dsnet address {{ .Peer.IP }}
set interfaces wireguard dsnet route-allowed-ips true

set interfaces wireguard dsnet peer {{ .DsnetConfig.PrivateKey.PublicKey.Key }} endpoint {{ .DsnetConfig.ExternalIP }}:{{ .DsnetConfig.ListenPort }}
set interfaces wireguard dsnet peer allowed-ips {{.AllowedIPs}}
set interfaces wireguard dsnet peer persistent-keepalive {{.AllowedIPs}}

{{- if .DsnetConfig.DNS }}
#set service dns forwarding name-server {{ .DsnetConfig.DNS }}
{{ end }}

set interfaces wireguard dsnet private-key {{ .Peer.PrivateKey.Key }}
set interfaces wireguard dsnet preshared-key {{ .Peer.PresharedKey.Key }}

commit; save
`

func Add() {
	if len(os.Args) != 3 {
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
		Added:        time.Now(),
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
	allowedIPsStr := make([]string, len(conf.Networks)+1)
	allowedIPsStr[0] = conf.Network.String()

	for i, net := range conf.Networks {
		allowedIPsStr[i+1] = net.String()
	}

	var peerConf string

	switch os.Getenv("DSNET_OUTPUT") {
	// https://manpages.debian.org/unstable/wireguard-tools/wg-quick.8.en.html
	case "", "wg-quick":
		peerConf = wgQuickPeerConf
	// https://github.com/WireGuard/wireguard-vyatta-ubnt/
	case "vyatta":
		peerConf = vyattaPeerConf
	default:
		ExitFail("Unrecognised DSNET_OUTPUT type")
	}

	t := template.Must(template.New("peerConf").Parse(peerConf))
	err := t.Execute(os.Stdout, map[string]interface{}{
		"Peer":        peer,
		"DsnetConfig": conf,
		"Keepalive":   time.Duration(KEEPALIVE).Seconds(),
		"AllowedIPs":  strings.Join(allowedIPsStr, ","),
	})
	check(err)
}
