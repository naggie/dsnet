package cli

import (
	"fmt"
	"os"

	"github.com/naggie/dsnet/lib"
	"github.com/spf13/viper"
)

func Regenerate(hostname string, confirm bool) error {
	config, err := LoadConfigFile()
	if err != nil {
		return fmt.Errorf("%w - failure to load config file", err)
	}
	server := GetServer(config)

	found := false

	if !confirm {
		ConfirmOrAbort("This will invalidate current configuration. Regenerate config for %s?", hostname)
	}

	for _, peer := range server.Peers {
		if peer.Hostname == hostname {
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

			err = config.RemovePeer(hostname)
			if err != nil {
				return fmt.Errorf("%w - failed to regenerate peer", err)
			}

			peerType := viper.GetString("output")

			peerConfigBytes, err := lib.AsciiPeerConfig(peer, peerType, *server)
			if err != nil {
				return fmt.Errorf("%w - failed to get peer configuration", err)
			}
			os.Stdout.Write(peerConfigBytes.Bytes())
			found = true
			if err = config.AddPeer(peer); err != nil {
				return fmt.Errorf("%w - failure to add peer", err)
			}

			break
		}
	}

	if !found {
		return fmt.Errorf("unknown hostname: %s", hostname)
	}

	// Get a new server configuration so we can update the wg interface with the new peer details
	server = GetServer(config)
	if err = config.Save(); err != nil {
		return fmt.Errorf("%w - failure saving config", err)
	}
	server.ConfigureDevice()
	return nil
}
