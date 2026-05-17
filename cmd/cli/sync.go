package cli

import (
	"context"
	"fmt"
)

func Sync() error {
	backend, err := OpenStore()
	if err != nil {
		return fmt.Errorf("%w - failed to open storage backend", err)
	}
	defer backend.Close()

	state, _, err := backend.Load(context.Background())
	if err != nil {
		return fmt.Errorf("%w - failed to load state", err)
	}
	network, err := DefaultNetwork(state)
	if err != nil {
		return err
	}
	server := network.Server

	if err := server.ConfigureDevice(); err != nil {
		return fmt.Errorf("%w - failed to sync device configuration", err)
	}

	if err := server.Up(); err != nil {
		return fmt.Errorf("%w - failed to bring up the interface", err)
	}
	return nil
}
