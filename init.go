package dsnet

import (
	"net"
	"math/rand"
	"fmt"
	"time"
	"encoding/json"

	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
	//"github.com/mikioh/ipaddr"
)

func Init() {
	// TODO check errors
	privateKey, _ := wgtypes.GeneratePrivateKey()
	presharedKey, _ := wgtypes.GenerateKey()

	conf := DsnetConfig{
		PrivateKey: privateKey,
		PresharedKey: presharedKey,
		ListenPort: DEFAULT_LISTEN_PORT,
		Network: getRandomNetwork(),
		Peers: make([]PeerConfig,0),
		Domain: "dsnet",
	}

	fmt.Println(conf.Network.String())
	fmt.Printf("%-+v/n",conf)

	_json, _ := json.MarshalIndent(conf, "", "    ")

	fmt.Println(string(_json))
}

// get a random /22 subnet on 10.0.0.0 (1023 hosts) (or /24?)
// TODO also the 20 bit block and 16 bit block?
func getRandomNetwork() net.IPNet {
	rbs := make([]byte, 2)
	rand.Seed(time.Now().UTC().UnixNano())
	rand.Read(rbs)

	return net.IPNet {
		net.IP{10,rbs[0],rbs[1]<<2,0},
		net.IPMask{255,255,252,0},
	}
}
