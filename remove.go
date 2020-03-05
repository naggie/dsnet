package dsnet

import (
	"os"
)

func Remove() {
	if len(os.Args) <= 2 {
		ExitFail("Hostname argument required: dsnet remove <hostname>")
	}
	conf := MustLoadDsnetConfig()
	hostname := os.Args[2]
	conf.MustRemovePeer(hostname)
	conf.MustSave()
	ConfigureDevice(conf)
}
