package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/naggie/dsnet/lib"
	"github.com/spf13/viper"
)

// Add prompts for the required information and creates a new peer
func Add(hostname string, privKey, pubKey bool, owner, description string, confirm bool) error {
	backend, err := OpenStore()
	if err != nil {
		return fmt.Errorf("%w - failed to open storage backend", err)
	}
	defer backend.Close()

	state, version, err := backend.Load(context.Background())
	if err != nil {
		return fmt.Errorf("%w - failed to load state", err)
	}
	network, err := DefaultNetwork(state)
	if err != nil {
		return err
	}
	server := network.Server

	var private, public string
	if privKey {
		if private, err = PromptString("private key", true); err != nil {
			return err
		}
	}
	if pubKey {
		if public, err = PromptString("public key", true); err != nil {
			return err
		}
	}
	if owner == "" {
		owner, err = PromptString("owner", true)
		if err != nil {
			return fmt.Errorf("%w - invalid input for owner", err)
		}
	}
	if description == "" {
		description, err = PromptString("Description", true)
		if err != nil {
			return fmt.Errorf("%w - invalid input for Description", err)
		}
	}

	if !confirm {
		if err := ConfirmOrAbort("\nDo you want to add the above configuration?"); err != nil {
			return err
		}
	}

	// newline (not on stdout) to separate config
	fmt.Fprintln(os.Stderr)

	peer, err := lib.NewPeer(server, private, public, owner, hostname, description)
	if err != nil {
		return fmt.Errorf("%w - failed to get new peer", err)
	}

	if err = server.AddPeer(peer); err != nil {
		return fmt.Errorf("%w - failed to add new peer", err)
	}

	peerType := viper.GetString("output")

	peerConfigBytes, err := lib.AsciiPeerConfig(peer, peerType, *server)
	if err != nil {
		return fmt.Errorf("%w - failed to get peer configuration", err)
	}
	os.Stdout.Write(peerConfigBytes.Bytes())

	if err := backend.Save(context.Background(), state, version); err != nil {
		return fmt.Errorf("%w - failed to save state", err)
	}

	if err := server.ConfigureDevice(); err != nil {
		return fmt.Errorf("%w - failed to configure device", err)
	}
	return nil
}
