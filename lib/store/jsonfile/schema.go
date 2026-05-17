package jsonfile

import (
	"net"
	"time"

	"github.com/naggie/dsnet/lib"
)

// see https://github.com/WireGuard/wgctrl-go/blob/master/wgtypes/types.go for definitions
type peerConfig struct {
	// Used to update DNS
	Hostname string `validate:"required,gte=1,lte=255"`
	// username of person running this host/router
	Owner string `validate:"required,gte=1,lte=255"`
	// Description of what the host is and/or does
	Description string `validate:"required,gte=1,lte=255"`
	// Internal VPN IP address. Added to AllowedIPs in server config as a /32
	IP    net.IP
	IP6   net.IP
	Added time.Time `validate:"required"`
	// TODO ExternalIP support (Endpoint)
	//ExternalIP     net.UDPAddr `validate:"required,udp4_addr"`
	// TODO support routing additional networks (AllowedIPs)
	Networks     []lib.JSONIPNet `validate:"required"`
	PublicKey    lib.JSONKey     `validate:"required"`
	PrivateKey   lib.JSONKey     `json:"-"` // omitted from config!
	PresharedKey lib.JSONKey     `validate:"required"`
}

type dsnetConfig struct {
	// When generating configs, the ExternalHostname has precendence for the
	// server Endpoint, followed by ExternalIP (IPv4) and ExternalIP6 (IPv6)
	// The IPs are discovered automatically on init. Define an ExternalHostname
	// if you're using dynamic DNS, want to change IPs without updating
	// configs, or want wireguard to be able to choose between IPv4/IPv6. It is
	// only possible to specify one Endpoint per peer entry in wireguard.
	ExternalHostname string
	ExternalIP       net.IP
	ExternalIP6      net.IP
	ListenPort       int `validate:"gte=1,lte=65535"`
	// domain to append to hostnames. Relies on separate DNS server for
	// resolution. Informational only.
	Domain        string `validate:"required,gte=1,lte=255"`
	InterfaceName string `validate:"required,gte=1,lte=255"`
	// IP network from which to allocate automatic sequential addresses
	// Network is chosen randomly when not specified
	Network  lib.JSONIPNet `validate:"required"`
	Network6 lib.JSONIPNet `validate:"required"`
	IP       net.IP
	IP6      net.IP
	DNS      net.IP
	// extra networks available, will be added to AllowedIPs
	Networks []lib.JSONIPNet `validate:"required"`
	// TODO Default subnets to route via VPN
	PrivateKey lib.JSONKey `validate:"required"`
	PostUp     string
	PostDown   string
	Peers      []peerConfig `validate:"dive"`
	// used for server and client
	PersistentKeepalive int `validate:"gte=0,lte=255"`
	MTU                 int `validate:"gte=0,lte=65535"`
}

func (c *dsnetConfig) toServer(fallbackWGBin string) *lib.Server {
	peers := make([]lib.Peer, 0, len(c.Peers))
	for _, p := range c.Peers {
		peers = append(peers, lib.Peer{
			Hostname:            p.Hostname,
			Owner:               p.Owner,
			Description:         p.Description,
			IP:                  p.IP,
			IP6:                 p.IP6,
			Added:               p.Added,
			PublicKey:           p.PublicKey,
			PrivateKey:          p.PrivateKey,
			PresharedKey:        p.PresharedKey,
			Networks:            p.Networks,
			PersistentKeepalive: c.PersistentKeepalive,
		})
	}
	return &lib.Server{
		ExternalHostname:    c.ExternalHostname,
		ExternalIP:          c.ExternalIP,
		ExternalIP6:         c.ExternalIP6,
		ListenPort:          c.ListenPort,
		Domain:              c.Domain,
		InterfaceName:       c.InterfaceName,
		Network:             c.Network,
		Network6:            c.Network6,
		IP:                  c.IP,
		IP6:                 c.IP6,
		DNS:                 c.DNS,
		PrivateKey:          c.PrivateKey,
		PostUp:              c.PostUp,
		PostDown:            c.PostDown,
		FallbackWGBin:       fallbackWGBin,
		Peers:               peers,
		Networks:            c.Networks,
		PersistentKeepalive: c.PersistentKeepalive,
		MTU:                 c.MTU,
	}
}

func fromServer(s *lib.Server) *dsnetConfig {
	peers := make([]peerConfig, 0, len(s.Peers))
	for _, p := range s.Peers {
		peers = append(peers, peerConfig{
			Hostname:     p.Hostname,
			Owner:        p.Owner,
			Description:  p.Description,
			IP:           p.IP,
			IP6:          p.IP6,
			Added:        p.Added,
			Networks:     p.Networks,
			PublicKey:    p.PublicKey,
			PrivateKey:   p.PrivateKey,
			PresharedKey: p.PresharedKey,
		})
	}
	return &dsnetConfig{
		ExternalHostname:    s.ExternalHostname,
		ExternalIP:          s.ExternalIP,
		ExternalIP6:         s.ExternalIP6,
		ListenPort:          s.ListenPort,
		Domain:              s.Domain,
		InterfaceName:       s.InterfaceName,
		Network:             s.Network,
		Network6:            s.Network6,
		IP:                  s.IP,
		IP6:                 s.IP6,
		DNS:                 s.DNS,
		Networks:            s.Networks,
		PrivateKey:          s.PrivateKey,
		PostUp:              s.PostUp,
		PostDown:            s.PostDown,
		Peers:               peers,
		PersistentKeepalive: s.PersistentKeepalive,
		MTU:                 s.MTU,
	}
}
