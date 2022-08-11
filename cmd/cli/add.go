package cli

import (
	"fmt"
	"os"

	"github.com/naggie/dsnet/lib"
	"github.com/spf13/viper"
)

// Add prompts for the required information and creates a new peer
func Add(hostname, owner, description string, confirm bool) error {
	// TODO accept existing pubkey
	config, err := LoadConfigFile()
	if err != nil {
		return wrapError(err, "failed to load configuration file")
	}
	server := GetServer(config)

	if owner == "" {
		owner = MustPromptString("owner", true)
	}
	if description == "" {
		description = MustPromptString("Description", true)
	}

	// publicKey := MustPromptString("PublicKey (optional)", false)
	if !confirm {
		ConfirmOrAbort("\nDo you want to add the above configuration?")
	}

	// newline (not on stdout) to separate config
	fmt.Fprintln(os.Stderr)

	peer, err := lib.NewPeer(server, owner, hostname, description)
	if err != nil {
		return wrapError(err, "failed to get new peer")
	}

	// TODO Some kind of recovery here would be nice, to avoid
	// leaving things in a potential broken state

	config.MustAddPeer(peer)

	peerType := viper.GetString("output")

	peerConfigBytes, err := lib.AsciiPeerConfig(peer, peerType, *server)
	if err != nil {
		return wrapError(err, "failed to get peer configuration")
	}
	os.Stdout.Write(peerConfigBytes.Bytes())

	config.MustSave()

	server = GetServer(config)
	err = server.ConfigureDevice()
	if err != nil {
		return wrapError(err, "failed to configure device")
	}
	return nil
}
