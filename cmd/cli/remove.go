package cli

import "fmt"

func Remove(hostname string, confirm bool) {
	conf := MustLoadConfigFile()

	err := conf.RemovePeer(hostname)
	check(err, "failed to update config")

	if !confirm {
		ConfirmOrAbort("Do you really want to remove %s?", hostname)
	}

	conf.MustSave()
	server := GetServer(conf)

	err = server.ConfigureDevice()
	check(err, fmt.Sprintf("failed to sync server config to wg interface: %s", server.InterfaceName))
}
