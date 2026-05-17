// Package store defines the persistence abstraction for dsnet state.
//
// Backends are registered by URL scheme; callers obtain one via Open and
// load/save *State values through it. The State type carries every network
// dsnet manages keyed by WireGuard interface name; in Branch A every CLI
// invocation operates on the single default network.
package store

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"sync"

	"github.com/naggie/dsnet/lib"
)

// Version is an opaque token identifying a particular snapshot of state.
// Backends fill it however they like — jsonfile uses the SHA-256 of the
// file contents. A zero Version disables optimistic concurrency checks
// (used by init when writing a fresh file).
type Version string

// State holds every network dsnet manages, keyed by interface name.
type State struct {
	Networks map[string]*Network
}

// Network is the per-interface dsnet state. Kept minimal in Branch A; Branch
// B is free to add metadata as needed.
type Network struct {
	Server *lib.Server
}

// Backend persists and loads dsnet state.
type Backend interface {
	Load(ctx context.Context) (*State, Version, error)
	Save(ctx context.Context, state *State, expected Version) error
	Close() error
}

// OpenFunc constructs a Backend from a parsed URL.
type OpenFunc func(*url.URL) (Backend, error)

var (
	registryMu sync.RWMutex
	registry   = map[string]OpenFunc{}
)

// Register associates an OpenFunc with a URL scheme. Calling Register twice
// for the same scheme panics — duplicate registration is a programmer error.
func Register(scheme string, f OpenFunc) {
	if scheme == "" {
		panic("store.Register: empty scheme")
	}
	if f == nil {
		panic("store.Register: nil OpenFunc")
	}
	registryMu.Lock()
	defer registryMu.Unlock()
	if _, ok := registry[scheme]; ok {
		panic(fmt.Sprintf("store.Register: scheme %q already registered", scheme))
	}
	registry[scheme] = f
}

// Open parses the given URL and dispatches to the registered backend.
func Open(rawURL string) (Backend, error) {
	if rawURL == "" {
		return nil, errors.New("empty storage URL")
	}
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("%w - invalid storage URL %q", err, rawURL)
	}
	if u.Scheme == "" {
		return nil, fmt.Errorf("storage URL %q is missing a scheme", rawURL)
	}
	registryMu.RLock()
	f, ok := registry[u.Scheme]
	registryMu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("unknown storage backend scheme: %s", u.Scheme)
	}
	return f(u)
}
