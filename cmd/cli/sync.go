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
	return nil
}
