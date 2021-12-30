package lib

import (
	"bytes"
	"errors"
	"fmt"
	"text/template"
	"time"
)

func getPeerConfTplString(peerType PeerType) (string, error) {
	switch peerType {
	case WGQuick:
		return wgQuickPeerConf, nil
	case Vyatta:
		return vyattaPeerConf, nil
	case NixOS:
		return nixosPeerConf, nil
	default:
		return "", fmt.Errorf("unrecognized peer type")
	}
}

func (p *Peer) getIfName() string {
	// derive deterministic interface name
	wgifSeed := 0
	for _, b := range p.IP {
		wgifSeed += int(b)
	}

	for _, b := range p.IP6 {
		wgifSeed += int(b)
	}
	return fmt.Sprintf("wg%d", wgifSeed%999)
}

// GetWGPeerTemplate returns a template string to be used when
// configuring a peer
func getWGPeerTemplate(peer Peer, peerType PeerType, server Server) (*bytes.Buffer, error) {
	peerConf, err := getPeerConfTplString(peerType)
	if err != nil {
		return nil, fmt.Errorf("failed to get wg template: %s", err)
	}

	// See DsnetConfig type for explanation
	var endpoint string

	if server.ExternalHostname != "" {
		endpoint = server.ExternalHostname
	} else if len(server.ExternalIP) > 0 {
		endpoint = server.ExternalIP.String()
	} else if len(server.ExternalIP6) > 0 {
		endpoint = server.ExternalIP6.String()
	} else {
		return nil, errors.New("server config requires at least one of ExternalIP, ExternalIP6 or ExternalHostname")
	}

	t := template.Must(template.New("peerConf").Parse(peerConf))
	cidrSize, _ := server.Network.IPNet.Mask.Size()
	cidrSize6, _ := server.Network6.IPNet.Mask.Size()

	var templateBuff bytes.Buffer
	err = t.Execute(&templateBuff, map[string]interface{}{
		"Peer":      peer,
		"Server":    server,
		"Keepalive": time.Duration(peer.KeepAlive).Seconds(),
		"CidrSize":  cidrSize,
		"CidrSize6": cidrSize6,
		// vyatta requires an interface in range/format wg0-wg999
		// deterministically choosing one in this range will probably allow use
		// of the config without a colliding interface name
		"Wgif":     peer.getIfName(),
		"Endpoint": endpoint,
	})
	if err != nil {
		return nil, err
	}
	return &templateBuff, nil
}

func AsciiPeerConfig(peer Peer, peerType string, server Server) (*bytes.Buffer, error) {
	switch peerType {
	case "wg-quick":
		return getWGPeerTemplate(peer, WGQuick, server)
	case "vyatta":
		return getWGPeerTemplate(peer, Vyatta, server)
	case "nixos":
		return getWGPeerTemplate(peer, NixOS, server)
	default:
		return nil, errors.New("unrecognised OUTPUT type")
	}
}
