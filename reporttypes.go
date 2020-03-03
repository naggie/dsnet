package dsnet

import (
	"net"
	"time"
)

type Status int

const (
	// Host has not been loaded into wireguard yet
	Pending = iota
	// Host has not transferred anything (not even a keepalive) for 30 seconds
	Offline
	// Host has transferred something in the last 30 seconds, keepalive counts
	Online
	// Host has not connected for 28 days and may be removed
	Expired
)

func (s Status) String() string {
	switch s {
		case Pending:
			return "pending"
		case Offline:
			return "offline"
		case Online:
			return "online"
		case Expired:
			return "expired"
		default:
			return "unknown"
	}
}

// note unmarshal not required
func (s Status) MarshalJSON() ([]byte, error) {
	return []byte("\"" + s.String() + "\""), nil
}

type DsnetReport struct {
	// domain to append to hostnames. Relies on separate DNS server for
	// resolution. Informational only.
	ExternalIP net.IP
	ListenPort int
	Domain     string
	// IP network from which to allocate automatic sequential addresses
	// Network is chosen randomly when not specified
	Network JSONIPNet
	IP      net.IP
	DNS     net.IP
	Peers   []PeerReport
}

type PeerReport struct {
	// Used to update DNS
	Hostname string
	// username of person running this host/router
	Owner string
	// Description of what the host is and/or does
	Description string
	// Internal VPN IP address. Added to AllowedIPs in server config as a /32
	IP net.IP
	// whether last heartbeat/rxdata was received (50% margin)
	Online bool
	// if no data for x days, consider revoking access
	Expired bool
	// TODO ExternalIP support (Endpoint)
	//ExternalIP     net.UDPAddr `validate:"required,udp4_addr"`
	// TODO support routing additional networks (AllowedIPs)
	Networks          []JSONIPNet
	LastHandshakeTime time.Time
	ReceiveBytes      int64
	TransmitBytes     int64
}
