package lib

import (
	"net"
	"strings"
	"testing"

	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

func testPeerAndServer(t *testing.T) (Peer, Server) {
	t.Helper()
	privKey, err := wgtypes.GeneratePrivateKey()
	if err != nil {
		t.Fatalf("failed to generate private key: %v", err)
	}
	peerPrivKey, err := wgtypes.GeneratePrivateKey()
	if err != nil {
		t.Fatalf("failed to generate peer private key: %v", err)
	}
	psk, err := wgtypes.GenerateKey()
	if err != nil {
		t.Fatalf("failed to generate preshared key: %v", err)
	}

	server := Server{
		ExternalHostname: "vpn.example.com",
		ListenPort:       51820,
		Domain:           "dsnet",
		InterfaceName:    "dsnet",
		Network: JSONIPNet{
			IPNet: net.IPNet{
				IP:   net.IP{10, 0, 0, 0},
				Mask: net.IPMask{255, 255, 252, 0},
			},
		},
		Network6: JSONIPNet{
			IPNet: net.IPNet{
				IP:   net.IP{0xfd, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
				Mask: net.IPMask{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0, 0, 0, 0, 0, 0, 0, 0},
			},
		},
		IP:                  net.IP{10, 0, 0, 1},
		IP6:                 net.IP{0xfd, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1},
		DNS:                 net.IP{10, 0, 0, 1},
		PrivateKey:          JSONKey{Key: privKey},
		Networks:            []JSONIPNet{},
		PersistentKeepalive: 25,
	}

	peer := Peer{
		Hostname:            "test-peer",
		Owner:               "alice",
		Description:         "Test peer",
		IP:                  net.IP{10, 0, 0, 2},
		IP6:                 net.IP{0xfd, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 2},
		PublicKey:           JSONKey{Key: peerPrivKey.PublicKey()},
		PrivateKey:          JSONKey{Key: peerPrivKey},
		PresharedKey:        JSONKey{Key: psk},
		Networks:            []JSONIPNet{},
		PersistentKeepalive: 25,
	}

	return peer, server
}

func TestGetWGPeerTemplateWGQuick(t *testing.T) {
	peer, server := testPeerAndServer(t)

	buf, err := GetWGPeerTemplate(peer, WGQuick, server)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()

	if !strings.Contains(output, "[Interface]") {
		t.Fatal("wg-quick config should contain [Interface]")
	}
	if !strings.Contains(output, "[Peer]") {
		t.Fatal("wg-quick config should contain [Peer]")
	}
	if !strings.Contains(output, peer.PrivateKey.Key.String()) {
		t.Fatal("config should contain peer private key")
	}
	if !strings.Contains(output, server.PrivateKey.Key.PublicKey().String()) {
		t.Fatal("config should contain server public key")
	}
	if !strings.Contains(output, peer.PresharedKey.Key.String()) {
		t.Fatal("config should contain preshared key")
	}
	if !strings.Contains(output, "vpn.example.com:51820") {
		t.Fatal("config should contain endpoint")
	}
	if !strings.Contains(output, "DNS=10.0.0.1") {
		t.Fatal("config should contain DNS")
	}
	if !strings.Contains(output, "Address=10.0.0.2/") {
		t.Fatal("config should contain peer IPv4 address")
	}
}

func TestGetWGPeerTemplateVyatta(t *testing.T) {
	peer, server := testPeerAndServer(t)

	buf, err := GetWGPeerTemplate(peer, Vyatta, server)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()

	if !strings.Contains(output, "configure") {
		t.Fatal("vyatta config should start with 'configure'")
	}
	if !strings.Contains(output, "set interfaces wireguard") {
		t.Fatal("vyatta config should contain wireguard interface setup")
	}
	if !strings.Contains(output, "commit; save") {
		t.Fatal("vyatta config should end with 'commit; save'")
	}
}

func TestGetWGPeerTemplateNixOS(t *testing.T) {
	peer, server := testPeerAndServer(t)

	buf, err := GetWGPeerTemplate(peer, NixOS, server)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()

	if !strings.Contains(output, "networking.wireguard.interfaces") {
		t.Fatal("NixOS config should contain networking.wireguard.interfaces")
	}
	if !strings.Contains(output, "privateKey") {
		t.Fatal("NixOS config should contain privateKey")
	}
	if !strings.Contains(output, "publicKey") {
		t.Fatal("NixOS config should contain publicKey")
	}
}

func TestGetWGPeerTemplateRouterOS(t *testing.T) {
	peer, server := testPeerAndServer(t)

	buf, err := GetWGPeerTemplate(peer, RouterOS, server)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()

	if !strings.Contains(output, "/interface wireguard") {
		t.Fatal("RouterOS config should contain /interface wireguard")
	}
	if !strings.Contains(output, "private-key=") {
		t.Fatal("RouterOS config should contain private-key")
	}
}

func TestGetWGPeerTemplateInvalidType(t *testing.T) {
	peer, server := testPeerAndServer(t)

	_, err := GetWGPeerTemplate(peer, PeerType(99), server)
	if err == nil {
		t.Fatal("expected error for invalid peer type")
	}
}

func TestAsciiPeerConfig(t *testing.T) {
	tests := []struct {
		name     string
		peerType string
		contains string
	}{
		{"wg-quick", "wg-quick", "[Interface]"},
		{"vyatta", "vyatta", "configure"},
		{"nixos", "nixos", "networking.wireguard"},
		{"routeros", "routeros", "/interface wireguard"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			peer, server := testPeerAndServer(t)
			buf, err := AsciiPeerConfig(peer, tt.peerType, server)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !strings.Contains(buf.String(), tt.contains) {
				t.Fatalf("%s output should contain %q", tt.name, tt.contains)
			}
		})
	}
}

func TestAsciiPeerConfigInvalid(t *testing.T) {
	peer, server := testPeerAndServer(t)

	_, err := AsciiPeerConfig(peer, "invalid", server)
	if err == nil {
		t.Fatal("expected error for invalid output type")
	}
}

func TestGetWGPeerTemplateEndpointPrecedence(t *testing.T) {
	t.Run("hostname", func(t *testing.T) {
		peer, server := testPeerAndServer(t)
		buf, err := GetWGPeerTemplate(peer, WGQuick, server)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(buf.String(), "vpn.example.com:51820") {
			t.Fatal("should use ExternalHostname when set")
		}
	})

	t.Run("ipv4_fallback", func(t *testing.T) {
		peer, server := testPeerAndServer(t)
		server.ExternalHostname = ""
		server.ExternalIP = net.IP{1, 2, 3, 4}
		buf, err := GetWGPeerTemplate(peer, WGQuick, server)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(buf.String(), "1.2.3.4:51820") {
			t.Fatal("should use ExternalIP when hostname is empty")
		}
	})

	t.Run("ipv6_fallback", func(t *testing.T) {
		peer, server := testPeerAndServer(t)
		server.ExternalHostname = ""
		server.ExternalIP = nil
		server.ExternalIP6 = net.ParseIP("2001:db8::1")
		buf, err := GetWGPeerTemplate(peer, WGQuick, server)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(buf.String(), "[2001:db8::1]:51820") {
			t.Fatal("should use ExternalIP6 as last resort")
		}
	})
}

func TestGetWGPeerTemplateNoEndpoint(t *testing.T) {
	peer, server := testPeerAndServer(t)
	server.ExternalHostname = ""
	server.ExternalIP = nil
	server.ExternalIP6 = nil

	_, err := GetWGPeerTemplate(peer, WGQuick, server)
	if err == nil {
		t.Fatal("expected error when no endpoint is available")
	}
}

func TestGetWGPeerTemplateIPv4Only(t *testing.T) {
	peer, server := testPeerAndServer(t)
	peer.IP6 = nil
	server.Network6 = JSONIPNet{}

	buf, err := GetWGPeerTemplate(peer, WGQuick, server)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Address=10.0.0.2/") {
		t.Fatal("should contain IPv4 address")
	}
}

func TestGetWGPeerTemplateWithServerNetworks(t *testing.T) {
	peer, server := testPeerAndServer(t)
	_, extraNet, _ := net.ParseCIDR("192.168.1.0/24")
	server.Networks = []JSONIPNet{{IPNet: *extraNet}}

	buf, err := GetWGPeerTemplate(peer, WGQuick, server)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "192.168.1.0/24") {
		t.Fatal("should contain extra server network in AllowedIPs")
	}
}

func TestGetWGPeerTemplateNoDNS(t *testing.T) {
	peer, server := testPeerAndServer(t)
	server.DNS = nil

	buf, err := GetWGPeerTemplate(peer, WGQuick, server)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if strings.Contains(output, "DNS=") {
		t.Fatal("should not contain DNS line when DNS is nil")
	}
}
