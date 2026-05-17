package render

import (
	"bytes"
	"errors"
	"fmt"
	"text/template"

	"github.com/naggie/dsnet/lib"
)

func getPeerConfTplString(peerType lib.PeerType) (string, error) {
	switch peerType {
	case lib.WGQuick:
		return wgQuickPeerConf, nil
	case lib.Vyatta:
		return vyattaPeerConf, nil
	case lib.NixOS:
		return nixosPeerConf, nil
	case lib.RouterOS:
		return routerosPeerConf, nil
	default:
		return "", fmt.Errorf("unrecognized peer type")
	}
}

// GetWGPeerTemplate returns a template string to be used when
// configuring a peer
func GetWGPeerTemplate(peer lib.Peer, peerType lib.PeerType, server lib.Server) (*bytes.Buffer, error) {
	peerConf, err := getPeerConfTplString(peerType)
	if err != nil {
		return nil, fmt.Errorf("failed to get wg template: %w", err)
	}

	// See DsnetConfig type for explanation
	var endpoint string

	if server.ExternalHostname != "" {
		endpoint = server.ExternalHostname
	} else if len(server.ExternalIP) > 0 {
		endpoint = server.ExternalIP.String()
	} else if len(server.ExternalIP6) > 0 {
		endpoint = "[" + server.ExternalIP6.String() + "]"
	} else {
		return nil, errors.New("server config requires at least one of ExternalIP, ExternalIP6 or ExternalHostname")
	}

	t := template.Must(template.New("peerConf").Parse(peerConf))
	cidrSize, _ := server.Network.IPNet.Mask.Size()
	cidrSize6, _ := server.Network6.IPNet.Mask.Size()

	var templateBuff bytes.Buffer
	err = t.Execute(&templateBuff, map[string]any{
		"Peer":      peer,
		"Server":    server,
		"CidrSize":  cidrSize,
		"CidrSize6": cidrSize6,
		// vyatta requires an interface in range/format wg0-wg999
		// deterministically choosing one in this range will probably allow use
		// of the config without a colliding interface name
		"Wgif":     peer.GetIfName(),
		"Endpoint": endpoint,
	})
	if err != nil {
		return nil, err
	}
	return &templateBuff, nil
}

func AsciiPeerConfig(peer lib.Peer, peerType string, server lib.Server) (*bytes.Buffer, error) {
	switch peerType {
	case "wg-quick":
		return GetWGPeerTemplate(peer, lib.WGQuick, server)
	case "vyatta":
		return GetWGPeerTemplate(peer, lib.Vyatta, server)
	case "nixos":
		return GetWGPeerTemplate(peer, lib.NixOS, server)
	case "routeros":
		return GetWGPeerTemplate(peer, lib.RouterOS, server)
	default:
		return nil, errors.New("unrecognised OUTPUT type")
	}
}
