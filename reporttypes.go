package dsnet

import (
	"net"
	"time"
)

type DsnetReport struct {
	// domain to append to hostnames. Relies on separate DNS server for
	// resolution. Informational only.
	ExternalIP net.IP `validate:"required,cidr"`
	ListenPort int    `validate:"gte=1024,lte=65535"`
	Domain     string `validate:"required,gte=1,lte=255"`
	// IP network from which to allocate automatic sequential addresses
	// Network is chosen randomly when not specified
	Network JSONIPNet `validate:"required"`
	IP      net.IP    `validate:"required,cidr"`
	DNS     net.IP    `validate:"required,cidr"`
	Peers   []PeerReport
}

type PeerReport struct {
	// Used to update DNS
	Hostname string `validate:"required,gte=1,lte=255"`
	// username of person running this host/router
	Owner string `validate:"required,gte=1,lte=255"`
	// Description of what the host is and/or does
	Description string `validate:"required,gte=1,lte=255"`
	// Internal VPN IP address. Added to AllowedIPs in server config as a /32
	IP           net.IP  `validate:"required,ip`
	PublicKey    JSONKey `validate:"required,len=44"`
	PrivateKey   JSONKey `json:"-"` // omitted from config!
	PresharedKey JSONKey `validate:"required,len=44"`
	// whether last heartbeat/rxdata was received (50% margin)
	Online bool
	// if no data for x days, consider revoking access
	Expired bool
	// TODO ExternalIP support (Endpoint)
	//ExternalIP     net.UDPAddr `validate:"required,udp4_addr"`
	// TODO support routing additional networks (AllowedIPs)
	Networks          []JSONIPNet `validate:"dive,cidr"`
	LastHandshakeTime time.Time
	ReceiveBytes      int64
	TransmitBytes     int64
}
