package cli

import (
	"net"
	"testing"
	"time"

	"github.com/naggie/dsnet/lib"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

func TestBytesToSIZero(t *testing.T) {
	result := BytesToSI(0)
	if result != "0 B" {
		t.Fatalf("expected '0 B', got '%s'", result)
	}
}

func TestBytesToSIBytes(t *testing.T) {
	result := BytesToSI(999)
	if result != "999 B" {
		t.Fatalf("expected '999 B', got '%s'", result)
	}
}

func TestBytesToSIKilo(t *testing.T) {
	result := BytesToSI(1000)
	if result != "1.0 kB" {
		t.Fatalf("expected '1.0 kB', got '%s'", result)
	}
}

func TestBytesToSIKiloFractional(t *testing.T) {
	result := BytesToSI(1500)
	if result != "1.5 kB" {
		t.Fatalf("expected '1.5 kB', got '%s'", result)
	}
}

func TestBytesToSIMega(t *testing.T) {
	result := BytesToSI(1000000)
	if result != "1.0 MB" {
		t.Fatalf("expected '1.0 MB', got '%s'", result)
	}
}

func TestBytesToSIGiga(t *testing.T) {
	result := BytesToSI(1000000000)
	if result != "1.0 GB" {
		t.Fatalf("expected '1.0 GB', got '%s'", result)
	}
}

func TestBytesToSITera(t *testing.T) {
	result := BytesToSI(1000000000000)
	if result != "1.0 TB" {
		t.Fatalf("expected '1.0 TB', got '%s'", result)
	}
}

func TestBytesToSILargeValue(t *testing.T) {
	// 2.5 GB
	result := BytesToSI(2500000000)
	if result != "2.5 GB" {
		t.Fatalf("expected '2.5 GB', got '%s'", result)
	}
}

func TestJsonPeerToDsnetPeerEmpty(t *testing.T) {
	result := jsonPeerToDsnetPeer([]PeerConfig{})
	if len(result) != 0 {
		t.Fatalf("expected 0 peers, got %d", len(result))
	}
}

func TestJsonPeerToDsnetPeer(t *testing.T) {
	privKey, _ := wgtypes.GeneratePrivateKey()
	psk, _ := wgtypes.GenerateKey()
	now := time.Now()
	_, subnet, _ := net.ParseCIDR("192.168.1.0/24")

	input := []PeerConfig{
		{
			Hostname:     "laptop",
			Owner:        "alice",
			Description:  "Alice's laptop",
			IP:           net.IP{10, 0, 0, 2},
			IP6:          net.IP{0xfd, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 2},
			Added:        now,
			PublicKey:    lib.JSONKey{Key: privKey.PublicKey()},
			PrivateKey:   lib.JSONKey{Key: privKey},
			PresharedKey: lib.JSONKey{Key: psk},
			Networks:     []lib.JSONIPNet{{IPNet: *subnet}},
		},
	}

	result := jsonPeerToDsnetPeer(input)
	if len(result) != 1 {
		t.Fatalf("expected 1 peer, got %d", len(result))
	}

	p := result[0]
	if p.Hostname != "laptop" {
		t.Fatalf("expected hostname 'laptop', got '%s'", p.Hostname)
	}
	if p.Owner != "alice" {
		t.Fatalf("expected owner 'alice', got '%s'", p.Owner)
	}
	if p.Description != "Alice's laptop" {
		t.Fatalf("expected description 'Alice's laptop', got '%s'", p.Description)
	}
	if !p.IP.Equal(net.IP{10, 0, 0, 2}) {
		t.Fatalf("expected IP 10.0.0.2, got %s", p.IP)
	}
	if !p.Added.Equal(now) {
		t.Fatal("Added time mismatch")
	}
	if p.PublicKey.Key != privKey.PublicKey() {
		t.Fatal("public key mismatch")
	}
	if p.PrivateKey.Key != privKey {
		t.Fatal("private key mismatch")
	}
	if p.PresharedKey.Key != psk {
		t.Fatal("preshared key mismatch")
	}
	if len(p.Networks) != 1 {
		t.Fatalf("expected 1 network, got %d", len(p.Networks))
	}
}

func TestJsonPeerToDsnetPeerMultiple(t *testing.T) {
	privKey1, _ := wgtypes.GeneratePrivateKey()
	privKey2, _ := wgtypes.GeneratePrivateKey()
	psk1, _ := wgtypes.GenerateKey()
	psk2, _ := wgtypes.GenerateKey()

	input := []PeerConfig{
		{
			Hostname:     "laptop1",
			Owner:        "alice",
			Description:  "first",
			IP:           net.IP{10, 0, 0, 2},
			Added:        time.Now(),
			PublicKey:    lib.JSONKey{Key: privKey1.PublicKey()},
			PrivateKey:   lib.JSONKey{Key: privKey1},
			PresharedKey: lib.JSONKey{Key: psk1},
			Networks:     []lib.JSONIPNet{},
		},
		{
			Hostname:     "laptop2",
			Owner:        "bob",
			Description:  "second",
			IP:           net.IP{10, 0, 0, 3},
			Added:        time.Now(),
			PublicKey:    lib.JSONKey{Key: privKey2.PublicKey()},
			PrivateKey:   lib.JSONKey{Key: privKey2},
			PresharedKey: lib.JSONKey{Key: psk2},
			Networks:     []lib.JSONIPNet{},
		},
	}

	result := jsonPeerToDsnetPeer(input)
	if len(result) != 2 {
		t.Fatalf("expected 2 peers, got %d", len(result))
	}
	if result[0].Hostname != "laptop1" {
		t.Fatalf("expected 'laptop1', got '%s'", result[0].Hostname)
	}
	if result[1].Hostname != "laptop2" {
		t.Fatalf("expected 'laptop2', got '%s'", result[1].Hostname)
	}
}
