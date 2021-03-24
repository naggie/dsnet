package dsnet

import (
	"fmt"
)

func Config(hostname string) {
	conf := MustLoadDsnetConfig()
	found := false

	for _, peer := range conf.Peers {
		if peer.Hostname == hostname {
			PrintPeerCfg(&peer, conf)
			found = true
		}
	}

	if !found {
		ExitFail(fmt.Sprintf("unknown hostname: %s", hostname))
	}
}
