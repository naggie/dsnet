package dsnet

import "fmt"

func Remove(hostname string, confirm bool) {
	conf := MustLoadDsnetConfig()
	conf.MustRemovePeer(hostname)
	if !confirm {
		ConfirmOrAbort("Do you really want to remove %s?", hostname)
	}
	conf.MustSave()
	ConfigureDevice(conf)
	if confirm {
		fmt.Printf("Removed %s\n", hostname)
	}
}
