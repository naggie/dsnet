package dsnet

func Add(hostname string, owner string, description string) {//, publicKey string) {
	conf := MustLoadDsnetConfig()

	privateKey := GenerateJSONPrivateKey()
	presharedKey := GenerateJSONKey()
	publicKey := privateKey.PublicKey()

	IP, err := conf.ChooseIP()

	check(err)

	peer := PeerConfig{
		Owner: owner,
		Hostname: hostname,
		Description: description,
		PublicKey: publicKey,
		PresharedKey: presharedKey,
		// TODO Endpoint:
		// TODO pick an available IP AllowedIPs
	}

	conf.MustAddPeer(peer)
	conf.MustSave()
}
