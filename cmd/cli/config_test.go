package cli

import (
	"net"
	"testing"
	"time"

	"github.com/naggie/dsnet/lib"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

func testDsnetConfig(t *testing.T) *DsnetConfig {
	t.Helper()
	privKey, err := wgtypes.GeneratePrivateKey()
	if err != nil {
		t.Fatalf("failed to generate key: %v", err)
	}
	return &DsnetConfig{
		ExternalHostname: "vpn.example.com",
		ListenPort:       51820,
		Domain:           "dsnet",
		InterfaceName:    "dsnet",
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
		IP:                  net.IP{10, 0, 0, 1},
		IP6:                 net.IP{0xfd, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1},
		PrivateKey:          lib.JSONKey{Key: privKey},
		Peers:               []PeerConfig{},
		Networks:            []lib.JSONIPNet{},
		PersistentKeepalive: 25,
	}
}

func testLibPeer(t *testing.T, hostname, owner string, ip net.IP) lib.Peer {
	t.Helper()
	privKey, err := wgtypes.GeneratePrivateKey()
	if err != nil {
		t.Fatalf("failed to generate key: %v", err)
	}
	psk, err := wgtypes.GenerateKey()
	if err != nil {
		t.Fatalf("failed to generate preshared key: %v", err)
	}
	return lib.Peer{
		Hostname:     hostname,
		Owner:        owner,
		Description:  "test peer",
		IP:           ip,
		IP6:          net.IP{0xfd, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, ip[3]},
		Added:        time.Now(),
		PublicKey:    lib.JSONKey{Key: privKey.PublicKey()},
		PrivateKey:   lib.JSONKey{Key: privKey},
		PresharedKey: lib.JSONKey{Key: psk},
		Networks:     []lib.JSONIPNet{},
	}
}

func TestAddPeerBasic(t *testing.T) {
	conf := testDsnetConfig(t)
	peer := testLibPeer(t, "laptop", "alice", net.IP{10, 0, 0, 2})

	err := conf.AddPeer(peer)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(conf.Peers) != 1 {
		t.Fatalf("expected 1 peer, got %d", len(conf.Peers))
	}

	p := conf.Peers[0]
	if p.Hostname != "laptop" {
		t.Fatalf("expected hostname 'laptop', got '%s'", p.Hostname)
	}
	if p.Owner != "alice" {
		t.Fatalf("expected owner 'alice', got '%s'", p.Owner)
	}
	if !p.IP.Equal(net.IP{10, 0, 0, 2}) {
		t.Fatalf("expected IP 10.0.0.2, got %s", p.IP)
	}
}

func TestAddPeerMultiple(t *testing.T) {
	conf := testDsnetConfig(t)

	peer1 := testLibPeer(t, "laptop1", "alice", net.IP{10, 0, 0, 2})
	peer2 := testLibPeer(t, "laptop2", "bob", net.IP{10, 0, 0, 3})

	if err := conf.AddPeer(peer1); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := conf.AddPeer(peer2); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(conf.Peers) != 2 {
		t.Fatalf("expected 2 peers, got %d", len(conf.Peers))
	}
}

func TestAddPeerDuplicateHostname(t *testing.T) {
	conf := testDsnetConfig(t)
	peer1 := testLibPeer(t, "laptop", "alice", net.IP{10, 0, 0, 2})
	peer2 := testLibPeer(t, "laptop", "bob", net.IP{10, 0, 0, 3})

	if err := conf.AddPeer(peer1); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err := conf.AddPeer(peer2)
	if err == nil {
		t.Fatal("expected error for duplicate hostname")
	}
}

func TestAddPeerDuplicatePublicKey(t *testing.T) {
	conf := testDsnetConfig(t)
	peer1 := testLibPeer(t, "laptop1", "alice", net.IP{10, 0, 0, 2})
	peer2 := testLibPeer(t, "laptop2", "bob", net.IP{10, 0, 0, 3})

	// Make them share the same public key
	peer2.PublicKey = peer1.PublicKey

	if err := conf.AddPeer(peer1); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err := conf.AddPeer(peer2)
	if err == nil {
		t.Fatal("expected error for duplicate public key")
	}
}

func TestAddPeerDuplicatePresharedKey(t *testing.T) {
	conf := testDsnetConfig(t)
	peer1 := testLibPeer(t, "laptop1", "alice", net.IP{10, 0, 0, 2})
	peer2 := testLibPeer(t, "laptop2", "bob", net.IP{10, 0, 0, 3})

	// Make them share the same preshared key
	peer2.PresharedKey = peer1.PresharedKey

	if err := conf.AddPeer(peer1); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err := conf.AddPeer(peer2)
	if err == nil {
		t.Fatal("expected error for duplicate preshared key")
	}
}

func TestAddPeerFieldMapping(t *testing.T) {
	conf := testDsnetConfig(t)
	peer := testLibPeer(t, "laptop", "alice", net.IP{10, 0, 0, 2})
	peer.Description = "specific description"

	if err := conf.AddPeer(peer); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	p := conf.Peers[0]
	if p.Description != "specific description" {
		t.Fatalf("description not mapped correctly: got '%s'", p.Description)
	}
	if p.PublicKey.Key != peer.PublicKey.Key {
		t.Fatal("public key not mapped correctly")
	}
	if p.PrivateKey.Key != peer.PrivateKey.Key {
		t.Fatal("private key not mapped correctly")
	}
	if p.PresharedKey.Key != peer.PresharedKey.Key {
		t.Fatal("preshared key not mapped correctly")
	}
	if !p.Added.Equal(peer.Added) {
		t.Fatal("Added time not mapped correctly")
	}
}

func TestRemovePeerBasic(t *testing.T) {
	conf := testDsnetConfig(t)
	peer := testLibPeer(t, "laptop", "alice", net.IP{10, 0, 0, 2})
	conf.AddPeer(peer)

	err := conf.RemovePeer("laptop")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(conf.Peers) != 0 {
		t.Fatalf("expected 0 peers, got %d", len(conf.Peers))
	}
}

func TestRemovePeerNotFound(t *testing.T) {
	conf := testDsnetConfig(t)

	err := conf.RemovePeer("nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent peer")
	}
}

func TestRemovePeerRetainsOrder(t *testing.T) {
	conf := testDsnetConfig(t)
	peer1 := testLibPeer(t, "laptop1", "alice", net.IP{10, 0, 0, 2})
	peer2 := testLibPeer(t, "laptop2", "bob", net.IP{10, 0, 0, 3})
	peer3 := testLibPeer(t, "laptop3", "carol", net.IP{10, 0, 0, 4})

	conf.AddPeer(peer1)
	conf.AddPeer(peer2)
	conf.AddPeer(peer3)

	// Remove the middle peer
	err := conf.RemovePeer("laptop2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(conf.Peers) != 2 {
		t.Fatalf("expected 2 peers, got %d", len(conf.Peers))
	}
	if conf.Peers[0].Hostname != "laptop1" {
		t.Fatalf("expected first peer 'laptop1', got '%s'", conf.Peers[0].Hostname)
	}
	if conf.Peers[1].Hostname != "laptop3" {
		t.Fatalf("expected second peer 'laptop3', got '%s'", conf.Peers[1].Hostname)
	}
}

func TestRemovePeerFirst(t *testing.T) {
	conf := testDsnetConfig(t)
	peer1 := testLibPeer(t, "laptop1", "alice", net.IP{10, 0, 0, 2})
	peer2 := testLibPeer(t, "laptop2", "bob", net.IP{10, 0, 0, 3})

	conf.AddPeer(peer1)
	conf.AddPeer(peer2)

	err := conf.RemovePeer("laptop1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(conf.Peers) != 1 {
		t.Fatalf("expected 1 peer, got %d", len(conf.Peers))
	}
	if conf.Peers[0].Hostname != "laptop2" {
		t.Fatalf("expected 'laptop2', got '%s'", conf.Peers[0].Hostname)
	}
}

func TestRemovePeerLast(t *testing.T) {
	conf := testDsnetConfig(t)
	peer1 := testLibPeer(t, "laptop1", "alice", net.IP{10, 0, 0, 2})
	peer2 := testLibPeer(t, "laptop2", "bob", net.IP{10, 0, 0, 3})

	conf.AddPeer(peer1)
	conf.AddPeer(peer2)

	err := conf.RemovePeer("laptop2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(conf.Peers) != 1 {
		t.Fatalf("expected 1 peer, got %d", len(conf.Peers))
	}
	if conf.Peers[0].Hostname != "laptop1" {
		t.Fatalf("expected 'laptop1', got '%s'", conf.Peers[0].Hostname)
	}
}
