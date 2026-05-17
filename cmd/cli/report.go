package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"time"

	"github.com/naggie/dsnet/lib"
	"github.com/spf13/viper"
	"github.com/vishvananda/netlink"
	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

type DsnetReport struct {
	ExternalIP       net.IP
	ExternalIP6      net.IP
	ExternalHostname string
	InterfaceName    string
	ListenPort       int
	// domain to append to hostnames. Relies on separate DNS server for
	// resolution. Informational only.
	Domain string
	IP     net.IP
	IP6    net.IP
	// IP network from which to allocate automatic sequential addresses
	// Network is chosen randomly when not specified
	Network         lib.JSONIPNet
	Network6        lib.JSONIPNet
	DNS             net.IP
	PeersOnline     int
	PeersTotal      int
	Peers           []PeerReport
	ReceiveBytes    uint64
	TransmitBytes   uint64
	ReceiveBytesSI  string
	TransmitBytesSI string
	// when the report was made
	Timestamp time.Time
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
	IP  net.IP
	IP6 net.IP
	// Last known external IP
	ExternalIP net.IP
	// TODO ExternalIP support (Endpoint)
	//ExternalIP     net.UDPAddr `validate:"required,udp4_addr"`
	// TODO support routing additional networks (AllowedIPs)
	Networks          []lib.JSONIPNet
	LastHandshakeTime time.Time
	ReceiveBytes      uint64
	TransmitBytes     uint64
	ReceiveBytesSI    string
	TransmitBytesSI   string
}

func GenerateReport() error {
	backend, err := OpenStore()
	if err != nil {
		return fmt.Errorf("%w - failed to open storage backend", err)
	}
	defer backend.Close()

	state, _, err := backend.Load(context.Background())
	if err != nil {
		return fmt.Errorf("%w - failed to load state", err)
	}
	network, err := DefaultNetwork(state)
	if err != nil {
		return err
	}
	server := network.Server

	wg, err := wgctrl.New()
	if err != nil {
		return fmt.Errorf("%w - failure to create new client", err)
	}
	defer wg.Close()

	dev, err := wg.Device(server.InterfaceName)
	if err != nil {
		return fmt.Errorf("%w - Could not retrieve device '%s'", err, server.InterfaceName)
	}

	report, err := GetReport(dev, server)
	if err != nil {
		return err
	}
	return report.Print()
}

func GetReport(dev *wgtypes.Device, server *lib.Server) (DsnetReport, error) {
	peerTimeout := viper.GetDuration("peer_timeout")
	peerExpiry := viper.GetDuration("peer_expiry")
	wgPeerIndex := make(map[wgtypes.Key]wgtypes.Peer)
	peerReports := make([]PeerReport, 0)
	peersOnline := 0

	linkDev, err := netlink.LinkByName(server.InterfaceName)
	if err != nil {
		return DsnetReport{}, fmt.Errorf("%w - error getting link", err)
	}

	stats := linkDev.Attrs().Statistics

	for _, peer := range dev.Peers {
		wgPeerIndex[peer.PublicKey] = peer
	}

	for _, peer := range server.Peers {
		wgPeer, known := wgPeerIndex[peer.PublicKey.Key]

		if !known {
			// dangling peer, sync will remove. Dangling peers aren't such a
			// problem now that add/remove performs a sync too.
			continue
		}

		online := time.Since(wgPeer.LastHandshakeTime) < peerTimeout
		dormant := !wgPeer.LastHandshakeTime.IsZero() && time.Since(wgPeer.LastHandshakeTime) > peerExpiry

		if online {
			peersOnline++
		}

		externalIP := net.IP{}
		if wgPeer.Endpoint != nil {
			externalIP = wgPeer.Endpoint.IP
		}

		uReceiveBytes := uint64(wgPeer.ReceiveBytes)
		uTransmitBytes := uint64(wgPeer.TransmitBytes)

		peerReports = append(peerReports, PeerReport{
			Hostname:          peer.Hostname,
			Online:            online,
			Dormant:           dormant,
			Owner:             peer.Owner,
			Description:       peer.Description,
			Added:             peer.Added,
			IP:                peer.IP,
			IP6:               peer.IP6,
			ExternalIP:        externalIP,
			Networks:          peer.Networks,
			LastHandshakeTime: wgPeer.LastHandshakeTime,
			ReceiveBytes:      uReceiveBytes,
			TransmitBytes:     uTransmitBytes,
			ReceiveBytesSI:    BytesToSI(uReceiveBytes),
			TransmitBytesSI:   BytesToSI(uTransmitBytes),
		})
	}

	return DsnetReport{
		ExternalIP:       server.ExternalIP,
		ExternalIP6:      server.ExternalIP6,
		ExternalHostname: server.ExternalHostname,
		InterfaceName:    server.InterfaceName,
		ListenPort:       server.ListenPort,
		Domain:           server.Domain,
		IP:               server.IP,
		IP6:              server.IP6,
		Network:          server.Network,
		Network6:         server.Network6,
		DNS:              server.DNS,
		Peers:            peerReports,
		PeersOnline:      peersOnline,
		PeersTotal:       len(peerReports),
		ReceiveBytes:     stats.RxBytes,
		TransmitBytes:    stats.TxBytes,
		ReceiveBytesSI:   BytesToSI(stats.RxBytes),
		TransmitBytesSI:  BytesToSI(stats.TxBytes),
		Timestamp:        time.Now(),
	}, nil
}

func (report *DsnetReport) Print() error {
	_json, err := json.MarshalIndent(report, "", "    ")
	if err != nil {
		return fmt.Errorf("failed to marshal report: %w", err)
	}
	_json = append(_json, '\n')
	fmt.Print(string(_json))
	return nil
}
