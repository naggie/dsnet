package dsnet

import (
	"net"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

// keepalive for everything
const KeepaliveSeconds = 21;
const ExpiryDays = 28;
const DefaultListenPort = 51820;

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
	ListenPort *int           `validate:"gte=1024,lte=65535"`
	FirewallMark *int
	Peers []PeerConfig
}

type Dsnet struct {
	Name string
	Type DeviceType
	PrivateKey Key
	PublicKey Key
	ListenPort int
	Peers []Peer
}
