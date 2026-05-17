package cli

import (
	"net"
	"testing"
	"time"

	"github.com/naggie/dsnet/lib"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

func TestBytesToSI(t *testing.T) {
	tests := []struct {
		input uint64
		want  string
	}{
		{0, "0 B"},
		{999, "999 B"},
		{1000, "1.0 kB"},
		{1500, "1.5 kB"},
		{1000000, "1.0 MB"},
		{1000000000, "1.0 GB"},
		{1000000000000, "1.0 TB"},
		{2500000000, "2.5 GB"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := BytesToSI(tt.input); got != tt.want {
				t.Fatalf("BytesToSI(%d) = %q, want %q", tt.input, got, tt.want)
			}
		})
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
