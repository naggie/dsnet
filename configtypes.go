package dsnet

import (
	"encoding/json"
	"io/ioutil"
	"net"
)

// see https://github.com/WireGuard/wgctrl-go/blob/master/wgtypes/types.go for definitions
type PeerConfig struct {
	// Used to update DNS
	Hostname string `validate:"required,gte=1,lte=255"`
	// username of person running this host/router
	Owner string `validate:"required,gte=1,lte=255"`
	// Description of what the host is and/or does
	Description string `validate:"required,gte=1,lte=255"`

	PublicKey    JSONKey `validate:"required,len=44"`
	PrivateKey   JSONKey `json:"-"` // omitted from config!
	PresharedKey JSONKey `validate:"required,len=44"`
	// TODO endpoint support
	//Endpoint     net.UDPAddr `validate:"required,udp4_addr"`
	AllowedIPs []JSONIPNet `validate:"dive,required,cidr"`
}

type DsnetConfig struct {
	// domain to append to hostnames. Relies on separate DNS server for
	// resolution. Informational only.
	ExternalIP net.IP `validate:"required,cidr"`
	ListenPort int    `validate:"gte=1024,lte=65535"`
	Domain     string `validate:"required,gte=1,lte=255"`
	// IP network from which to allocate automatic sequential addresses
	// Network is chosen randomly when not specified
	Network     JSONIPNet `validate:"required"`
	InternalIP  net.IP    `validate:"required,cidr"`
	InternalDNS net.IP    `validate:"required,cidr"`
	// TODO Default subnets to route via VPN
	ReportFile   string  `validate:"required"`
	PrivateKey   JSONKey `validate:"required,len=44"`
	PresharedKey JSONKey `validate:"required,len=44"`
	Peers        []PeerConfig
}

func MustLoadDsnetConfig() *DsnetConfig {
	raw, err := ioutil.ReadFile(CONFIG_FILE)
	check(err)
	conf := DsnetConfig{}
	err = json.Unmarshal(raw, &conf)
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

	for _, peerIPNet := range peer.AllowedIPs {
		if conf.IPAllocated(peerIPNet.IPNet.IP) {
			ExitFail("%s is already allocated", peerIPNet)
		}
	}

	conf.Peers = append(conf.Peers, peer)
}

func (conf DsnetConfig) IPAllocated(IP net.IP) bool {
	if IP.Equal(conf.InternalIP) {
		return true
	}

	for _, peer := range conf.Peers {
		for _, peerIPNet := range peer.AllowedIPs {
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
