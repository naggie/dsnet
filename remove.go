package dsnet

import (
	"os"
)

func Remove() {
	if len(os.Args) != 3 {
		// TODO non-red
		ExitFail("Hostname argument required: dsnet remove <hostname>")
	}

	conf := MustLoadDsnetConfig()
	hostname := os.Args[2]
	conf.MustRemovePeer(hostname)
	ConfirmOrAbort("Do you really want to remove %s?", os.Args[2])
	conf.MustSave()
	ConfigureDevice(conf)
}
