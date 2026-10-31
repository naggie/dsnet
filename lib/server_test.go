package lib

import (
	"encoding/json"
	"net"
	"testing"

	"github.com/vishvananda/netlink"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

func testServer(t *testing.T) *Server {
	t.Helper()
	privKey, err := wgtypes.GeneratePrivateKey()
	if err != nil {
		t.Fatalf("failed to generate key: %v", err)
	}
	return &Server{
		ExternalHostname: "vpn.example.com",
		ListenPort:       51820,
		Domain:           "dsnet",
		InterfaceName:    "wg0",
		Network: JSONIPNet{
			IPNet: net.IPNet{
				IP:   net.IP{10, 0, 0, 0},
				Mask: net.IPMask{255, 255, 252, 0}, // /22
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
		PrivateKey:          JSONKey{Key: privKey},
		Peers:               []Peer{},
		Networks:            []JSONIPNet{},
		PersistentKeepalive: 25,
	}
}

func TestAllocateIPFirst(t *testing.T) {
	s := testServer(t)

	ip, err := s.AllocateIP()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Server has .1, so first allocation should be .2
	expected := net.IP{10, 0, 0, 2}
	if !ip.Equal(expected) {
		t.Fatalf("expected %s, got %s", expected, ip)
	}
}

func TestAllocateIPSequential(t *testing.T) {
	s := testServer(t)

	// Add a peer at .2
	s.Peers = append(s.Peers, Peer{
		IP: net.IP{10, 0, 0, 2},
	})

	ip, err := s.AllocateIP()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := net.IP{10, 0, 0, 3}
	if !ip.Equal(expected) {
		t.Fatalf("expected %s, got %s", expected, ip)
	}
}

func TestAllocateIPSkipsServer(t *testing.T) {
	s := testServer(t)

	// Verify that the server IP (.1) is skipped
	if !s.IPAllocated(net.IP{10, 0, 0, 1}) {
		t.Fatal("server IP should be marked as allocated")
	}
}

func TestAllocateIPExhausted(t *testing.T) {
	s := testServer(t)
	// Use a /30 network (only 2 usable IPs: .1 and .2)
	s.Network = JSONIPNet{
		IPNet: net.IPNet{
			IP:   net.IP{10, 0, 0, 0},
			Mask: net.IPMask{255, 255, 255, 252}, // /30
		},
	}
	// Server takes .1
	s.IP = net.IP{10, 0, 0, 1}
	// Peer takes .2
	s.Peers = []Peer{{IP: net.IP{10, 0, 0, 2}}}

	_, err := s.AllocateIP()
	if err == nil {
		t.Fatal("expected error for exhausted IP range")
	}
}

func TestAllocateIP6(t *testing.T) {
	s := testServer(t)

	ip, err := s.AllocateIP6()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should be in the fd00::/64 network
	if ip[0] != 0xfd {
		t.Fatalf("expected fd prefix, got %x", ip[0])
	}

	// Should not equal the server IP
	if ip.Equal(s.IP6) {
		t.Fatal("allocated IP should not equal server IP")
	}
}

func TestAllocateIP6Unique(t *testing.T) {
	s := testServer(t)

	ip1, err := s.AllocateIP6()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	s.Peers = append(s.Peers, Peer{IP6: ip1})

	ip2, err := s.AllocateIP6()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ip1.Equal(ip2) {
		t.Fatal("two sequential IPv6 allocations should differ")
	}
}

func TestIPAllocatedServerIP(t *testing.T) {
	s := testServer(t)

	if !s.IPAllocated(s.IP) {
		t.Fatal("server IPv4 should be allocated")
	}
	if !s.IPAllocated(s.IP6) {
		t.Fatal("server IPv6 should be allocated")
	}
}

func TestIPAllocatedPeerIP(t *testing.T) {
	s := testServer(t)
	peerIP := net.IP{10, 0, 0, 5}
	s.Peers = append(s.Peers, Peer{IP: peerIP})

	if !s.IPAllocated(peerIP) {
		t.Fatal("peer IP should be allocated")
	}

	if s.IPAllocated(net.IP{10, 0, 0, 6}) {
		t.Fatal("unused IP should not be allocated")
	}
}

func TestIPAllocatedPeerNetwork(t *testing.T) {
	s := testServer(t)
	_, subnet, _ := net.ParseCIDR("192.168.1.0/24")
	s.Peers = append(s.Peers, Peer{
		IP: net.IP{10, 0, 0, 5},
		Networks: []JSONIPNet{
			{IPNet: *subnet},
		},
	})

	if !s.IPAllocated(net.IP{192, 168, 1, 0}) {
		t.Fatal("peer network IP should be allocated")
	}
}

func TestGetPeersEmpty(t *testing.T) {
	s := testServer(t)
	peers := s.GetPeers()
	if len(peers) != 0 {
		t.Fatalf("expected 0 peers, got %d", len(peers))
	}
}

func TestGetPeersWithPeer(t *testing.T) {
	s := testServer(t)
	peerKey, _ := wgtypes.GeneratePrivateKey()
	psk, _ := wgtypes.GenerateKey()

	s.Peers = append(s.Peers, Peer{
		Hostname:     "test-peer",
		IP:           net.IP{10, 0, 0, 2},
		IP6:          net.IP{0xfd, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 2},
		PublicKey:    JSONKey{Key: peerKey.PublicKey()},
		PresharedKey: JSONKey{Key: psk},
		Networks:     []JSONIPNet{},
	})

	wgPeers := s.GetPeers()
	if len(wgPeers) != 1 {
		t.Fatalf("expected 1 peer, got %d", len(wgPeers))
	}

	p := wgPeers[0]
	if p.PublicKey != peerKey.PublicKey() {
		t.Fatal("public key mismatch")
	}
	if p.PresharedKey == nil {
		t.Fatal("preshared key should not be nil")
	}
	if *p.PresharedKey != psk {
		t.Fatal("preshared key mismatch")
	}
	if !p.ReplaceAllowedIPs {
		t.Fatal("ReplaceAllowedIPs should be true")
	}
	if p.Remove {
		t.Fatal("Remove should be false")
	}

	// Should have 2 AllowedIPs: one /32 for IPv4 and one /128 for IPv6
	if len(p.AllowedIPs) != 2 {
		t.Fatalf("expected 2 AllowedIPs, got %d", len(p.AllowedIPs))
	}
}

func TestGetPeersAllowedIPsIncludesNetworks(t *testing.T) {
	s := testServer(t)
	peerKey, _ := wgtypes.GeneratePrivateKey()
	psk, _ := wgtypes.GenerateKey()
	_, subnet, _ := net.ParseCIDR("192.168.1.0/24")

	s.Peers = append(s.Peers, Peer{
		Hostname:     "test-peer",
		IP:           net.IP{10, 0, 0, 2},
		PublicKey:    JSONKey{Key: peerKey.PublicKey()},
		PresharedKey: JSONKey{Key: psk},
		Networks:     []JSONIPNet{{IPNet: *subnet}},
	})

	wgPeers := s.GetPeers()
	p := wgPeers[0]

	// Should have /32 for IP + 1 extra network = 2
	if len(p.AllowedIPs) != 2 {
		t.Fatalf("expected 2 AllowedIPs, got %d", len(p.AllowedIPs))
	}
}

func TestGetPeersPresharedKeyIsolation(t *testing.T) {
	s := testServer(t)
	key1, _ := wgtypes.GeneratePrivateKey()
	key2, _ := wgtypes.GeneratePrivateKey()
	psk1, _ := wgtypes.GenerateKey()
	psk2, _ := wgtypes.GenerateKey()

	s.Peers = append(s.Peers,
		Peer{
			Hostname:     "peer1",
			IP:           net.IP{10, 0, 0, 2},
			PublicKey:    JSONKey{Key: key1.PublicKey()},
			PresharedKey: JSONKey{Key: psk1},
			Networks:     []JSONIPNet{},
		},
		Peer{
			Hostname:     "peer2",
			IP:           net.IP{10, 0, 0, 3},
			PublicKey:    JSONKey{Key: key2.PublicKey()},
			PresharedKey: JSONKey{Key: psk2},
			Networks:     []JSONIPNet{},
		},
	)

	wgPeers := s.GetPeers()

	// Each peer should have its own preshared key (not sharing pointers)
	if *wgPeers[0].PresharedKey == *wgPeers[1].PresharedKey {
		t.Fatal("peers should have different preshared keys")
	}
	if *wgPeers[0].PresharedKey != psk1 {
		t.Fatal("peer1 preshared key mismatch")
	}
	if *wgPeers[1].PresharedKey != psk2 {
		t.Fatal("peer2 preshared key mismatch")
	}
}

func TestGetPeersOmitsPaused(t *testing.T) {
	s := testServer(t)
	activeKey, _ := wgtypes.GeneratePrivateKey()
	pausedKey, _ := wgtypes.GeneratePrivateKey()

	s.Peers = append(s.Peers,
		Peer{
			Hostname:  "paused",
			IP:        net.IP{10, 0, 0, 2},
			PublicKey: JSONKey{Key: pausedKey.PublicKey()},
			Paused:    true,
		},
		Peer{
			Hostname:  "active",
			IP:        net.IP{10, 0, 0, 3},
			PublicKey: JSONKey{Key: activeKey.PublicKey()},
		},
	)

	wgPeers := s.GetPeers()
	if len(wgPeers) != 1 {
		t.Fatalf("expected 1 peer, got %d", len(wgPeers))
	}
	if wgPeers[0].PublicKey != activeKey.PublicKey() {
		t.Fatal("expected only the active peer")
	}
}

func TestIPAllocatedPausedPeer(t *testing.T) {
	s := testServer(t)
	peerIP := net.IP{10, 0, 0, 2}
	s.Peers = append(s.Peers, Peer{IP: peerIP, Paused: true})

	// a paused peer must keep its IP so resume does not cause a clash
	if !s.IPAllocated(peerIP) {
		t.Fatal("paused peer IP should stay allocated")
	}
	ip, err := s.AllocateIP()
	if err != nil {
		t.Fatal(err)
	}
	if ip.Equal(peerIP) {
		t.Fatal("allocated IP of paused peer")
	}
}

func TestRoutedNetworks(t *testing.T) {
	s := testServer(t)

	// parse via JSON as the config file does: it keeps host bits, which the
	// kernel rejects in a route
	var hostBits, v6, blank, dup JSONIPNet
	for raw, n := range map[string]*JSONIPNet{
		`"10.182.1.5/16"`:    &hostBits,
		`"fd00:1:2:3::9/64"`: &v6,
		`""`:                 &blank,
		`"192.168.7.200/24"`: &dup,
	} {
		if err := json.Unmarshal([]byte(raw), n); err != nil {
			t.Fatalf("failed to parse %s: %v", raw, err)
		}
	}
	_, other, _ := net.ParseCIDR("192.168.7.0/24")

	s.Peers = []Peer{
		{Hostname: "a", IP: net.IP{10, 0, 0, 2}, Networks: []JSONIPNet{hostBits, blank, dup}},
		// paused peers keep their routes so traffic is dropped, not leaked
		{Hostname: "b", IP: net.IP{10, 0, 0, 3}, Paused: true, Networks: []JSONIPNet{v6, {IPNet: *other}}},
		{Hostname: "c", IP: net.IP{10, 0, 0, 4}, Networks: []JSONIPNet{}},
	}

	got := s.routedNetworks()

	want := []string{"10.182.0.0/16", "fd00:1:2:3::/64", "192.168.7.0/24"}
	if len(got) != len(want) {
		t.Fatalf("expected %v, got %v", want, got)
	}
	for _, cidr := range want {
		dst, ok := got[cidr]
		if !ok {
			t.Fatalf("expected route for %s, got %v", cidr, got)
		}
		if dst.String() != cidr {
			t.Fatalf("route key %s holds %s", cidr, dst.String())
		}
	}
}

func TestPlanRoutes(t *testing.T) {
	// RTPROT_BOOT, what `ip route add` uses by default
	const rtprotBoot = 3
	cidr := func(s string) net.IPNet {
		_, n, err := net.ParseCIDR(s)
		if err != nil {
			t.Fatalf("failed to parse %s: %v", s, err)
		}
		return *n
	}
	manual, ours, stale, unrelated, missing := cidr("2a11:29c0:11:120::/60"), cidr("10.1.0.0/16"), cidr("10.2.0.0/16"), cidr("10.3.0.0/16"), cidr("fd00::/64")

	existing := []netlink.Route{
		// added by hand or PostUp for a peer network: adopt, do not re-add
		{Dst: &manual, Protocol: rtprotBoot},
		{Dst: &ours, Protocol: dsnetRouteProtocol},
		// dsnet's own, no longer wanted
		{Dst: &stale, Protocol: dsnetRouteProtocol},
		// not dsnet's, never touched
		{Dst: &unrelated, Protocol: rtprotBoot},
		// e.g. a default route
		{Dst: nil, Protocol: rtprotBoot},
	}
	wanted := map[string]net.IPNet{
		manual.String():  manual,
		ours.String():    ours,
		missing.String(): missing,
	}

	toAdd, toDelete := planRoutes(existing, wanted)

	if len(toAdd) != 1 || toAdd[0].String() != missing.String() {
		t.Fatalf("expected to add only %s, got %v", missing.String(), toAdd)
	}
	if len(toDelete) != 1 || toDelete[0].Dst.String() != stale.String() {
		t.Fatalf("expected to delete only %s, got %v", stale.String(), toDelete)
	}
}
