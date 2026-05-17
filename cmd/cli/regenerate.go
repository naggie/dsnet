package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/naggie/dsnet/lib"
	"github.com/spf13/viper"
)

func Regenerate(hostname string, confirm bool) error {
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

	if !confirm {
		if err := ConfirmOrAbort("This will invalidate current configuration. Regenerate config for %s?", hostname); err != nil {
			return err
		}
	}

	found := false
	for _, peer := range server.Peers {
		if peer.Hostname != hostname {
			continue
		}
		privateKey, err := lib.GenerateJSONPrivateKey()
		if err != nil {
			return fmt.Errorf("%w - failed to generate private key", err)
		}
		preshareKey, err := lib.GenerateJSONKey()
		if err != nil {
			return fmt.Errorf("%w - failed to generate preshared key", err)
		}

		peer.PrivateKey = privateKey
		peer.PublicKey = privateKey.PublicKey()
		peer.PresharedKey = preshareKey

		if err := server.RemovePeer(hostname); err != nil {
			return fmt.Errorf("%w - failed to regenerate peer", err)
		}

		peerType := viper.GetString("output")
		peerConfigBytes, err := lib.AsciiPeerConfig(peer, peerType, *server)
		if err != nil {
			return fmt.Errorf("%w - failed to get peer configuration", err)
		}
		os.Stdout.Write(peerConfigBytes.Bytes())

		if err := server.AddPeer(peer); err != nil {
			return fmt.Errorf("%w - failure to add peer", err)
		}
		found = true
		break
	}

	if !found {
		return fmt.Errorf("unknown hostname: %s", hostname)
	}

	if err := backend.Save(context.Background(), state, version); err != nil {
		return fmt.Errorf("%w - failed to save state", err)
	}
	if err := server.ConfigureDevice(); err != nil {
		return fmt.Errorf("%w - failed to configure device", err)
	}
	return nil
}
