package cli

import (
	"context"
	"fmt"
)

func Remove(hostname string, confirm bool) error {
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

	if err := server.RemovePeer(hostname); err != nil {
		return fmt.Errorf("%w - failed to update state", err)
	}

	if !confirm {
		if err := ConfirmOrAbort("Do you really want to remove %s?", hostname); err != nil {
			return err
		}
	}

	if err := backend.Save(context.Background(), state, version); err != nil {
		return fmt.Errorf("%w - failed to save state", err)
	}

	if err := server.ConfigureDevice(); err != nil {
		return fmt.Errorf("%w - failed to sync server config to wg interface: %s", err, server.InterfaceName)
	}
	return nil
}
