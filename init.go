package dsnet

import (
	"net"

	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
	//"github.com/mikioh/ipaddr"
)

func Init() {
	conf := DsnetConfig {
		PrivateKey = wgtypes.GeneratePrivateKey(),
		PresharedKey = wgtypes.GenerateKey(),
		ListenPort = DEFAULT_LISTEN_PORT,
		Network = getRandomNetwork(),
		Domain = "dsnet"
	}
}

// get a random /22 (1023 hosts) (or /24?)
// TODO implement
func getRandomNetwork() net.IPNet {
	return net.IPNet {
		IP{10,129,123,0},
		Mask{255,255,255,240},
	}
}
