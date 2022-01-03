package cli

import (
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/naggie/dsnet/lib"
	"github.com/spf13/viper"
)

func Init() {
	reportFile := viper.GetString("report_file")
	listenPort := viper.GetInt("listen_port")
	configFile := viper.GetString("config_file")
	interfaceName := viper.GetString("interface_name")

	_, err := os.Stat(configFile)

	if !os.IsNotExist(err) {
		ExitFail("Refusing to overwrite existing %s", configFile)
	}

	privateKey, err := lib.GenerateJSONPrivateKey()
	check(err, fmt.Sprintf("failed to generate private key: %s", err))

	externalIPV4, err := getExternalIP()
	check(err)

	conf := &DsnetConfig{
		PrivateKey:    privateKey,
		ListenPort:    listenPort,
		Network:       getPrivateNet(),
		Network6:      getULANet(),
		Peers:         []PeerConfig{},
		Domain:        "dsnet",
		ReportFile:    reportFile,
		ExternalIP:    externalIPV4,
		ExternalIP6:   getExternalIP6(),
		InterfaceName: interfaceName,
		Networks:      []lib.JSONIPNet{},
	}

	server := GetServer(conf)

	ipv4, err := server.AllocateIP()
	check(err, fmt.Sprintf("failed to allocate ipv4 address: %s", err))

	ipv6, err := server.AllocateIP6()
	check(err, fmt.Sprintf("failed to allocate ipv6 address: %s", err))

	conf.IP = ipv4
	conf.IP6 = ipv6

	if len(conf.ExternalIP) == 0 && len(conf.ExternalIP6) == 0 {
		ExitFail("Could not determine any external IP, v4 or v6")
	}

	conf.MustSave()

	fmt.Printf("Config written to %s. Please check/edit.\n", configFile)
}

// get a random IPv4  /22 subnet on 10.0.0.0 (1023 hosts) (or /24?)
func getPrivateNet() lib.JSONIPNet {
	rbs := make([]byte, 2)
	rand.Seed(time.Now().UTC().UnixNano())
	rand.Read(rbs)

	return lib.JSONIPNet{
		IPNet: net.IPNet{
			IP:   net.IP{10, rbs[0], rbs[1] << 2, 0},
			Mask: net.IPMask{255, 255, 252, 0},
		},
	}
}

func getULANet() lib.JSONIPNet {
	rbs := make([]byte, 5)
	rand.Seed(time.Now().UTC().UnixNano())
	rand.Read(rbs)

	// fd00 prefix with 40 bit global id and zero (16 bit) subnet ID
	return lib.JSONIPNet{
		IPNet: net.IPNet{
			IP:   net.IP{0xfd, 0, rbs[0], rbs[1], rbs[2], rbs[3], rbs[4], 0, 0, 0, 0, 0, 0, 0, 0, 0},
			Mask: net.IPMask{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0, 0, 0, 0, 0, 0, 0, 0},
		},
	}
}

// TODO factor getExternalIP + getExternalIP6
func getExternalIP() (net.IP, error) {
	var IP net.IP
	// arbitrary external IP is used (one that's guaranteed to route outside.
	// In this case, Google's DNS server. Doesn't actually need to be online.)
	conn, err := net.Dial("udp", "8.8.8.8:53")
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().String()
	IP = net.ParseIP(strings.Split(localAddr, ":")[0])
	IP = IP.To4()

	if !(IP[0] == 10 || (IP[0] == 172 && IP[1] >= 16 && IP[1] <= 31) || (IP[0] == 192 && IP[1] == 168)) {
		// not private, so public
		return IP, nil
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
		return IP.To4(), nil
	}

	return nil, errors.New("failed to determine external ip")
}

func getExternalIP6() net.IP {
	var IP net.IP
	conn, err := net.Dial("udp", "2001:4860:4860::8888:53")
	if err == nil {
		defer conn.Close()

		localAddr := conn.LocalAddr().String()
		IP = net.ParseIP(strings.Split(localAddr, ":")[0])

		// check is not a ULA
		if IP[0] != 0xfd && IP[0] != 0xfc {
			return IP
		}
	}

	client := http.Client{
		Timeout: 5 * time.Second,
	}
	resp, err := client.Get("https://ipv6.icanhazip.com/")
	if err == nil {
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			body, err := ioutil.ReadAll(resp.Body)
			check(err)
			IP = net.ParseIP(strings.TrimSpace(string(body)))
			return IP
		}
	}

	return net.IP{}
}
