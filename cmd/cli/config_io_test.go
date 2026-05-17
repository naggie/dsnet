package cli

import (
	"encoding/json"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/naggie/dsnet/lib"
	"github.com/spf13/viper"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

func setupViperForTest(t *testing.T, configPath string) {
	t.Helper()
	old := viper.GetString("config_file")
	viper.Set("config_file", configPath)
	t.Cleanup(func() { viper.Set("config_file", old) })
}

func writeTestConfig(t *testing.T, path string, conf *DsnetConfig) {
	t.Helper()
	b, err := json.MarshalIndent(conf, "", "    ")
	if err != nil {
		t.Fatalf("failed to marshal config: %v", err)
	}
	b = append(b, '\n')
	if err := os.WriteFile(path, b, 0o600); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}
}

func TestSaveAndLoadRoundTrip(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "dsnetconfig.json")
	setupViperForTest(t, configPath)

	privKey, err := wgtypes.GeneratePrivateKey()
	if err != nil {
		t.Fatalf("failed to generate private key: %v", err)
	}
	peerKey, err := wgtypes.GeneratePrivateKey()
	if err != nil {
		t.Fatalf("failed to generate peer key: %v", err)
	}
	psk, err := wgtypes.GenerateKey()
	if err != nil {
		t.Fatalf("failed to generate preshared key: %v", err)
	}

	original := &DsnetConfig{
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
		Networks:            []lib.JSONIPNet{},
		PostUp:              "iptables -A FORWARD -i dsnet -j ACCEPT",
		PostDown:            "iptables -D FORWARD -i dsnet -j ACCEPT",
		PersistentKeepalive: 25,
		Peers: []PeerConfig{
			{
				Hostname:     "laptop",
				Owner:        "alice",
				Description:  "Alice's laptop",
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

	// Save
	err = original.Save()
	if err != nil {
		t.Fatalf("Save error: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatal("config file was not created")
	}

	// Load
	loaded, err := LoadConfigFile()
	if err != nil {
		t.Fatalf("LoadConfigFile error: %v", err)
	}

	// Compare fields
	if loaded.ExternalHostname != original.ExternalHostname {
		t.Fatalf("ExternalHostname mismatch: %s != %s", loaded.ExternalHostname, original.ExternalHostname)
	}
	if loaded.ListenPort != original.ListenPort {
		t.Fatalf("ListenPort mismatch: %d != %d", loaded.ListenPort, original.ListenPort)
	}
	if loaded.Domain != original.Domain {
		t.Fatalf("Domain mismatch: %s != %s", loaded.Domain, original.Domain)
	}
	if loaded.InterfaceName != original.InterfaceName {
		t.Fatalf("InterfaceName mismatch: %s != %s", loaded.InterfaceName, original.InterfaceName)
	}
	if loaded.Network.IPNet.String() != original.Network.IPNet.String() {
		t.Fatalf("Network mismatch: %s != %s", loaded.Network.IPNet.String(), original.Network.IPNet.String())
	}
	if loaded.Network6.IPNet.String() != original.Network6.IPNet.String() {
		t.Fatalf("Network6 mismatch")
	}
	if !loaded.IP.Equal(original.IP) {
		t.Fatalf("IP mismatch: %s != %s", loaded.IP, original.IP)
	}
	if loaded.PrivateKey.Key != original.PrivateKey.Key {
		t.Fatal("PrivateKey mismatch")
	}
	if loaded.PostUp != original.PostUp {
		t.Fatalf("PostUp mismatch: %s != %s", loaded.PostUp, original.PostUp)
	}
	if loaded.PostDown != original.PostDown {
		t.Fatalf("PostDown mismatch: %s != %s", loaded.PostDown, original.PostDown)
	}
	if loaded.PersistentKeepalive != original.PersistentKeepalive {
		t.Fatalf("PersistentKeepalive mismatch: %d != %d", loaded.PersistentKeepalive, original.PersistentKeepalive)
	}

	// Check peer
	if len(loaded.Peers) != 1 {
		t.Fatalf("expected 1 peer, got %d", len(loaded.Peers))
	}
	p := loaded.Peers[0]
	if p.Hostname != "laptop" {
		t.Fatalf("peer hostname mismatch: %s", p.Hostname)
	}
	if p.Owner != "alice" {
		t.Fatalf("peer owner mismatch: %s", p.Owner)
	}
	if p.PublicKey.Key != peerKey.PublicKey() {
		t.Fatal("peer public key mismatch")
	}
	if p.PresharedKey.Key != psk {
		t.Fatal("peer preshared key mismatch")
	}
}

func TestSaveFilePermissions(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "dsnetconfig.json")
	setupViperForTest(t, configPath)

	conf := testDsnetConfig(t)
	if err := conf.Save(); err != nil {
		t.Fatalf("Save error: %v", err)
	}

	info, err := os.Stat(configPath)
	if err != nil {
		t.Fatalf("stat error: %v", err)
	}

	// Should be 0600 (owner read/write only)
	perm := info.Mode().Perm()
	if perm != 0o600 {
		t.Fatalf("expected permissions 0600, got %o", perm)
	}
}

func TestLoadConfigFileNotExist(t *testing.T) {
	setupViperForTest(t, "/tmp/nonexistent-dsnet-test-config.json")

	_, err := LoadConfigFile()
	if err == nil {
		t.Fatal("expected error for nonexistent config")
	}
}

func TestLoadConfigFileInvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "dsnetconfig.json")
	setupViperForTest(t, configPath)

	if err := os.WriteFile(configPath, []byte("not json"), 0o600); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	_, err := LoadConfigFile()
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestLoadConfigFileDefaultKeepalive(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "dsnetconfig.json")
	setupViperForTest(t, configPath)

	// Write config without PersistentKeepalive
	privKey, err := wgtypes.GeneratePrivateKey()
	if err != nil {
		t.Fatalf("failed to generate key: %v", err)
	}
	raw := []byte(`{
		"ExternalHostname": "vpn.example.com",
		"ListenPort":       51820,
		"Domain":           "dsnet",
		"InterfaceName":    "dsnet",
		"Network":          "10.0.0.0/22",
		"Network6":         "fd00::/64",
		"IP":               "10.0.0.1",
		"IP6":              "fd00::1",
		"PrivateKey":       "` + privKey.String() + `",
		"Networks":         [],
		"Peers":            []
	}`)
	if !json.Valid(raw) {
		t.Fatalf("test fixture is not valid JSON: %s", raw)
	}
	if err := os.WriteFile(configPath, raw, 0o600); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	loaded, err := LoadConfigFile()
	if err != nil {
		t.Fatalf("LoadConfigFile error: %v", err)
	}

	// Default PersistentKeepalive should be 25
	if loaded.PersistentKeepalive != 25 {
		t.Fatalf("expected default PersistentKeepalive 25, got %d", loaded.PersistentKeepalive)
	}
}

func TestLoadConfigFileMissingEndpoint(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "dsnetconfig.json")
	setupViperForTest(t, configPath)

	privKey, err := wgtypes.GeneratePrivateKey()
	if err != nil {
		t.Fatalf("failed to generate key: %v", err)
	}
	conf := &DsnetConfig{
		// No ExternalHostname, ExternalIP, or ExternalIP6
		ListenPort:    51820,
		Domain:        "dsnet",
		InterfaceName: "dsnet",
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
		Networks:            []lib.JSONIPNet{},
		Peers:               []PeerConfig{},
		PersistentKeepalive: 25,
	}

	writeTestConfig(t, configPath, conf)

	_, err = LoadConfigFile()
	if err == nil {
		t.Fatal("expected error when no endpoint is configured")
	}
}

func TestSaveCreatesValidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "dsnetconfig.json")
	setupViperForTest(t, configPath)

	conf := testDsnetConfig(t)
	if err := conf.Save(); err != nil {
		t.Fatalf("Save error: %v", err)
	}

	// Read raw file and verify it's valid JSON
	raw, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read error: %v", err)
	}

	if !json.Valid(raw) {
		t.Fatalf("saved file is not valid JSON: %s", raw)
	}

	// Should end with newline
	if raw[len(raw)-1] != '\n' {
		t.Fatal("config file should end with newline")
	}
}

func TestSaveOverwritesExisting(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "dsnetconfig.json")
	setupViperForTest(t, configPath)

	conf := testDsnetConfig(t)
	if err := conf.Save(); err != nil {
		t.Fatalf("first Save error: %v", err)
	}

	// Add a peer and save again
	peer := testLibPeer(t, "laptop", "alice", net.IP{10, 0, 0, 2})
	conf.AddPeer(peer)
	if err := conf.Save(); err != nil {
		t.Fatalf("second Save error: %v", err)
	}

	// Load and verify peer is there
	loaded, err := LoadConfigFile()
	if err != nil {
		t.Fatalf("LoadConfigFile error: %v", err)
	}
	if len(loaded.Peers) != 1 {
		t.Fatalf("expected 1 peer after overwrite, got %d", len(loaded.Peers))
	}
}
