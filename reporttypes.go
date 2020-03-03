package dsnet

import (
	"net"
	"time"

	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

type DsnetReport struct {
	Name       string
	PrivateKey wgtypes.Key
	PublicKey  wgtypes.Key
	ListenPort int
	Peers      []Peer
}

type PeerReport struct {
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

	PublicKey    wgtypes.Key
	PresharedKey wgtypes.Key
	// TODO peer endpoint support
	//Endpoint          *net.UDPAddr
	LastHandshakeTime time.Time
	ReceiveBytes      int64
	TransmitBytes     int64
	AllowedIPs        []net.IPNet
}
