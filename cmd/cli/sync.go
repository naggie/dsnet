package cli

import "fmt"

func Sync() error {
	// TODO check device settings first
	conf, err := LoadConfigFile()
	if err != nil {
		return fmt.Errorf("%w - failed to load configuration file", err)
	}
	server := GetServer(conf)

	err = server.ConfigureDevice()
	if err != nil {
		return fmt.Errorf("%w - failed to sync device configuration", err)
	}

	err = server.Up() // set IPs, interface must be up by this point
	if err != nil {
		return fmt.Errorf("%w - failed to bring up the interface", err)
	}
	return nil
}
