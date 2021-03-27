package dsnet

import (
	"fmt"
)

func Regenerate(hostname string) {
	conf := MustLoadDsnetConfig()
	found := false

	privateKey := GenerateJSONPrivateKey()

	for _, peer := range conf.Peers {
		if peer.Hostname == hostname {
			peer.PrivateKey = privateKey
			peer.PublicKey = privateKey.PublicKey()
			peer.PresharedKey = GenerateJSONKey()
			PrintPeerCfg(&peer, conf)
			found = true
		}
	}

	if !found {
		ExitFail(fmt.Sprintf("unknown hostname: %s", hostname))
	}

	conf.MustSave()
	MustConfigureDevice(conf)
}
