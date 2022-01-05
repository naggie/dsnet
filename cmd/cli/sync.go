package cli

func Sync() {
	// TODO check device settings first
	conf, err := LoadConfigFile()
	check(err, "failed to load configuration file")
	server := GetServer(conf)
	err = server.ConfigureDevice()
	check(err, "failed to sync device configuration")
}
