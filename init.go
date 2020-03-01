package dsnet

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net"
	"time"
	"io/ioutil"
	//"github.com/mikioh/ipaddr"
)

func Init() {
	privateKey := GenerateJSONPrivateKey()
	presharedKey := GenerateJSONKey()

	conf := DsnetConfig{
		PrivateKey:   privateKey,
		PresharedKey: presharedKey,
		ListenPort:   DEFAULT_LISTEN_PORT,
		Network:      getRandomNetwork(),
		Peers:        make([]PeerConfig, 0),
		Domain:       "dsnet",
		ReportFile:   DEFAULT_REPORT_FILE,
	}

	_json, _ := json.MarshalIndent(conf, "", "    ")
	err := ioutil.WriteFile(CONFIG_FILE, _json, 0600)
    check(err)

	fmt.Printf("Config written to %s. Please edit.", CONFIG_FILE)
}

// get a random /22 subnet on 10.0.0.0 (1023 hosts) (or /24?)
func getRandomNetwork() JSONIPNet {
	rbs := make([]byte, 2)
	rand.Seed(time.Now().UTC().UnixNano())
	rand.Read(rbs)

	return JSONIPNet{
		IPNet: net.IPNet{
			net.IP{10, rbs[0], rbs[1] << 2, 0},
			net.IPMask{255, 255, 252, 0},
		},
	}
}
