package dsnet

func Update() {
	// TODO check device settings first
	conf := MustLoadDsnetConfig()
	ConfigureDevice(conf)
}
