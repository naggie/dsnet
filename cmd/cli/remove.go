package cli

import "fmt"

func Remove(hostname string, confirm bool) error {
	conf := MustLoadConfigFile()

	err := conf.RemovePeer(hostname)
	if err != nil {
		return wrapError(err, "failed to update config")
	}

	if !confirm {
		ConfirmOrAbort("Do you really want to remove %s?", hostname)
	}

	conf.MustSave()
	server := GetServer(conf)

	err = server.ConfigureDevice()
	if err != nil {
		return wrapError(err, fmt.Sprintf("failed to sync server config to wg interface: %s", server.InterfaceName))
	}
	return nil
}
