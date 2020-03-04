package dsnet

func Sync() {
	// TODO check device settings first
	conf := MustLoadDsnetConfig()
	ConfigureDevice(conf)
}
