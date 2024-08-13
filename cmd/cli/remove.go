package cli

import "fmt"

func Remove(hostname string, confirm bool) error {
	conf, err := LoadConfigFile()
	if err != nil {
		return fmt.Errorf("%w - failed to load config", err)
	}

	if err = conf.RemovePeer(hostname); err != nil {
		return fmt.Errorf("%w - failed to update config", err)
	}

	if !confirm {
		ConfirmOrAbort("Do you really want to remove %s?", hostname)
	}

	if err = conf.Save(); err != nil {
		return fmt.Errorf("%w - failure to save config", err)
	}
	server := GetServer(conf)

	if err = server.ConfigureDevice(); err != nil {
		return fmt.Errorf("%w - failed to sync server config to wg interface: %s", err, server.InterfaceName)
	}
	return nil
}
