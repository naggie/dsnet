package jsonfile

import (
	"context"
	"encoding/json"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/naggie/dsnet/lib"
	"github.com/naggie/dsnet/lib/store"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

func newTestBackend(t *testing.T) (*Backend, string) {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "dsnetconfig.json")
	be, err := Open(&url.URL{Scheme: "jsonfile", Path: path})
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	return be.(*Backend), path
}

func sampleServer(t *testing.T) *lib.Server {
	t.Helper()
	privKey, err := wgtypes.GeneratePrivateKey()
	if err != nil {
		t.Fatalf("GeneratePrivateKey: %v", err)
	}
	peerPriv, err := wgtypes.GeneratePrivateKey()
	if err != nil {
		t.Fatalf("GeneratePrivateKey (peer): %v", err)
	}
	psk, err := wgtypes.GenerateKey()
	if err != nil {
		t.Fatalf("GenerateKey (psk): %v", err)
	}
	return &lib.Server{
		ExternalHostname: "vpn.example.com",
		ExternalIP:       net.IP{1, 2, 3, 4},
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
		DNS:                 net.IP{10, 0, 0, 1},
		PrivateKey:          lib.JSONKey{Key: privKey},
		PostUp:              "iptables -A FORWARD -i dsnet -j ACCEPT",
		PostDown:            "iptables -D FORWARD -i dsnet -j ACCEPT",
		Networks:            []lib.JSONIPNet{},
		PersistentKeepalive: 25,
		MTU:                 1420,
		Peers: []lib.Peer{
			{
				Hostname:            "laptop",
				Owner:               "alice",
				Description:         "Alice's laptop",
				IP:                  net.IP{10, 0, 0, 2},
				IP6:                 net.IP{0xfd, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 2},
				Added:               time.Now().UTC().Truncate(time.Second),
				PublicKey:           lib.JSONKey{Key: peerPriv.PublicKey()},
				PrivateKey:          lib.JSONKey{Key: peerPriv},
				PresharedKey:        lib.JSONKey{Key: psk},
				Networks:            []lib.JSONIPNet{},
				PersistentKeepalive: 25,
			},
		},
	}
}

func wrapState(s *lib.Server) *store.State {
	return &store.State{Networks: map[string]*store.Network{s.InterfaceName: {Server: s}}}
}

func TestBackendRoundTrip(t *testing.T) {
	be, path := newTestBackend(t)
	defer be.Close()

	srv := sampleServer(t)
	if err := be.Save(context.Background(), wrapState(srv), ""); err != nil {
		t.Fatalf("Save: %v", err)
	}

	if _, err := os.Stat(path); err != nil {
		t.Fatalf("file not created: %v", err)
	}

	state, version, err := be.Load(context.Background())
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if version == "" {
		t.Fatal("expected non-empty version")
	}
	if got, want := len(state.Networks), 1; got != want {
		t.Fatalf("Networks length: got %d, want %d", got, want)
	}
	loaded := state.Networks[srv.InterfaceName].Server
	if loaded.ExternalHostname != srv.ExternalHostname {
		t.Fatalf("ExternalHostname: %q vs %q", loaded.ExternalHostname, srv.ExternalHostname)
	}
	if loaded.ListenPort != srv.ListenPort {
		t.Fatalf("ListenPort: %d vs %d", loaded.ListenPort, srv.ListenPort)
	}
	if loaded.PrivateKey.Key != srv.PrivateKey.Key {
		t.Fatal("PrivateKey mismatch")
	}
	if len(loaded.Peers) != 1 {
		t.Fatalf("Peers length: %d", len(loaded.Peers))
	}
	if loaded.Peers[0].PublicKey.Key != srv.Peers[0].PublicKey.Key {
		t.Fatal("peer PublicKey mismatch")
	}
	if loaded.Peers[0].PresharedKey.Key != srv.Peers[0].PresharedKey.Key {
		t.Fatal("peer PresharedKey mismatch")
	}
}

func TestBackendSaveFilePermissions(t *testing.T) {
	be, path := newTestBackend(t)
	defer be.Close()

	if err := be.Save(context.Background(), wrapState(sampleServer(t)), ""); err != nil {
		t.Fatalf("Save: %v", err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0o600 {
		t.Fatalf("permissions: got %o, want 0600", perm)
	}
}

func TestBackendSaveProducesValidJSON(t *testing.T) {
	be, path := newTestBackend(t)
	defer be.Close()

	if err := be.Save(context.Background(), wrapState(sampleServer(t)), ""); err != nil {
		t.Fatalf("Save: %v", err)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if !json.Valid(raw) {
		t.Fatalf("output is not valid JSON: %s", raw)
	}
	if raw[len(raw)-1] != '\n' {
		t.Fatal("output should end with newline")
	}
}

func TestBackendLoadMissingFile(t *testing.T) {
	dir := t.TempDir()
	be, err := Open(&url.URL{Scheme: "jsonfile", Path: filepath.Join(dir, "no-such-file.json")})
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer be.Close()
	if _, _, err := be.Load(context.Background()); err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestBackendLoadInvalidJSON(t *testing.T) {
	be, path := newTestBackend(t)
	defer be.Close()
	if err := os.WriteFile(path, []byte("not json"), 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	if _, _, err := be.Load(context.Background()); err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestBackendLoadMissingEndpoint(t *testing.T) {
	be, _ := newTestBackend(t)
	defer be.Close()
	srv := sampleServer(t)
	srv.ExternalHostname = ""
	srv.ExternalIP = nil
	srv.ExternalIP6 = nil
	if err := be.Save(context.Background(), wrapState(srv), ""); err != nil {
		t.Fatalf("Save: %v", err)
	}
	if _, _, err := be.Load(context.Background()); err == nil {
		t.Fatal("expected error for missing endpoint")
	}
}

func TestBackendSaveRejectsMultiNetwork(t *testing.T) {
	be, _ := newTestBackend(t)
	defer be.Close()
	state := &store.State{Networks: map[string]*store.Network{
		"dsnet":  {Server: sampleServer(t)},
		"dsnet2": {Server: sampleServer(t)},
	}}
	if err := be.Save(context.Background(), state, ""); err == nil {
		t.Fatal("expected error for multi-network state")
	}
}

func TestBackendVersionStaleSave(t *testing.T) {
	be, _ := newTestBackend(t)
	defer be.Close()

	if err := be.Save(context.Background(), wrapState(sampleServer(t)), ""); err != nil {
		t.Fatalf("initial Save: %v", err)
	}
	_, version1, err := be.Load(context.Background())
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	srv := sampleServer(t)
	srv.ListenPort = 12345
	if err := be.Save(context.Background(), wrapState(srv), version1); err != nil {
		t.Fatalf("Save with correct version: %v", err)
	}

	// Now the on-disk version has changed; saving again with the stale
	// version must fail.
	srv.ListenPort = 23456
	if err := be.Save(context.Background(), wrapState(srv), version1); err == nil {
		t.Fatal("expected stale-version Save to fail")
	}
}

func TestBackendVersionFreshSaveOK(t *testing.T) {
	be, _ := newTestBackend(t)
	defer be.Close()
	// Empty expected version skips the check — used by init.
	if err := be.Save(context.Background(), wrapState(sampleServer(t)), ""); err != nil {
		t.Fatalf("Save: %v", err)
	}
}

func TestBackendRoundTripByteIdentical(t *testing.T) {
	be, path := newTestBackend(t)
	defer be.Close()

	srv := sampleServer(t)
	if err := be.Save(context.Background(), wrapState(srv), ""); err != nil {
		t.Fatalf("Save: %v", err)
	}
	golden, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	state, version, err := be.Load(context.Background())
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if err := be.Save(context.Background(), state, version); err != nil {
		t.Fatalf("re-Save: %v", err)
	}
	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile after: %v", err)
	}
	if string(golden) != string(after) {
		t.Fatalf("round-trip not byte-identical\nbefore:\n%s\nafter:\n%s", golden, after)
	}
}

func TestRegistryDispatch(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "dsnetconfig.json")
	be, err := store.Open("jsonfile://" + path)
	if err != nil {
		t.Fatalf("store.Open: %v", err)
	}
	defer be.Close()
	if err := be.Save(context.Background(), wrapState(sampleServer(t)), ""); err != nil {
		t.Fatalf("Save: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected file at %s: %v", path, err)
	}
}

func TestRegistryUnknownScheme(t *testing.T) {
	_, err := store.Open("postgres://example/db")
	if err == nil {
		t.Fatal("expected error for unknown scheme")
	}
}
