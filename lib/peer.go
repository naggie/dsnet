package lib

import (
	"errors"
	"fmt"
	"net"
	"time"
)

// PeerType is what configuration to use when generating
// peer config files
type PeerType int

const (
	// WGQuick is used by wg-quick to set up a peer
	// https://manpages.debian.org/unstable/wireguard-tools/wg-quick.8.en.html
	WGQuick PeerType = iota
	// Vyatta is used by Ubiquiti routers
	// https://github.com/WireGuard/wireguard-vyatta-ubnt/
	Vyatta
	// NixOS is a declartive linux distro
	// https://nixos.wiki/wiki/Wireguard
	NixOS
)

type Peer struct {
	Hostname     string
	Owner        string
	Description  string
	IP           net.IP
	IP6          net.IP
	Added        time.Time
	PublicKey    JSONKey
	PrivateKey   JSONKey
	PresharedKey JSONKey
	Networks     []JSONIPNet
	KeepAlive    time.Duration
}

func NewPeer(server *Server, owner string, hostname string, description string) (Peer, error) {
	if owner == "" {
		return Peer{}, errors.New("missing owner")
	}
	if hostname == "" {
		return Peer{}, errors.New("missing hostname")
	}

	privateKey, err := GenerateJSONPrivateKey()
	if err != nil {
		return Peer{}, fmt.Errorf("failed to generate private key: %s", err)
	}
	publicKey := privateKey.PublicKey()

	presharedKey, err := GenerateJSONKey()
	if err != nil {
		return Peer{}, fmt.Errorf("failed to generate private key: %s")
	}

	newPeer := Peer{
		Owner:        owner,
		Hostname:     hostname,
		Description:  description,
		Added:        time.Now(),
		PublicKey:    publicKey,
		PrivateKey:   privateKey,
		PresharedKey: presharedKey,
		Networks:     []JSONIPNet{},
	}

	if len(server.Network.IPNet.Mask) > 0 {
		newIP, err := server.AllocateIP()
		if err != nil {
			return Peer{}, fmt.Errorf("failed to allocate ipv4 address: %s", err)
		}
		newPeer.IP = newIP
	}

	if len(server.Network6.IPNet.Mask) > 0 {
		newIPV6, err := server.AllocateIP6()
		if err != nil {
			return Peer{}, fmt.Errorf("failed to allocate ipv6 address: %s", err)
		}
		newPeer.IP = newIPV6
	}

	if len(server.IP) == 0 && len(server.IP6) == 0 {
		return Peer{}, fmt.Errorf("no IPv4 or IPv6 network defined in config")
	}
	return newPeer, nil
}
