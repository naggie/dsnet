package dsnet

import (
	"encoding/json"
	"io/ioutil"
	"net"
	"os"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/vishvananda/netlink"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

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
	Network         JSONIPNet
	DNS             net.IP
	PeersOnline     int
	PeersTotal      int
	Peers           []PeerReport
	ReceiveBytes    uint64
	TransmitBytes   uint64
	ReceiveBytesSI  string
	TransmitBytesSI string
}

func GenerateReport(dev *wgtypes.Device, conf *DsnetConfig, oldReport *DsnetReport) DsnetReport {
	wgPeerIndex := make(map[wgtypes.Key]wgtypes.Peer)
	peerReports := make([]PeerReport, len(conf.Peers))
	oldPeerReportIndex := make(map[string]PeerReport)
	peersOnline := 0

	linkDev, err := netlink.LinkByName(conf.InterfaceName)
	check(err)

	stats := linkDev.Attrs().Statistics

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

		if !known {
			// dangling peer, sync will remove. Dangling peers aren't such a
			// problem now that add/remove performs a sync too.
			continue
		}

		online := time.Since(wgPeer.LastHandshakeTime) < TIMEOUT
		dormant := !wgPeer.LastHandshakeTime.IsZero() && time.Since(wgPeer.LastHandshakeTime) > EXPIRY

		if online {
			peersOnline++
		}

		externalIP := net.IP{}
		if wgPeer.Endpoint != nil {
			externalIP = wgPeer.Endpoint.IP
		}

		uReceiveBytes := uint64(wgPeer.ReceiveBytes)
		uTransmitBytes := uint64(wgPeer.TransmitBytes)

		peerReports[i] = PeerReport{
			Hostname:          peer.Hostname,
			Online:            online,
			Dormant:           dormant,
			Owner:             peer.Owner,
			Description:       peer.Description,
			Added:             peer.Added,
			IP:                peer.IP,
			ExternalIP:        externalIP,
			Networks:          peer.Networks,
			LastHandshakeTime: wgPeer.LastHandshakeTime,
			ReceiveBytes:      uReceiveBytes,
			TransmitBytes:     uTransmitBytes,
			ReceiveBytesSI:    BytesToSI(uReceiveBytes),
			TransmitBytesSI:   BytesToSI(uTransmitBytes),
		}
	}

	return DsnetReport{
		ExternalIP:      conf.ExternalIP,
		InterfaceName:   conf.InterfaceName,
		ListenPort:      conf.ListenPort,
		Domain:          conf.Domain,
		IP:              conf.IP,
		Network:         conf.Network,
		DNS:             conf.DNS,
		Peers:           peerReports,
		PeersOnline:     peersOnline,
		PeersTotal:      len(peerReports),
		ReceiveBytes:    stats.RxBytes,
		TransmitBytes:   stats.TxBytes,
		ReceiveBytesSI:  BytesToSI(stats.RxBytes),
		TransmitBytesSI: BytesToSI(stats.TxBytes),
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
	// Has a handshake occurred in the last 3 mins?
	Online bool
	// No handshake for 28 days
	Dormant bool
	// date peer was added to dsnet config
	Added time.Time
	// Internal VPN IP address. Added to AllowedIPs in server config as a /32
	IP net.IP
	// Last known external IP
	ExternalIP net.IP
	// TODO ExternalIP support (Endpoint)
	//ExternalIP     net.UDPAddr `validate:"required,udp4_addr"`
	// TODO support routing additional networks (AllowedIPs)
	Networks          []JSONIPNet
	LastHandshakeTime time.Time
	ReceiveBytes      uint64
	TransmitBytes     uint64
	ReceiveBytesSI    string
	TransmitBytesSI   string
}
