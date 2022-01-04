package cli

import "fmt"

func Sync() {
	// TODO check device settings first
	conf, err := LoadConfigFile()
	check(err, fmt.Sprintf("failed to load configuration file: %s", err))
	server := GetServer(conf)
	err = server.ConfigureDevice()
	check(err, fmt.Sprintf("failed to sync device configuration: %s", err))
}
