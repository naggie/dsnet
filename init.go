package dsnet

import (
	"fmt"
	"math/rand"
	"net"
	"time"
)

func Init() {
	privateKey := GenerateJSONPrivateKey()
	presharedKey := GenerateJSONKey()

	conf := DsnetConfig{
		PrivateKey:   privateKey,
		PresharedKey: presharedKey,
		Port:         DEFAULT_LISTEN_PORT,
		Network:      getRandomNetwork(),
		Peers:        make([]PeerConfig, 0),
		Domain:       "dsnet",
		ReportFile:   DEFAULT_REPORT_FILE,
	}

	IP := conf.MustAllocateIP()
	conf.IP = IP
	conf.DNS = IP

	conf.MustSave()

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
