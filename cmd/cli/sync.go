package cli

func Sync() error {
	// TODO check device settings first
	conf, err := LoadConfigFile()
	if err != nil {
		return wrapError(err, "failed to load configuration file")
	}
	server := GetServer(conf)
	err = server.ConfigureDevice()
	if err != nil {
		return wrapError(err, "failed to sync device configuration")
	}
	return nil
}
