package dsnet

import (
	"net"
	"time"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

// see https://github.com/WireGuard/wgctrl-go/blob/master/wgtypes/types.go for definitions
type PeerConfig struct {
	// username of person running this host/router
	Owner string              `validate:"required,gte=1,lte=255"`
	// Used to update DNS
	Hostname string           `validate:"required,gte=1,lte=255"`
	// Description of what the host is and/or does
	Description string        `validate:"required,gte=1,lte=255"`

	PublicKey wgtypes.Key     `validate:"required,len=44"`
	PresharedKey wgtypes.Key  `validate:"required,len=44"`
	Endpoint *net.UDPAddr     `validate:"required,udp4_addr"`
	AllowedIPs []net.IPNet    `validate:"dive,required,cidr"`
}

type Peer struct {
	// username of person running this host/router
	Owner string
	// Used to update DNS
	Hostname string
	// Description of what the host is and/or does
	Description string
	// whether last heartbeat/rxdata was received (50% margin)
	Online bool
	// if no data for x days, consider revoking access
	Expired bool

	PublicKey wgtypes.Key
	PresharedKey wgtypes.Key
	Endpoint *net.UDPAddr
	LastHandshakeTime time.Time
	ReceiveBytes int64
	TransmitBytes int64
	AllowedIPs []net.IPNet
}

type DsnetConfig struct {
	PrivateKey *wgtypes.Key   `validate:"required,len=44"`
	PresharedKey wgtypes.Key  `validate:"required,len=44"`
	ListenPort *int           `validate:"gte=1024,lte=65535"`
	Peers []PeerConfig
	// IP network from which to allocate automatic sequential addresses
	// Network is chosen randomly when not specified
	Network net.IPNet         `validate:"required"`
	// domain to append to hostnames. Relies on separate DNS server for
	// resolution. Informational only.
	Domain string             `validate:"required,gte=1,lte=255"`
}

type Dsnet struct {
	Name string
	PrivateKey wgtypes.Key
	PublicKey wgtypes.Key
	ListenPort int
	Peers []Peer
}
