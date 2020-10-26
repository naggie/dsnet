package dsnet

import (
	"fmt"
	"os"
	"net"
	"text/template"
	"time"
)

const wgQuickPeerConf = `[Interface]
Address={{ .Peer.IP }}/22
Address={{ .Peer.IP6 }}/64
PrivateKey={{ .Peer.PrivateKey.Key }}
{{- if .DsnetConfig.DNS }}
DNS={{ .DsnetConfig.DNS }}
{{ end }}

[Peer]
PublicKey={{ .DsnetConfig.PrivateKey.PublicKey.Key }}
PresharedKey={{ .Peer.PresharedKey.Key }}
Endpoint={{ .DsnetConfig.ExternalIP }}:{{ .DsnetConfig.ListenPort }}
PersistentKeepalive={{ .Keepalive }}
{{ range .AllowedIPs -}}
AllowedIPs={{ . }}
{{ end }}
`

// TODO use random wg0-wg999 to hopefully avoid conflict by default?
const vyattaPeerConf = `configure
set interfaces wireguard wg0 address {{ .Peer.IP }}/{{ .Cidrmask }}
set interfaces wireguard wg0 route-allowed-ips true
set interfaces wireguard wg0 private-key {{ .Peer.PrivateKey.Key }}
set interfaces wireguard wg0 description {{ conf.InterfaceName }}
{{- if .DsnetConfig.DNS }}
#set service dns forwarding name-server {{ .DsnetConfig.DNS }}
{{ end }}

set interfaces wireguard wg0 peer {{ .DsnetConfig.PrivateKey.PublicKey.Key }} endpoint {{ .DsnetConfig.ExternalIP }}:{{ .DsnetConfig.ListenPort }}
set interfaces wireguard wg0 peer {{ .DsnetConfig.PrivateKey.PublicKey.Key }} persistent-keepalive {{ .Keepalive }}
set interfaces wireguard wg0 peer {{ .DsnetConfig.PrivateKey.PublicKey.Key }} preshared-key {{ .Peer.PresharedKey.Key }}
{{ range .AllowedIPs -}}
set interfaces wireguard wg0 peer {{ .DsnetConfig.PrivateKey.PublicKey.Key }} allowed-ips {{ . }}
{{ end }}
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

	peer := PeerConfig{
		Owner:        owner,
		Hostname:     hostname,
		Description:  description,
		Added:        time.Now(),
		PublicKey:    publicKey,
		PrivateKey:   privateKey, // omitted from server config JSON!
		PresharedKey: GenerateJSONKey(),
		Networks:     []JSONIPNet{},
	}

	if len(conf.Network.IPNet.Mask) > 0 {
		peer.IP = conf.MustAllocateIP()
	}

	if len(conf.Network6.IPNet.Mask) > 0 {
		peer.IP6 = conf.MustAllocateIP6()
	}

	if len(conf.IP) == 0 && len(conf.IP6) == 0 {
		ExitFail("No IPv4 or IPv6 network defined in config")
	}

	conf.MustAddPeer(peer)
	PrintPeerCfg(peer, conf)
	conf.MustSave()
	ConfigureDevice(conf)
}

func PrintPeerCfg(peer PeerConfig, conf *DsnetConfig) {
	allowedIPs := make([]JSONIPNet, len(conf.Networks)+2)
	allowedIPs[0] = conf.Network
	allowedIPs[1] = conf.Network6
	allowedIPs = append(allowedIPs, conf.Networks...)

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

	cidrmask, _ := conf.Network.IPNet.Mask.Size()

	t := template.Must(template.New("peerConf").Parse(peerConf))
	err := t.Execute(os.Stdout, map[string]interface{}{
		"Peer":        peer,
		"DsnetConfig": conf,
		"Keepalive":   time.Duration(KEEPALIVE).Seconds(),
		"AllowedIPs":  allowedIPs,
		"Cidrmask":    cidrmask,
		"Address": net.IPNet{
			IP: peer.IP,
			Mask: conf.Network.IPNet.Mask,
		},
		"Address6": net.IPNet{
			IP: peer.IP6,
			Mask: conf.Network6.IPNet.Mask,
		},
	})
	check(err)
}
