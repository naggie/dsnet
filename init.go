package dsnet

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

func Init() {
	_, err := os.Stat(CONFIG_FILE)

	if !os.IsNotExist(err) {
		ExitFail("Refusing to overwrite existing %s", CONFIG_FILE)
	}

	conf := DsnetConfig{
		PrivateKey:    GenerateJSONPrivateKey(),
		ListenPort:    DEFAULT_LISTEN_PORT,
		Network:       getRandomNetwork(),
		Peers:         []PeerConfig{},
		Domain:        "dsnet",
		ReportFile:    DEFAULT_REPORT_FILE,
		ExternalIP:    getExternalIP(),
		InterfaceName: DEFAULT_INTERFACE_NAME,
		Networks:      []JSONIPNet{},
	}

	IP := conf.MustAllocateIP()
	conf.IP = IP
	// DNS not set by default
	//conf.DNS = IP

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
	IP = IP.To4()

	if !(IP[0] == 10 || (IP[0] == 172 && IP[1] >= 16 && IP[1] <= 31) || (IP[0] == 192 && IP[1] == 168)) {
		// not private, so public
		return IP
	}

	// detect private IP and use icanhazip.com instead
	client := http.Client{
		Timeout: 5 * time.Second,
	}
	resp, err := client.Get("https://ipv4.icanhazip.com/")
	check(err)
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		body, err := ioutil.ReadAll(resp.Body)
		check(err)
		IP = net.ParseIP(strings.TrimSpace(string(body)))
		return IP.To4()
	}

	return net.IP{}
}
