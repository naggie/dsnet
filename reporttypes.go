package dsnet

import (
	"encoding/json"
	"io/ioutil"
	"net"
	"os"
	"time"

	"github.com/go-playground/validator/v10"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

type Status int

const (
	StatusUnknown = iota
	// Host has not been loaded into wireguard yet
	StatusSyncRequired
	// No handshake in 3 minutes
	StatusOffline
	// Handshake in 3 minutes
	StatusOnline
	// Host has not connected for 28 days and may be removed
	StatusExpired
)

// TODO pending/unknown

func (s Status) String() string {
	switch s {
	case StatusSyncRequired:
		return "syncrequired"
	case StatusOffline:
		return "offline"
	case StatusOnline:
		return "online"
	case StatusExpired:
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
	ExternalIP    net.IP
	InterfaceName string
	ListenPort    int
	// domain to append to hostnames. Relies on separate DNS server for
	// resolution. Informational only.
	Domain string
	IP     net.IP
	// IP network from which to allocate automatic sequential addresses
	// Network is chosen randomly when not specified
	Network     JSONIPNet
	DNS         net.IP
	Peers       []PeerReport
	PeersOnline int
	PeersTotal  int
}

func GenerateReport(dev *wgtypes.Device, conf *DsnetConfig, oldReport *DsnetReport) DsnetReport {
	wgPeerIndex := make(map[wgtypes.Key]wgtypes.Peer)
	peerReports := make([]PeerReport, len(conf.Peers))
	oldPeerReportIndex := make(map[string]PeerReport)
	peersOnline := 0

	for _, peer := range dev.Peers {
		wgPeerIndex[peer.PublicKey] = peer
	}

	if oldReport != nil {
		for _, report := range oldReport.Peers {
			oldPeerReportIndex[report.Hostname] = report
		}
	}

	for i, peer := range conf.Peers {
		wgPeer, known := wgPeerIndex[peer.PublicKey.Key]

		status := Status(StatusUnknown)

		if !known {
			status = StatusSyncRequired
		} else if time.Since(wgPeer.LastHandshakeTime) < TIMEOUT {
			status = StatusOnline
			peersOnline += 1
			// TODO same test but with rx byte data from last report (otherwise
			// peer can fake online status by disabling handshake)
		} else if !wgPeer.LastHandshakeTime.IsZero() && time.Since(wgPeer.LastHandshakeTime) > EXPIRY {
			status = StatusExpired
		} else {
			status = StatusOffline
		}

		peerReports[i] = PeerReport{
			Hostname:          peer.Hostname,
			Owner:             peer.Owner,
			Description:       peer.Description,
			IP:                peer.IP,
			Status:            status,
			Networks:          peer.Networks,
			LastHandshakeTime: wgPeer.LastHandshakeTime,
			ReceiveBytes:      wgPeer.ReceiveBytes,
			TransmitBytes:     wgPeer.TransmitBytes,
			ReceiveBytesSI:    BytesToSI(wgPeer.ReceiveBytes),
			TransmitBytesSI:   BytesToSI(wgPeer.TransmitBytes),
		}
	}

	return DsnetReport{
		ExternalIP:    conf.ExternalIP,
		InterfaceName: conf.InterfaceName,
		ListenPort:    conf.ListenPort,
		Domain:        conf.Domain,
		IP:            conf.IP,
		Network:       conf.Network,
		DNS:           conf.DNS,
		Peers:         peerReports,
		PeersOnline:   peersOnline,
		PeersTotal:    len(peerReports),
	}
}

func (report *DsnetReport) MustSave(filename string) {
	_json, _ := json.MarshalIndent(report, "", "    ")
	err := ioutil.WriteFile(filename, _json, 0644)
	check(err)
}

func MustLoadDsnetReport() *DsnetReport {
	raw, err := ioutil.ReadFile(CONFIG_FILE)

	if os.IsNotExist(err) {
		return nil
	} else if os.IsPermission(err) {
		ExitFail("%s cannot be accessed. Check read permissions.", CONFIG_FILE)
	} else {
		check(err)
	}

	report := DsnetReport{}
	err = json.Unmarshal(raw, &report)
	check(err)

	err = validator.New().Struct(report)
	check(err)

	return &report
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
	ReceiveBytesSI    string
	TransmitBytesSI   string
}
