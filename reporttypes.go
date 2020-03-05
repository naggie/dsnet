package dsnet

import (
	"encoding/json"
	"io/ioutil"
	"net"
	"time"

	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
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

// TODO pending/unknown

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
	ExternalIP net.IP
	InterfaceName string
	ListenPort int
	// domain to append to hostnames. Relies on separate DNS server for
	// resolution. Informational only.
	Domain     string
	IP      net.IP
	// IP network from which to allocate automatic sequential addresses
	// Network is chosen randomly when not specified
	Network JSONIPNet
	DNS     net.IP
	Peers   []PeerReport
}

func GenerateReport(dev *wgtypes.Device, conf *DsnetConfig) DsnetReport {
	wgPeerIndex := make(map[wgtypes.Key]wgtypes.Peer)
	peerReports := make([]PeerReport, len(conf.Peers))

	for _, peer := range dev.Peers {
		wgPeerIndex[peer.PublicKey] = peer
	}

	for i, peer := range conf.Peers {
		wgPeer, known := wgPeerIndex[peer.PublicKey.Key]

		peerReports[i] = PeerReport{
			Hostname: peer.Hostname,
			Owner: peer.Owner,
			Description: peer.Description,
			IP: peer.IP,
			// TODO Status
			Networks: peer.Networks,
			LastHandshakeTime: wgPeer.LastHandshakeTime,
			ReceiveBytes: wgPeer.ReceiveBytes,
			TransmitBytes: wgPeer.TransmitBytes,
		}
	}

	return DsnetReport{
		ExternalIP: conf.ExternalIP,
		InterfaceName: conf.InterfaceName,
		ListenPort: conf.ListenPort,
		Domain: conf.Domain,
		IP: conf.IP,
		Network: conf.Network,
		DNS: conf.DNS,
		Peers: peerReports,
	}
}

func (report *DsnetReport) MustSave(filename string) {
	_json, _ := json.MarshalIndent(report, "", "    ")
	err := ioutil.WriteFile(filename, _json, 0644)
	check(err)
}

type PeerReport struct {
	// Used to update DNS
	Hostname string
	// username of person running this host/router
	Owner string
	// Description of what the host is and/or does
	Description string
	// Internal VPN IP address. Added to AllowedIPs in server config as a /32
	IP     net.IP
	Status Status
	// TODO ExternalIP support (Endpoint)
	//ExternalIP     net.UDPAddr `validate:"required,udp4_addr"`
	// TODO support routing additional networks (AllowedIPs)
	Networks          []JSONIPNet
	LastHandshakeTime time.Time
	ReceiveBytes      int64
	TransmitBytes     int64
}
