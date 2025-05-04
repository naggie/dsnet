package cli

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
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
	PrivateKey lib.JSONKey `validate:"required,len=44"`
	PostUp     string
	PostDown   string
	Peers      []PeerConfig `validate:"dive"`
	// used for server and client
	PersistentKeepalive int `validate:"gte=0,lte=255"`
}

// LoadConfigFile parses the json config file, validates and stuffs
// it in to a struct
func LoadConfigFile() (*DsnetConfig, error) {
	configFile := viper.GetString("config_file")
	raw, err := ioutil.ReadFile(configFile)

	if os.IsNotExist(err) {
		return nil, fmt.Errorf("%s does not exist. `dsnet init` may be required", configFile)
	} else if os.IsPermission(err) {
		return nil, fmt.Errorf("%s cannot be accessed. Sudo may be required", configFile)
	} else if err != nil {
		return nil, err
	}

	conf := DsnetConfig{
		// set default for if key is not set. If it is set, this will not be
		// used _even if value is zero!_
		// Effectively, this is a migration
		PersistentKeepalive: 25,
	}

	err = json.Unmarshal(raw, &conf)
	if err != nil {
		return nil, err
	}

	err = validator.New().Struct(conf)
	if err != nil {
		return nil, err
	}

	if conf.ExternalHostname == "" && len(conf.ExternalIP) == 0 && len(conf.ExternalIP6) == 0 {
		return nil, fmt.Errorf("config does not contain ExternalIP, ExternalIP6 or ExternalHostname")
	}

	return &conf, nil
}

// Save writes the configuration to disk
func (conf *DsnetConfig) Save() error {
	configFile := viper.GetString("config_file")
	_json, _ := json.MarshalIndent(conf, "", "    ")
	_json = append(_json, '\n')
	err := ioutil.WriteFile(configFile, _json, 0600)
	if err != nil {
		return err
	}
	return nil
}

// AddPeer adds a provided peer to the Peers list in the conf
func (conf *DsnetConfig) AddPeer(peer lib.Peer) error {
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

	newPeerConfig := PeerConfig{
		Hostname:     peer.Hostname,
		Description:  peer.Description,
		Owner:        peer.Owner,
		IP:           peer.IP,
		IP6:          peer.IP6,
		Added:        peer.Added,
		Networks:     peer.Networks,
		PublicKey:    peer.PublicKey,
		PrivateKey:   peer.PrivateKey,
		PresharedKey: peer.PresharedKey,
	}

	conf.Peers = append(conf.Peers, newPeerConfig)
	return nil
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
		return fmt.Errorf("failed to find peer with hostname %s", hostname)
	}

	// remove peer from slice, retaining order
	copy(conf.Peers[peerIndex:], conf.Peers[peerIndex+1:]) // shift left
	conf.Peers = conf.Peers[:len(conf.Peers)-1]            // truncate
	return nil
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

func (conf *DsnetConfig) Merge(patch DsnetConfig) err {
	// Merge the patch into the config
	if patch.ExternalHostname != "" {
		conf.ExternalHostname = patch.ExternalHostname
	}
	if len(patch.ExternalIP) > 0 {
		conf.ExternalIP = patch.ExternalIP
	}
	if len(patch.ExternalIP6) > 0 {
		conf.ExternalIP6 = patch.ExternalIP6
	}
	if patch.ListenPort != 0 {
		conf.ListenPort = patch.ListenPort
	}
	if patch.Domain != "" {
		conf.Domain = patch.Domain
	}
	if patch.InterfaceName != "" {
		conf.InterfaceName = patch.InterfaceName
	}
	if len(patch.Network.IPNet.IP) > 0 {
		conf.Network = patch.Network
	}
	if len(patch.Network6.IPNet.IP) > 0 {
		conf.Network6 = patch.Network6
	}
	if len(patch.IP) > 0 {
		conf.IP = patch.IP
	}
	if len(patch.IP6) > 0 {
		conf.IP6 = patch.IP6
	}
	if len(patch.DNS) > 0 {
		conf.DNS = patch.DNS
	}
	if len(patch.Networks) > 0 {
		conf.Networks = patch.Networks
	}
	if len(patch.PrivateKey.Key) > 0 {
		conf.PrivateKey = patch.PrivateKey
	}
	if len(patch.PostUp) > 0 {
		conf.PostUp = patch.PostUp
	}
	if len(patch.PostDown) > 0 {
		conf.PostDown = patch.PostDown
	}
	if len(patch.Peers) > 0 {
		conf.Peers = patch.Peers
	}

	err = validator.New().Struct(conf)
	if err != nil {
		return nil, err
	}
}
