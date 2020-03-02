package dsnet

import (
	"encoding/json"
	"io/ioutil"
	"net"
	"strings"
	"time"
	"fmt"
	"errors"

	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

// see https://github.com/WireGuard/wgctrl-go/blob/master/wgtypes/types.go for definitions
type PeerConfig struct {
	// username of person running this host/router
	Owner string `validate:"required,gte=1,lte=255"`
	// Used to update DNS
	Hostname string `validate:"required,gte=1,lte=255"`
	// Description of what the host is and/or does
	Description string `validate:"required,gte=1,lte=255"`

	PublicKey    JSONKey     `validate:"required,len=44"`
	PresharedKey JSONKey     `validate:"required,len=44"`
	Endpoint     net.UDPAddr `validate:"required,udp4_addr"`
	AllowedIPs   []net.IPNet `validate:"dive,required,cidr"`
}

type Peer struct {
	// username of person running this host/router
	Owner string
	// Used to update DNS
	Hostname string
	// Description of what the host is and/or does
	Description string
	// whether last heartbeat/rxdata was received (50% margin)
	Online bool
	// if no data for x days, consider revoking access
	Expired bool

	PublicKey         wgtypes.Key
	PresharedKey      wgtypes.Key
	Endpoint          *net.UDPAddr
	LastHandshakeTime time.Time
	ReceiveBytes      int64
	TransmitBytes     int64
	AllowedIPs        []net.IPNet
}

type DsnetConfig struct {
	PrivateKey   JSONKey `validate:"required,len=44"`
	PresharedKey JSONKey `validate:"required,len=44"`
	ListenPort   int     `validate:"gte=1024,lte=65535"`
	Peers        []PeerConfig
	// IP network from which to allocate automatic sequential addresses
	// Network is chosen randomly when not specified
	Network JSONIPNet `validate:"required"`
	// domain to append to hostnames. Relies on separate DNS server for
	// resolution. Informational only.
	Domain string `validate:"required,gte=1,lte=255"`
	// TODO Default subnets to route via VPN
	ReportFile string `validate:"required"`
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
	// TODO validate PeerConfig!!!
	conf.Peers = append(conf.Peers, peer)
}

func (conf DsnetConfig) IPAllocated(IP net.IP) bool {
	for _, peer := range conf.Peers {
		for _, peerIPNet := range peer.AllowedIPs {
			if IP.Equal(peerIPNet.IP) {
				return false
			}
		}
	}

	return true
}

// choose a free IP for a new Peer
func (conf DsnetConfig) ChooseIP() (net.IP, error) {
	network := conf.Network.IPNet
	ones, bits := network.Mask.Size()
	zeros := bits - ones
	min := 1  // avoids network addr
	max := (1 << zeros) -2  // avoids broadcast addr + overflow

	for i := min; i <= max; i++ {
		IP := make(net.IP, len(network.IP))
		copy(IP, network.IP)

		// OR the host part with the network part
		for j := 0; j < len(IP); j++ {
			shift := (len(IP) - j - 1) * 8
			IP[j] = IP[j] | byte(i>>shift)

			if conf.IPAllocated(IP) {
				return IP, nil
			}
		}
	}

	return net.IP{}, errors.New("IP range exhausted")
}

type Dsnet struct {
	Name       string
	PrivateKey wgtypes.Key
	PublicKey  wgtypes.Key
	ListenPort int
	Peers      []Peer
}

type JSONIPNet struct {
	IPNet net.IPNet
}

func (n JSONIPNet) MarshalJSON() ([]byte, error) {
	return []byte("\"" + n.IPNet.String() + "\""), nil
}

func (n *JSONIPNet) UnmarshalJSON(b []byte) error {
	cidr := strings.Trim(string(b), "\"")
	_, IPNet, err := net.ParseCIDR(cidr)
	n.IPNet = *IPNet
	return err
}

func (n *JSONIPNet) String() string {
	return n.IPNet.String()
}

type JSONKey struct {
	Key wgtypes.Key
}

func (k JSONKey) MarshalJSON() ([]byte, error) {
	return []byte("\"" + k.Key.String() + "\""), nil
}

func (k JSONKey) PublicKey() JSONKey {
	return JSONKey{
		Key: k.Key.PublicKey(),
	}
}

func (k *JSONKey) UnmarshalJSON(b []byte) error {
	b64Key := strings.Trim(string(b), "\"")
	key, err := wgtypes.ParseKey(b64Key)
	k.Key = key
	return err
}

func GenerateJSONPrivateKey() JSONKey {
	privateKey, err := wgtypes.GeneratePrivateKey()

	check(err)

	return JSONKey{
		Key: privateKey,
	}
}

func GenerateJSONKey() JSONKey {
	privateKey, err := wgtypes.GenerateKey()

	check(err)

	return JSONKey{
		Key: privateKey,
	}
}
