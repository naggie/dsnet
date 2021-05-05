package dsnet

import (
	"fmt"
)

func Regenerate(hostname string, confirm bool) {
	conf := MustLoadDsnetConfig()
	found := false

	if !confirm {
		ConfirmOrAbort("This will invalidate current configuration. Regenerate config for %s?", hostname)
	}

	for _, peer := range conf.Peers {
		if peer.Hostname == hostname {
			privateKey := GenerateJSONPrivateKey()
			peer.PrivateKey = privateKey
			peer.PublicKey = privateKey.PublicKey()
			peer.PresharedKey = GenerateJSONKey()

			conf.MustRemovePeer(hostname)
			PrintPeerCfg(&peer, conf)
			found = true
			conf.MustAddPeer(peer)

			break
		}
	}

	if !found {
		ExitFail(fmt.Sprintf("unknown hostname: %s", hostname))
	}

	conf.MustSave()
	MustConfigureDevice(conf)
}
