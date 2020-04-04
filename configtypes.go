package dsnet

import (
	"encoding/json"
	"io/ioutil"
	"net"
	"os"
	"time"

	"github.com/go-playground/validator/v10"
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
	IP    net.IP    `validate:"required`
	Added time.Time `validate:"required"`
	// TODO ExternalIP support (Endpoint)
	//ExternalIP     net.UDPAddr `validate:"required,udp4_addr"`
	// TODO support routing additional networks (AllowedIPs)
	Networks     []JSONIPNet `validate:"required"`
	PublicKey    JSONKey     `validate:"required,len=44"`
	PrivateKey   JSONKey     `json:"-"` // omitted from config!
	PresharedKey JSONKey     `validate:"required,len=44"`
}

type DsnetConfig struct {
	// domain to append to hostnames. Relies on separate DNS server for
	// resolution. Informational only.
	ExternalIP    net.IP `validate:"required"`
	ListenPort    int    `validate:"gte=1024,lte=65535"`
	Domain        string `validate:"required,gte=1,lte=255"`
	InterfaceName string `validate:"required,gte=1,lte=255"`
	// IP network from which to allocate automatic sequential addresses
	// Network is chosen randomly when not specified
	Network JSONIPNet `validate:"required"`
	IP      net.IP    `validate:"required"`
	DNS     net.IP
	// extra networks available, will be added to AllowedIPs
	Networks []JSONIPNet `validate:"required"`
	// TODO Default subnets to route via VPN
	ReportFile string       `validate:"required"`
	PrivateKey JSONKey      `validate:"required,len=44"`
	Peers      []PeerConfig `validate:"dive"`
}

func MustLoadDsnetConfig() *DsnetConfig {
	raw, err := ioutil.ReadFile(CONFIG_FILE)

	if os.IsNotExist(err) {
		ExitFail("%s does not exist. `dsnet init` may be required.", CONFIG_FILE)
	} else if os.IsPermission(err) {
		ExitFail("%s cannot be accessed. Sudo may be required.", CONFIG_FILE)
	} else {
		check(err)
	}

	conf := DsnetConfig{}
	err = json.Unmarshal(raw, &conf)
	check(err)

	err = validator.New().Struct(conf)
	check(err)

	return &conf
}

func (conf *DsnetConfig) MustSave() {
	_json, _ := json.MarshalIndent(conf, "", "    ")
	err := ioutil.WriteFile(CONFIG_FILE, _json, 0600)
	check(err)
}

func (conf *DsnetConfig) MustAddPeer(peer PeerConfig) {
	// TODO validate all PeerConfig (keys etc)

	for _, p := range conf.Peers {
		if peer.Hostname == p.Hostname {
			ExitFail("%s is not an unique hostname", peer.Hostname)
		}
	}

	for _, p := range conf.Peers {
		if peer.PublicKey.Key == p.PublicKey.Key {
			ExitFail("%s is not an unique public key", peer.Hostname)
		}
	}

	for _, p := range conf.Peers {
		if peer.PresharedKey.Key == p.PresharedKey.Key {
			ExitFail("%s is not an unique preshared key", peer.Hostname)
		}
	}

	if conf.IPAllocated(peer.IP) {
		ExitFail("%s is already allocated", peer.IP)
	}

	for _, peerIPNet := range peer.Networks {
		if conf.IPAllocated(peerIPNet.IPNet.IP) {
			ExitFail("%s is already allocated", peerIPNet)
		}
	}

	conf.Peers = append(conf.Peers, peer)
}

func (conf *DsnetConfig) MustRemovePeer(hostname string) {
	peerIndex := -1

	for i, peer := range conf.Peers {
		if peer.Hostname == hostname {
			peerIndex = i
		}
	}

	if peerIndex == -1 {
		ExitFail("Could not find peer with hostname %s", hostname)
	}

	// remove peer from slice, retaining order
	copy(conf.Peers[peerIndex:], conf.Peers[peerIndex+1:]) // shift left
	conf.Peers = conf.Peers[:len(conf.Peers)-1] // truncate
}

func (conf DsnetConfig) IPAllocated(IP net.IP) bool {
	if IP.Equal(conf.IP) {
		return true
	}

	for _, peer := range conf.Peers {
		if IP.Equal(peer.IP) {
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

// choose a free IP for a new Peer
func (conf DsnetConfig) MustAllocateIP() net.IP {
	network := conf.Network.IPNet
	ones, bits := network.Mask.Size()
	zeros := bits - ones
	min := 1                // avoids network addr
	max := (1 << zeros) - 2 // avoids broadcast addr + overflow

	for i := min; i <= max; i++ {
		IP := make(net.IP, len(network.IP))
		copy(IP, network.IP)

		// OR the host part with the network part
		for j := 0; j < len(IP); j++ {
			shift := (len(IP) - j - 1) * 8
			IP[j] = IP[j] | byte(i>>shift)
		}

		if !conf.IPAllocated(IP) {
			return IP
		}
	}

	ExitFail("IP range exhausted")

	return net.IP{}
}

func (conf DsnetConfig) GetWgPeerConfigs() []wgtypes.PeerConfig {
	wgPeers := make([]wgtypes.PeerConfig, 0, len(conf.Peers))

	for _, peer := range conf.Peers {
		// create a new PSK in memory to avoid passing the same value by
		// pointer to each peer (d'oh)
		presharedKey := peer.PresharedKey.Key

		// AllowedIPs = private IP + defined networks
		allowedIPs := make([]net.IPNet, len(peer.Networks)+1)
		allowedIPs[0] = net.IPNet{
			IP:   peer.IP,
			Mask: net.IPMask{255, 255, 255, 255},
		}

		for i, net := range peer.Networks {
			allowedIPs[i+1] = net.IPNet
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
