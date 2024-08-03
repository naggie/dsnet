package cli

import (
	"fmt"
	"os"

	"github.com/naggie/dsnet/lib"
	"github.com/spf13/viper"
)

// Add prompts for the required information and creates a new peer
func Add(hostname string, privKey, pubKey bool, owner, description string, confirm bool) {
	config, err := LoadConfigFile()
	if err != nil {
		return fmt.Errorf("%w - failed to load configuration file", err)
	}
	server := GetServer(config)

	var private, public string
	if privKey {
		private = MustPromptString("private key", true)
	}
	if pubKey {
		public = MustPromptString("public key", true)
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

	// publicKey := MustPromptString("PublicKey (optional)", false)
	if !confirm {
		ConfirmOrAbort("\nDo you want to add the above configuration?")
	}

	// newline (not on stdout) to separate config
	fmt.Fprintln(os.Stderr)

	peer, err := lib.NewPeer(server, private, public, owner, hostname, description)
	if err != nil {
		return fmt.Errorf("%w - failed to get new peer", err)
	}

	// TODO Some kind of recovery here would be nice, to avoid
	// leaving things in a potential broken state

	if err = config.AddPeer(peer); err != nil {
		return fmt.Errorf("%w - failed to add new peer", err)
	}

	peerType := viper.GetString("output")

	peerConfigBytes, err := lib.AsciiPeerConfig(peer, peerType, *server)
	if err != nil {
		return fmt.Errorf("%w - failed to get peer configuration", err)
	}
	os.Stdout.Write(peerConfigBytes.Bytes())

	if err = config.Save(); err != nil {
		return fmt.Errorf("%w - failed to save config file", err)
	}

	server = GetServer(config)
	if err = server.ConfigureDevice(); err != nil {
		return fmt.Errorf("%w - failed to configure device", err)
	}
	return nil
}
