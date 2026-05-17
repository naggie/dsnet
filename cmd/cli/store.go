package cli

import (
	"fmt"
	"os"

	"github.com/naggie/dsnet/lib/store"
	"github.com/spf13/viper"
)

// OpenStore opens the configured storage backend. DSNET_STORE (URL form)
// is the canonical setting; if it is unset but the legacy DSNET_CONFIG_FILE
// env var is present, a jsonfile URL is built from it and a one-line
// deprecation notice is emitted on stderr.
func OpenStore() (store.Backend, error) {
	_, storeSet := os.LookupEnv("DSNET_STORE")
	if !storeSet {
		if legacy, ok := os.LookupEnv("DSNET_CONFIG_FILE"); ok {
			fmt.Fprintf(os.Stderr, "DSNET_CONFIG_FILE is deprecated; use DSNET_STORE=jsonfile://%s\n", legacy)
			return store.Open("jsonfile://" + legacy)
		}
	}
	return store.Open(viper.GetString("store"))
}

// DefaultNetwork returns the single network from state. In Branch A every CLI
// invocation operates on the default network; multi-network is supported at
// the type level only and surfaces as an error here until Branch B adds the
// --network flag.
func DefaultNetwork(state *store.State) (*store.Network, error) {
	if state == nil || len(state.Networks) == 0 {
		return nil, fmt.Errorf("state has no networks")
	}
	if len(state.Networks) > 1 {
		return nil, fmt.Errorf("multiple networks present (%d); --network selection not supported in this release", len(state.Networks))
	}
	for _, n := range state.Networks {
		return n, nil
	}
	return nil, fmt.Errorf("state has no networks")
}
