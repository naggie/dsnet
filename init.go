package dsnet

import (
	"fmt"
	"math/rand"
	"net"
	"strings"
	"time"
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
		ExternalIP:   getExternalIP(),
	}

	IP := conf.MustAllocateIP()
	conf.InternalIP = IP
	conf.InternalDNS = IP

	conf.MustSave()

	fmt.Printf("Config written to %s. Please check/edit.", CONFIG_FILE)
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

// TODO support IPv6
func getExternalIP() net.IP {
	conn, _ := net.Dial("udp", "8.8.8.8:80")
	defer conn.Close()

	localAddr := conn.LocalAddr().String()
	IP := net.ParseIP(strings.Split(localAddr, ":")[0])

	if !(IP[0] == 10 || (IP[0] == 172 && IP[1] >= 16 && IP[1] <= 31) || (IP[0] == 192 && IP[1] == 168)) {
		// not private, so public
		return IP
	}
	// TODO detect private IP and use icanhazip.com instead
	return net.IP{}
}
