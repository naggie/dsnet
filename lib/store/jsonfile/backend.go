// Package jsonfile implements the legacy single-file JSON storage backend.
// Registered under the "jsonfile" URL scheme; the URL path is the file
// location, e.g. jsonfile:///etc/dsnetconfig.json.
package jsonfile

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"

	"github.com/go-playground/validator/v10"
	"github.com/naggie/dsnet/lib"
	"github.com/naggie/dsnet/lib/store"
	"github.com/spf13/viper"
)

func init() {
	store.Register("jsonfile", Open)
}

// Backend persists a single-network dsnet State as a JSON document on disk.
type Backend struct {
	path string
}

// Open returns a Backend bound to the path in u (u.Path).
func Open(u *url.URL) (store.Backend, error) {
	if u == nil {
		return nil, errors.New("jsonfile: nil URL")
	}
	if u.Path == "" {
		return nil, errors.New("jsonfile: storage URL is missing a path")
	}
	return &Backend{path: u.Path}, nil
}

// Close releases any held resources. The basic Backend holds none.
func (b *Backend) Close() error { return nil }

// Load reads the JSON file, validates it, and returns the resulting State
// together with a Version (the SHA-256 of the file contents). Returns a
// helpful error if the file is missing or unreadable.
func (b *Backend) Load(_ context.Context) (*store.State, store.Version, error) {
	raw, err := os.ReadFile(b.path)
	if os.IsNotExist(err) {
		return nil, "", fmt.Errorf("%s does not exist. `dsnet init` may be required", b.path)
	} else if os.IsPermission(err) {
		return nil, "", fmt.Errorf("%s cannot be accessed. Sudo may be required", b.path)
	} else if err != nil {
		return nil, "", fmt.Errorf("%w - failed to read %s", err, b.path)
	}

	conf := dsnetConfig{
		PersistentKeepalive: 25,
		MTU:                 1420,
	}
	if err := json.Unmarshal(raw, &conf); err != nil {
		return nil, "", fmt.Errorf("%w - failed to parse %s", err, b.path)
	}
	if err := validator.New().Struct(conf); err != nil {
		return nil, "", fmt.Errorf("%w - %s failed validation", err, b.path)
	}
	if conf.ExternalHostname == "" && len(conf.ExternalIP) == 0 && len(conf.ExternalIP6) == 0 {
		return nil, "", fmt.Errorf("config does not contain ExternalIP, ExternalIP6 or ExternalHostname")
	}

	server := conf.toServer(viper.GetString("fallback_wg_bin"))
	state := &store.State{
		Networks: map[string]*store.Network{
			server.InterfaceName: {Server: server},
		},
	}
	return state, hashVersion(raw), nil
}

// Save writes the State to disk. If expected is non-empty, Save first
// re-hashes the on-disk file and returns an error if it differs (another
// writer changed the file under us).
func (b *Backend) Save(_ context.Context, state *store.State, expected store.Version) error {
	if state == nil {
		return errors.New("jsonfile: nil state")
	}
	if len(state.Networks) != 1 {
		return fmt.Errorf("jsonfile: multi-network state not supported (got %d networks)", len(state.Networks))
	}

	var server = singleServer(state)
	if server == nil {
		return errors.New("jsonfile: state has no server")
	}

	if expected != "" {
		current, err := b.currentVersion()
		if err != nil {
			return err
		}
		if current != expected {
			return fmt.Errorf("config changed on disk: expected version %s, found %s", expected, current)
		}
	}

	conf := fromServer(server)
	out, err := json.MarshalIndent(conf, "", "    ")
	if err != nil {
		return fmt.Errorf("%w - failed to marshal config", err)
	}
	out = append(out, '\n')

	dir := filepath.Dir(b.path)
	tmp, err := os.CreateTemp(dir, ".dsnetconfig.*.tmp")
	if err != nil {
		return fmt.Errorf("%w - failed to create temp file in %s", err, dir)
	}
	tmpName := tmp.Name()
	cleanup := func() { _ = os.Remove(tmpName) }
	if _, err := tmp.Write(out); err != nil {
		_ = tmp.Close()
		cleanup()
		return fmt.Errorf("%w - failed to write temp file %s", err, tmpName)
	}
	if err := tmp.Chmod(0o600); err != nil {
		_ = tmp.Close()
		cleanup()
		return fmt.Errorf("%w - failed to chmod temp file %s", err, tmpName)
	}
	if err := tmp.Close(); err != nil {
		cleanup()
		return fmt.Errorf("%w - failed to close temp file %s", err, tmpName)
	}
	if err := os.Rename(tmpName, b.path); err != nil {
		cleanup()
		return fmt.Errorf("%w - failed to rename %s to %s", err, tmpName, b.path)
	}
	return nil
}

func singleServer(state *store.State) *lib.Server {
	for _, n := range state.Networks {
		if n != nil {
			return n.Server
		}
	}
	return nil
}

func (b *Backend) currentVersion() (store.Version, error) {
	raw, err := os.ReadFile(b.path)
	if os.IsNotExist(err) {
		return "", nil
	} else if err != nil {
		return "", fmt.Errorf("%w - failed to read %s", err, b.path)
	}
	return hashVersion(raw), nil
}

func hashVersion(raw []byte) store.Version {
	sum := sha256.Sum256(raw)
	return store.Version(hex.EncodeToString(sum[:]))
}
