package cli

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net"
	"os"
	"time"

	"github.com/go-playground/validator"
	"github.com/naggie/dsnet/lib"
	"github.com/spf13/viper"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

// see https://github.com/WireGuard/wgctrl-go/blob/master/wgtypes/types.go for definitions
type PeerConfig struct {
	// Used to update DNS
	Hostname string `validate:"required,gte=1,lte=255"`
	// username of person running this host/router
	Owner string `validate:"required,gte=1,lte=255"`
	// Description of what the host is and/or does
	Description string `validate:"required,gte=1,lte=255"`
	// Internal VPN IP address. Added to AllowedIPs in server config as a /32
	IP    net.IP
	IP6   net.IP
	Added time.Time `validate:"required"`
	// TODO ExternalIP support (Endpoint)
	//ExternalIP     net.UDPAddr `validate:"required,udp4_addr"`
	// TODO support routing additional networks (AllowedIPs)
	Networks     []lib.JSONIPNet `validate:"required"`
	PublicKey    lib.JSONKey     `validate:"required,len=44"`
	PrivateKey   lib.JSONKey     `json:"-"` // omitted from config!
	PresharedKey lib.JSONKey     `validate:"required,len=44"`
}

type DsnetConfig struct {
	// When generating configs, the ExternalHostname has precendence for the
	// server Endpoint, followed by ExternalIP (IPv4) and ExternalIP6 (IPv6)
	// The IPs are discovered automatically on init. Define an ExternalHostname
	// if you're using dynamic DNS, want to change IPs without updating
	// configs, or want wireguard to be able to choose between IPv4/IPv6. It is
	// only possible to specify one Endpoint per peer entry in wireguard.
	ExternalHostname string
	ExternalIP       net.IP
	ExternalIP6      net.IP
	ListenPort       int `validate:"gte=1,lte=65535"`
	// domain to append to hostnames. Relies on separate DNS server for
	// resolution. Informational only.
	Domain        string `validate:"required,gte=1,lte=255"`
	InterfaceName string `validate:"required,gte=1,lte=255"`
	// IP network from which to allocate automatic sequential addresses
	// Network is chosen randomly when not specified
	Network  lib.JSONIPNet `validate:"required"`
	Network6 lib.JSONIPNet `validate:"required"`
	IP       net.IP
	IP6      net.IP
	DNS      net.IP
	// extra networks available, will be added to AllowedIPs
	Networks []lib.JSONIPNet `validate:"required"`
	// TODO Default subnets to route via VPN
	ReportFile string      `validate:"required"`
	PrivateKey lib.JSONKey `validate:"required,len=44"`
	PostUp     string
	PostDown   string
	Peers      []PeerConfig `validate:"dive"`
}

// MustLoadDsnetConfig is like LoadDsnetConfig, except it exits on error
func MustLoadDsnetConfig() *DsnetConfig {
	conf, err := LoadDsnetConfig()
	check(err)
	return conf
}

// LoadDsnetConfig parses the json config file, validates and stuffs
// it in to a struct
func LoadDsnetConfig() (*DsnetConfig, error) {
	configFile := viper.GetString("config_file")
	raw, err := ioutil.ReadFile(configFile)

	if os.IsNotExist(err) {
		return nil, fmt.Errorf("%s does not exist. `dsnet init` may be required.", configFile)
	} else if os.IsPermission(err) {
		return nil, fmt.Errorf("%s cannot be accessed. Sudo may be required.", configFile)
	} else if err != nil {
		return nil, err
	}

	conf := DsnetConfig{}
	err = json.Unmarshal(raw, &conf)
	if err != nil {
		return nil, err
	}

	err = validator.New().Struct(conf)
	if err != nil {
		return nil, err
	}

	if conf.ExternalHostname == "" && len(conf.ExternalIP) == 0 && len(conf.ExternalIP6) == 0 {
		return nil, fmt.Errorf("Config does not contain ExternalIP, ExternalIP6 or ExternalHostname")
	}

	return &conf, nil
}

// Save writes the configuration to disk
func (conf *DsnetConfig) Save() error {
	configFile := viper.GetString("config_file")
	_json, _ := json.MarshalIndent(conf, "", "    ")
	err := ioutil.WriteFile(configFile, _json, 0600)
	if err != nil {
		return err
	}
	return nil
}

// MustSave is like Save except it exits on error
func (conf *DsnetConfig) MustSave() {
	err := conf.Save()
	check(err)
}

// AddPeer adds a provided peer to the Peers list in the conf
func (conf *DsnetConfig) AddPeer(peer PeerConfig) error {
	// TODO validate all PeerConfig (keys etc)

	for _, p := range conf.Peers {
		if peer.Hostname == p.Hostname {
			return fmt.Errorf("%s is not an unique hostname", peer.Hostname)
		}
	}

	for _, p := range conf.Peers {
		if peer.PublicKey.Key == p.PublicKey.Key {
			return fmt.Errorf("%s is not an unique public key", peer.Hostname)
		}
	}

	for _, p := range conf.Peers {
		if peer.PresharedKey.Key == p.PresharedKey.Key {
			return fmt.Errorf("%s is not an unique preshared key", peer.Hostname)
		}
	}

	if conf.IPAllocated(peer.IP) {
		return fmt.Errorf("%s is already allocated", peer.IP)
	}

	for _, peerIPNet := range peer.Networks {
		if conf.IPAllocated(peerIPNet.IPNet.IP) {
			return fmt.Errorf("%s is already allocated", peerIPNet)
		}
	}

	conf.Peers = append(conf.Peers, peer)
	return nil
}

// MustAddPeer is like AddPeer, except it exist on error
func (conf *DsnetConfig) MustAddPeer(peer PeerConfig) {
	err := conf.AddPeer(peer)
	check(err)
}

// RemovePeer removes a peer from the peer list based on hostname
func (conf *DsnetConfig) RemovePeer(hostname string) error {
	peerIndex := -1

	for i, peer := range conf.Peers {
		if peer.Hostname == hostname {
			peerIndex = i
		}
	}

	if peerIndex == -1 {
		return fmt.Errorf("Could not find peer with hostname %s", hostname)
	}

	// remove peer from slice, retaining order
	copy(conf.Peers[peerIndex:], conf.Peers[peerIndex+1:]) // shift left
	conf.Peers = conf.Peers[:len(conf.Peers)-1]            // truncate
	return nil
}

// MustRemovePeer is like RemovePeer, except it exits on error
func (conf *DsnetConfig) MustRemovePeer(hostname string) {
	err := conf.RemovePeer(hostname)
	check(err)
}

// IPAllocated checks the existing used ips and returns bool
// depending on if the IP is in use
func (conf DsnetConfig) IPAllocated(IP net.IP) bool {
	if IP.Equal(conf.IP) || IP.Equal(conf.IP6) {
		return true
	}

	for _, peer := range conf.Peers {
		if IP.Equal(peer.IP) || IP.Equal(peer.IP6) {
			return true
		}

		for _, peerIPNet := range peer.Networks {
			if IP.Equal(peerIPNet.IPNet.IP) {
				return true
			}
		}
	}

	return false
}

// AllocateIP finds a free IPv4 for a new Peer (sequential allocation)
func (conf DsnetConfig) AllocateIP() (net.IP, error) {
	network := conf.Network.IPNet
	ones, bits := network.Mask.Size()
	zeros := bits - ones

	// avoids network addr
	min := 1
	// avoids broadcast addr + overflow
	max := (1 << zeros) - 2

	IP := make(net.IP, len(network.IP))

	for i := min; i <= max; i++ {
		// dst, src!
		copy(IP, network.IP)

		// OR the host part with the network part
		for j := 0; j < len(IP); j++ {
			shift := (len(IP) - j - 1) * 8
			IP[j] = IP[j] | byte(i>>shift)
		}

		if !conf.IPAllocated(IP) {
			return IP, nil
		}
	}

	return nil, fmt.Errorf("IP range exhausted")
}

// MustAllocateIP is like AllocateIP, except it exits on error
func (conf DsnetConfig) MustAllocateIP() net.IP {
	ip, err := conf.AllocateIP()
	check(err)
	return ip
}

// AllocateIP6 finds a free IPv6 for a new Peer (pseudorandom allocation)
func (conf DsnetConfig) AllocateIP6() (net.IP, error) {
	network := conf.Network6.IPNet
	ones, bits := network.Mask.Size()
	zeros := bits - ones

	rbs := make([]byte, zeros)
	rand.Seed(time.Now().UTC().UnixNano())

	IP := make(net.IP, len(network.IP))

	for i := 0; i <= 10000; i++ {
		rand.Read(rbs)
		// dst, src! Copy prefix of IP
		copy(IP, network.IP)

		// OR the host part with the network part
		for j := ones / 8; j < len(IP); j++ {
			IP[j] = IP[j] | rbs[j]
		}

		if !conf.IPAllocated(IP) {
			return IP, nil
		}
	}

	return nil, fmt.Errorf("Could not allocate random IPv6 after 10000 tries. This was highly unlikely!")
}

// MustAllocateIP6 is like AllocateIP6, except it exits on error
func (conf DsnetConfig) MustAllocateIP6() net.IP {
	ip, err := conf.AllocateIP6()
	check(err)
	return ip
}

func (conf DsnetConfig) GetWgPeerConfigs() []wgtypes.PeerConfig {
	wgPeers := make([]wgtypes.PeerConfig, 0, len(conf.Peers))

	for _, peer := range conf.Peers {
		// create a new PSK in memory to avoid passing the same value by
		// pointer to each peer (d'oh)
		presharedKey := peer.PresharedKey.Key

		// AllowedIPs = private IP + defined networks
		allowedIPs := make([]net.IPNet, 0, len(peer.Networks)+2)

		if len(peer.IP) > 0 {
			allowedIPs = append(
				allowedIPs,
				net.IPNet{
					IP:   peer.IP,
					Mask: net.IPMask{255, 255, 255, 255},
				},
			)
		}

		if len(peer.IP6) > 0 {
			allowedIPs = append(
				allowedIPs,
				net.IPNet{
					IP:   peer.IP6,
					Mask: net.IPMask{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
				},
			)
		}

		for _, net := range peer.Networks {
			allowedIPs = append(allowedIPs, net.IPNet)
		}

		wgPeers = append(wgPeers, wgtypes.PeerConfig{
			PublicKey:         peer.PublicKey.Key,
			Remove:            false,
			UpdateOnly:        false,
			PresharedKey:      &presharedKey,
			Endpoint:          nil,
			ReplaceAllowedIPs: true,
			AllowedIPs:        allowedIPs,
		})
	}

	return wgPeers
}
