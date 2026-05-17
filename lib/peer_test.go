package lib

import (
	"net"
	"testing"

	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

func TestNewPeerBasic(t *testing.T) {
	s := testServer(t)

	peer, err := NewPeer(s, "", "", "alice", "laptop", "Alice's laptop")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if peer.Hostname != "laptop" {
		t.Fatalf("expected hostname 'laptop', got '%s'", peer.Hostname)
	}
	if peer.Owner != "alice" {
		t.Fatalf("expected owner 'alice', got '%s'", peer.Owner)
	}
	if peer.Description != "Alice's laptop" {
		t.Fatalf("expected description \"Alice's laptop\", got '%s'", peer.Description)
	}
	if peer.Added.IsZero() {
		t.Fatal("Added time should not be zero")
	}
	if peer.PersistentKeepalive != 25 {
		t.Fatalf("expected PersistentKeepalive 25, got %d", peer.PersistentKeepalive)
	}
}

func TestNewPeerGeneratesKeys(t *testing.T) {
	s := testServer(t)

	peer, err := NewPeer(s, "", "", "alice", "laptop", "test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	zero := wgtypes.Key{}
	if peer.PrivateKey.Key == zero {
		t.Fatal("private key should not be zero")
	}
	if peer.PublicKey.Key == zero {
		t.Fatal("public key should not be zero")
	}
	if peer.PresharedKey.Key == zero {
		t.Fatal("preshared key should not be zero")
	}

	// Public key should derive from private key
	expected := peer.PrivateKey.Key.PublicKey()
	if peer.PublicKey.Key != expected {
		t.Fatal("public key should derive from private key")
	}
}

func TestNewPeerAllocatesIP(t *testing.T) {
	s := testServer(t)

	peer, err := NewPeer(s, "", "", "alice", "laptop", "test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(peer.IP) == 0 {
		t.Fatal("peer should have an IPv4 address")
	}
	if len(peer.IP6) == 0 {
		t.Fatal("peer should have an IPv6 address")
	}

	// IPv4 should be in the 10.0.0.0/22 range
	network := s.Network.IPNet
	if !network.Contains(peer.IP) {
		t.Fatalf("peer IP %s not in network %s", peer.IP, network.String())
	}
}

func TestNewPeerSequentialIPs(t *testing.T) {
	s := testServer(t)

	peer1, err := NewPeer(s, "", "", "alice", "laptop1", "test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Add peer1 to server so peer2 gets a different IP
	s.Peers = append(s.Peers, peer1)

	peer2, err := NewPeer(s, "", "", "bob", "laptop2", "test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if peer1.IP.Equal(peer2.IP) {
		t.Fatal("two peers should have different IPs")
	}
}

func TestNewPeerMissingOwner(t *testing.T) {
	s := testServer(t)

	_, err := NewPeer(s, "", "", "", "laptop", "test")
	if err == nil {
		t.Fatal("expected error for missing owner")
	}
}

func TestNewPeerMissingHostname(t *testing.T) {
	s := testServer(t)

	_, err := NewPeer(s, "", "", "alice", "", "test")
	if err == nil {
		t.Fatal("expected error for missing hostname")
	}
}

func TestNewPeerWithPrivateKey(t *testing.T) {
	s := testServer(t)

	privKey, _ := wgtypes.GeneratePrivateKey()
	b64 := privKey.String()

	peer, err := NewPeer(s, b64, "", "alice", "laptop", "test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if peer.PrivateKey.Key != privKey {
		t.Fatal("private key should match the provided key")
	}
	if peer.PublicKey.Key != privKey.PublicKey() {
		t.Fatal("public key should derive from the provided private key")
	}
}

func TestNewPeerWithPublicKeyOnly(t *testing.T) {
	s := testServer(t)

	privKey, _ := wgtypes.GeneratePrivateKey()
	pubKey := privKey.PublicKey()

	peer, err := NewPeer(s, "", pubKey.String(), "alice", "laptop", "test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if peer.PublicKey.Key != pubKey {
		t.Fatal("public key should match the provided key")
	}

	// Private key should be zeroed when only public key is provided
	zeroKey := wgtypes.Key([wgtypes.KeyLen]byte{})
	if peer.PrivateKey.Key != zeroKey {
		t.Fatal("private key should be zeroed when only public key is provided")
	}
}

func TestNewPeerWithMatchingKeyPair(t *testing.T) {
	s := testServer(t)

	privKey, _ := wgtypes.GeneratePrivateKey()
	pubKey := privKey.PublicKey()

	peer, err := NewPeer(s, privKey.String(), pubKey.String(), "alice", "laptop", "test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if peer.PrivateKey.Key != privKey {
		t.Fatal("private key mismatch")
	}
	if peer.PublicKey.Key != pubKey {
		t.Fatal("public key mismatch")
	}
}

func TestNewPeerWithMismatchedKeys(t *testing.T) {
	s := testServer(t)

	privKey1, _ := wgtypes.GeneratePrivateKey()
	privKey2, _ := wgtypes.GeneratePrivateKey()
	pubKey2 := privKey2.PublicKey()

	_, err := NewPeer(s, privKey1.String(), pubKey2.String(), "alice", "laptop", "test")
	if err == nil {
		t.Fatal("expected error for mismatched key pair")
	}
}

func TestNewPeerNoNetwork(t *testing.T) {
	s := testServer(t)
	s.IP = nil
	s.IP6 = nil
	s.Network = JSONIPNet{}
	s.Network6 = JSONIPNet{}

	_, err := NewPeer(s, "", "", "alice", "laptop", "test")
	if err == nil {
		t.Fatal("expected error when no network is defined")
	}
}

func TestNewPeerIPv4Only(t *testing.T) {
	s := testServer(t)
	s.IP6 = nil
	s.Network6 = JSONIPNet{}

	peer, err := NewPeer(s, "", "", "alice", "laptop", "test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(peer.IP) == 0 {
		t.Fatal("peer should have IPv4")
	}
	if len(peer.IP6) != 0 {
		t.Fatal("peer should not have IPv6")
	}
}

func TestNewPeerEmptyNetworks(t *testing.T) {
	s := testServer(t)

	peer, err := NewPeer(s, "", "", "alice", "laptop", "test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if peer.Networks == nil {
		t.Fatal("Networks should be initialized, not nil")
	}
	if len(peer.Networks) != 0 {
		t.Fatalf("Networks should be empty, got %d", len(peer.Networks))
	}
}

func TestPeerGetIfName(t *testing.T) {
	p := &Peer{
		IP:  net.IP{10, 0, 0, 5},
		IP6: net.IP{},
	}
	name := p.GetIfName()
	// 10+0+0+5 = 15, 15%999 = 15
	if name != "wg15" {
		t.Fatalf("expected wg15, got %s", name)
	}
}

func TestPeerGetIfNameDeterministic(t *testing.T) {
	p := &Peer{
		IP:  net.IP{10, 0, 0, 5},
		IP6: net.IP{0xfd, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1},
	}
	name1 := p.GetIfName()
	name2 := p.GetIfName()
	if name1 != name2 {
		t.Fatalf("getIfName should be deterministic: %s != %s", name1, name2)
	}
}
