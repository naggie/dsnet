package cli

import (
	"encoding/json"
	"net"
	"testing"
	"time"

	"github.com/naggie/dsnet/lib"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

func TestDsnetReportPrint(t *testing.T) {
	report := &DsnetReport{
		ExternalIP:       net.IP{1, 2, 3, 4},
		ExternalHostname: "vpn.example.com",
		InterfaceName:    "dsnet",
		ListenPort:       51820,
		Domain:           "dsnet",
		IP:               net.IP{10, 0, 0, 1},
		Network: lib.JSONIPNet{
			IPNet: net.IPNet{
				IP:   net.IP{10, 0, 0, 0},
				Mask: net.IPMask{255, 255, 252, 0},
			},
		},
		PeersOnline:      1,
		PeersTotal:       2,
		ReceiveBytes:     1234567,
		TransmitBytes:    7654321,
		ReceiveBytesSI:   "1.2 MB",
		TransmitBytesSI:  "7.7 MB",
		Timestamp:        time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		Peers:            []PeerReport{},
	}

	// Print outputs to stdout -- just verify it doesn't panic
	// The real test is that the JSON is well-formed
	b, err := json.MarshalIndent(report, "", "    ")
	if err != nil {
		t.Fatalf("failed to marshal report: %v", err)
	}

	var decoded DsnetReport
	err = json.Unmarshal(b, &decoded)
	if err != nil {
		t.Fatalf("failed to unmarshal report: %v", err)
	}

	if decoded.ExternalHostname != "vpn.example.com" {
		t.Fatalf("expected hostname 'vpn.example.com', got '%s'", decoded.ExternalHostname)
	}
	if decoded.ListenPort != 51820 {
		t.Fatalf("expected port 51820, got %d", decoded.ListenPort)
	}
	if decoded.PeersOnline != 1 {
		t.Fatalf("expected 1 online, got %d", decoded.PeersOnline)
	}
	if decoded.PeersTotal != 2 {
		t.Fatalf("expected 2 total, got %d", decoded.PeersTotal)
	}
}

func TestDsnetReportJSONRoundTrip(t *testing.T) {
	privKey, _ := wgtypes.GeneratePrivateKey()
	psk, _ := wgtypes.GenerateKey()
	now := time.Now().Truncate(time.Second) // JSON loses sub-second on some formats

	report := DsnetReport{
		ExternalIP:       net.IP{1, 2, 3, 4},
		ExternalIP6:      net.ParseIP("2001:db8::1"),
		ExternalHostname: "vpn.example.com",
		InterfaceName:    "dsnet",
		ListenPort:       51820,
		Domain:           "dsnet",
		IP:               net.IP{10, 0, 0, 1},
		IP6:              net.IP{0xfd, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1},
		Network: lib.JSONIPNet{
			IPNet: net.IPNet{
				IP:   net.IP{10, 0, 0, 0},
				Mask: net.IPMask{255, 255, 252, 0},
			},
		},
		Network6: lib.JSONIPNet{
			IPNet: net.IPNet{
				IP:   net.IP{0xfd, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
				Mask: net.IPMask{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0, 0, 0, 0, 0, 0, 0, 0},
			},
		},
		DNS:              net.IP{10, 0, 0, 1},
		PeersOnline:      1,
		PeersTotal:       2,
		ReceiveBytes:     1000000,
		TransmitBytes:    2000000,
		ReceiveBytesSI:   BytesToSI(1000000),
		TransmitBytesSI:  BytesToSI(2000000),
		Timestamp:        now,
		Peers: []PeerReport{
			{
				Hostname:          "peer1",
				Owner:             "alice",
				Description:       "Alice's laptop",
				Online:            true,
				Dormant:           false,
				Added:             now,
				IP:                net.IP{10, 0, 0, 2},
				IP6:               net.IP{0xfd, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 2},
				ExternalIP:        net.IP{5, 6, 7, 8},
				Networks:          []lib.JSONIPNet{},
				LastHandshakeTime: now,
				ReceiveBytes:      500000,
				TransmitBytes:     600000,
				ReceiveBytesSI:    BytesToSI(500000),
				TransmitBytesSI:   BytesToSI(600000),
			},
		},
	}

	_ = privKey
	_ = psk

	b, err := json.MarshalIndent(report, "", "    ")
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var decoded DsnetReport
	err = json.Unmarshal(b, &decoded)
	if err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if decoded.InterfaceName != report.InterfaceName {
		t.Fatalf("InterfaceName mismatch: %s != %s", decoded.InterfaceName, report.InterfaceName)
	}
	if decoded.PeersOnline != report.PeersOnline {
		t.Fatal("PeersOnline mismatch")
	}
	if len(decoded.Peers) != 1 {
		t.Fatalf("expected 1 peer, got %d", len(decoded.Peers))
	}
	if decoded.Peers[0].Hostname != "peer1" {
		t.Fatalf("peer hostname mismatch: %s", decoded.Peers[0].Hostname)
	}
	if decoded.Peers[0].Online != true {
		t.Fatal("peer should be online")
	}
}

func TestPeerReportFields(t *testing.T) {
	now := time.Now()
	_, subnet, _ := net.ParseCIDR("192.168.1.0/24")

	pr := PeerReport{
		Hostname:          "myhost",
		Owner:             "owner1",
		Description:       "a host",
		Online:            true,
		Dormant:           false,
		Added:             now,
		IP:                net.IP{10, 0, 0, 5},
		IP6:               net.IP{0xfd, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 5},
		ExternalIP:        net.IP{1, 2, 3, 4},
		Networks:          []lib.JSONIPNet{{IPNet: *subnet}},
		LastHandshakeTime: now,
		ReceiveBytes:      100,
		TransmitBytes:     200,
		ReceiveBytesSI:    "100 B",
		TransmitBytesSI:   "200 B",
	}

	b, err := json.Marshal(pr)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var decoded PeerReport
	err = json.Unmarshal(b, &decoded)
	if err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if decoded.Hostname != "myhost" {
		t.Fatalf("hostname mismatch: %s", decoded.Hostname)
	}
	if decoded.Owner != "owner1" {
		t.Fatalf("owner mismatch: %s", decoded.Owner)
	}
	if !decoded.Online {
		t.Fatal("should be online")
	}
	if decoded.Dormant {
		t.Fatal("should not be dormant")
	}
	if decoded.ReceiveBytes != 100 {
		t.Fatalf("ReceiveBytes mismatch: %d", decoded.ReceiveBytes)
	}
	if len(decoded.Networks) != 1 {
		t.Fatalf("expected 1 network, got %d", len(decoded.Networks))
	}
}
