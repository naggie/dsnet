package dsnet

import (
	"os"
)

func Remove() {
	if len(os.Args) <= 2 {
		// TODO non-red
		ExitFail("Hostname argument required: dsnet remove <hostname>")
	}

	ConfirmOrAbort("Do you really want to remove %s?", os.Args[2])

	conf := MustLoadDsnetConfig()
	hostname := os.Args[2]
	conf.MustRemovePeer(hostname)
	conf.MustSave()
	ConfigureDevice(conf)
}
