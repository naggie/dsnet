package cli

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/naggie/dsnet/lib"
	"github.com/naggie/dsnet/lib/store"
	"github.com/spf13/viper"
)

func Init() error {
	listenPort := viper.GetInt("listen_port")
	MTU := viper.GetInt("mtu")
	interfaceName := viper.GetString("interface_name")
	fallbackWGBin := viper.GetString("fallback_wg_bin")

	backend, err := OpenStore()
	if err != nil {
		return fmt.Errorf("%w - failed to open storage backend", err)
	}
	defer backend.Close()

	// Refuse to overwrite an existing state.
	if _, _, err := backend.Load(context.Background()); err == nil {
		return fmt.Errorf("refusing to overwrite existing state")
	}

	privateKey, err := lib.GenerateJSONPrivateKey()
	if err != nil {
		return fmt.Errorf("%w - failed to generate private key", err)
	}

	externalIPV4, err := getExternalIP()
	if err != nil {
		return err
	}

	externalIPV6, err := getExternalIP6()
	if err != nil {
		return err
	}

	network, err := getPrivateNet()
	if err != nil {
		return fmt.Errorf("%w - failed to generate private network", err)
	}

	network6, err := getULANet()
	if err != nil {
		return fmt.Errorf("%w - failed to generate ULA network", err)
	}

	server := &lib.Server{
		PrivateKey:          privateKey,
		ListenPort:          listenPort,
		Network:             network,
		Network6:            network6,
		Peers:               []lib.Peer{},
		Domain:              "dsnet",
		ExternalIP:          externalIPV4,
		ExternalIP6:         externalIPV6,
		InterfaceName:       interfaceName,
		Networks:            []lib.JSONIPNet{},
		PersistentKeepalive: 25,
		MTU:                 MTU,
		FallbackWGBin:       fallbackWGBin,
	}

	ipv4, err := server.AllocateIP()
	if err != nil {
		return fmt.Errorf("%w - failed to allocate ipv4 address", err)
	}

	ipv6, err := server.AllocateIP6()
	if err != nil {
		return fmt.Errorf("%w - failed to allocate ipv6 address", err)
	}

	server.IP = ipv4
	server.IP6 = ipv6

	if len(server.ExternalIP) == 0 && len(server.ExternalIP6) == 0 {
		return fmt.Errorf("Could not determine any external IP, v4 or v6")
	}

	state := &store.State{
		Networks: map[string]*store.Network{
			server.InterfaceName: {Server: server},
		},
	}
	if err := backend.Save(context.Background(), state, ""); err != nil {
		return fmt.Errorf("%w - failed to save state", err)
	}

	fmt.Printf("Config written. Please check/edit.\n")
	return nil
}

// get a random IPv4  /22 subnet on 10.0.0.0 (1023 hosts) (or /24?)
func getPrivateNet() (lib.JSONIPNet, error) {
	rbs := make([]byte, 2)
	if _, err := rand.Read(rbs); err != nil {
		return lib.JSONIPNet{}, fmt.Errorf("%w - failed to read random bytes", err)
	}

	return lib.JSONIPNet{
		IPNet: net.IPNet{
			IP:   net.IP{10, rbs[0], rbs[1] << 2, 0},
			Mask: net.IPMask{255, 255, 252, 0},
		},
	}, nil
}

func getULANet() (lib.JSONIPNet, error) {
	rbs := make([]byte, 5)
	if _, err := rand.Read(rbs); err != nil {
		return lib.JSONIPNet{}, fmt.Errorf("%w - failed to read random bytes", err)
	}

	// fd00 prefix with 40 bit global id and zero (16 bit) subnet ID
	return lib.JSONIPNet{
		IPNet: net.IPNet{
			IP:   net.IP{0xfd, 0, rbs[0], rbs[1], rbs[2], rbs[3], rbs[4], 0, 0, 0, 0, 0, 0, 0, 0, 0},
			Mask: net.IPMask{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0, 0, 0, 0, 0, 0, 0, 0},
		},
	}, nil
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
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		IP = net.ParseIP(strings.TrimSpace(string(body)))
		return IP.To4(), nil
	}

	return nil, errors.New("failed to determine external ip")
}

func getExternalIP6() (net.IP, error) {
	var IP net.IP
	conn, err := net.Dial("udp", "2001:4860:4860::8888:53")
	if err == nil {
		defer conn.Close()

		localAddr := conn.LocalAddr().String()
		IP = net.ParseIP(strings.Split(localAddr, ":")[0])

		// check is not a ULA
		if IP[0] != 0xfd && IP[0] != 0xfc {
			return IP, nil
		}
	}

	client := http.Client{
		Timeout: 5 * time.Second,
	}
	resp, err := client.Get("https://ipv6.icanhazip.com/")
	if err == nil {
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return nil, err
			}
			IP = net.ParseIP(strings.TrimSpace(string(body)))
			return IP, nil
		}
	}

	return net.IP{}, nil
}
