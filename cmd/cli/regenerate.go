package cli

import (
	"fmt"
	"os"

	"github.com/naggie/dsnet/lib"
	"github.com/spf13/viper"
)

func Regenerate(hostname string, confirm bool) {
	config := MustLoadConfigFile()
	server := GetServer(config)

	found := false

	if !confirm {
		ConfirmOrAbort("This will invalidate current configuration. Regenerate config for %s?", hostname)
	}

	for _, peer := range server.Peers {
		if peer.Hostname == hostname {
			privateKey, err := lib.GenerateJSONPrivateKey()
			check(err, "failed to generate private key")

			preshareKey, err := lib.GenerateJSONKey()
			check(err, "failed to generate preshared key")

			peer.PrivateKey = privateKey
			peer.PublicKey = privateKey.PublicKey()
			peer.PresharedKey = preshareKey

			err = config.RemovePeer(hostname)
			check(err, "failed to regenerate peer")

			peerType := viper.GetString("output")

			peerConfigBytes, err := lib.AsciiPeerConfig(peer, peerType, *server)
			check(err, "failed to get peer configuration")
			os.Stdout.Write(peerConfigBytes.Bytes())
			found = true
			config.MustAddPeer(peer)

			break
		}
	}

	if !found {
		ExitFail(fmt.Sprintf("unknown hostname: %s", hostname))
	}

	// Get a new server configuration so we can update the wg interface with the new peer details
	server = GetServer(config)
	config.MustSave()
	server.ConfigureDevice()
}
