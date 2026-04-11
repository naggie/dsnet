package cli

import (
	"net"
	"testing"
	"time"

	"github.com/naggie/dsnet/lib"
	"github.com/spf13/viper"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

func TestGetServerFieldMapping(t *testing.T) {
	viper.Set("fallback_wg_bin", "/usr/bin/wg")
	t.Cleanup(func() { viper.Set("fallback_wg_bin", "") })

	privKey, _ := wgtypes.GeneratePrivateKey()
	peerKey, _ := wgtypes.GeneratePrivateKey()
	psk, _ := wgtypes.GenerateKey()
	_, extraNet, _ := net.ParseCIDR("192.168.1.0/24")

	config := &DsnetConfig{
		ExternalHostname: "vpn.example.com",
		ExternalIP:       net.IP{1, 2, 3, 4},
		ExternalIP6:      net.ParseIP("2001:db8::1"),
		ListenPort:       51820,
		Domain:           "dsnet",
		InterfaceName:    "wg0",
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
		IP:       net.IP{10, 0, 0, 1},
		IP6:      net.IP{0xfd, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1},
		DNS:      net.IP{10, 0, 0, 1},
		PrivateKey: lib.JSONKey{Key: privKey},
		PostUp:   "iptables -A",
		PostDown: "iptables -D",
		Networks: []lib.JSONIPNet{{IPNet: *extraNet}},
		PersistentKeepalive: 30,
		Peers: []PeerConfig{
			{
				Hostname:     "peer1",
				Owner:        "alice",
				Description:  "test",
				IP:           net.IP{10, 0, 0, 2},
				IP6:          net.IP{0xfd, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 2},
				Added:        time.Now(),
				PublicKey:    lib.JSONKey{Key: peerKey.PublicKey()},
				PrivateKey:   lib.JSONKey{Key: peerKey},
				PresharedKey: lib.JSONKey{Key: psk},
				Networks:     []lib.JSONIPNet{},
			},
		},
	}

	server := GetServer(config)

	if server.ExternalHostname != "vpn.example.com" {
		t.Fatalf("ExternalHostname mismatch: %s", server.ExternalHostname)
	}
	if !server.ExternalIP.Equal(net.IP{1, 2, 3, 4}) {
		t.Fatalf("ExternalIP mismatch: %s", server.ExternalIP)
	}
	if server.ListenPort != 51820 {
		t.Fatalf("ListenPort mismatch: %d", server.ListenPort)
	}
	if server.Domain != "dsnet" {
		t.Fatalf("Domain mismatch: %s", server.Domain)
	}
	if server.InterfaceName != "wg0" {
		t.Fatalf("InterfaceName mismatch: %s", server.InterfaceName)
	}
	if server.Network.IPNet.String() != config.Network.IPNet.String() {
		t.Fatal("Network mismatch")
	}
	if server.Network6.IPNet.String() != config.Network6.IPNet.String() {
		t.Fatal("Network6 mismatch")
	}
	if !server.IP.Equal(config.IP) {
		t.Fatal("IP mismatch")
	}
	if !server.IP6.Equal(config.IP6) {
		t.Fatal("IP6 mismatch")
	}
	if !server.DNS.Equal(config.DNS) {
		t.Fatal("DNS mismatch")
	}
	if server.PrivateKey.Key != privKey {
		t.Fatal("PrivateKey mismatch")
	}
	if server.PostUp != "iptables -A" {
		t.Fatalf("PostUp mismatch: %s", server.PostUp)
	}
	if server.PostDown != "iptables -D" {
		t.Fatalf("PostDown mismatch: %s", server.PostDown)
	}
	if server.FallbackWGBin != "/usr/bin/wg" {
		t.Fatalf("FallbackWGBin mismatch: %s", server.FallbackWGBin)
	}
	if server.PersistentKeepalive != 30 {
		t.Fatalf("PersistentKeepalive mismatch: %d", server.PersistentKeepalive)
	}
	if len(server.Networks) != 1 {
		t.Fatalf("expected 1 network, got %d", len(server.Networks))
	}

	// Check peers were converted
	if len(server.Peers) != 1 {
		t.Fatalf("expected 1 peer, got %d", len(server.Peers))
	}
	if server.Peers[0].Hostname != "peer1" {
		t.Fatalf("peer hostname mismatch: %s", server.Peers[0].Hostname)
	}
	if server.Peers[0].PublicKey.Key != peerKey.PublicKey() {
		t.Fatal("peer public key mismatch")
	}
}

func TestGetServerEmptyPeers(t *testing.T) {
	config := testDsnetConfig(t)
	server := GetServer(config)

	if len(server.Peers) != 0 {
		t.Fatalf("expected 0 peers, got %d", len(server.Peers))
	}
}
