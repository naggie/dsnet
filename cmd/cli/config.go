package cli

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"time"
	"strings"

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

func (conf *DsnetConfig) Merge(patch map[string]interface{}) error {
	// Merge the patch into the config
	
	if val, ok := patch["ExternalHostname"].(string); ok && val != "" {
		conf.ExternalHostname = val
	}
	if val, ok := patch["ExternalIP"].(string); ok && len(val) > 0 {
		conf.ExternalIP = net.ParseIP(val)
	}
	if val, ok := patch["ExternalIP6"].(string); ok && len(val) > 0 {
		conf.ExternalIP6 = net.ParseIP(val)
	}
	if val, ok := patch["ListenPort"].(int); ok && val > 0 {
		conf.ListenPort = val
	}
	if val, ok := patch["Domain"].(string); ok && val != "" {
		conf.Domain = val
	}
	if val, ok := patch["InterfaceName"].(string); ok && val != "" {
		conf.InterfaceName = val
	}
	if val, ok := patch["Network"].(string); ok && len(val) > 0 {
		net, err := lib.ParseJSONIPNet(val)
		if err != nil {
			return fmt.Errorf("failed to parse network: %w", err)
		}
		conf.Network = net
	}
	if val, ok := patch["Network6"].(string); ok && len(val) > 0 {
		net, err := lib.ParseJSONIPNet(val)
		if err != nil {
			return fmt.Errorf("failed to parse network6: %w", err)
		}
		conf.Network6 = net
	}
	if val, ok := patch["IP"].(string); ok && len(val) > 0 {
		conf.IP = net.ParseIP(val)
	}
	if val, ok := patch["IP6"].(string); ok && len(val) > 0 {
		conf.IP6 = net.ParseIP(val)
	}
	if val, ok := patch["DNS"].(string); ok && len(val) > 0 {
		conf.DNS = net.ParseIP(val)
	}
	if val, ok := patch["Networks"].([]string); ok && len(val) > 0 {
		conf.Networks = make([]lib.JSONIPNet, len(val))
		for i, v := range val {
			net, err := lib.ParseJSONIPNet(v)
			if err != nil {
				return fmt.Errorf("failed to parse network: %w", err)
			}
			conf.Networks[i] = net
		}
	}
	if val, ok := patch["PrivateKey"].(string); ok && len(val) > 0 {
		conf.PrivateKey = lib.JSONKey{}
		b64Key := strings.Trim(val, "\"")
		key, err := wgtypes.ParseKey(b64Key)
		if err != nil {
			return fmt.Errorf("failed to parse private key: %w", err)
		}
		conf.PrivateKey.Key = key
	}
	if val, ok := patch["PostUp"].(string); ok && len(val) > 0 {
		conf.PostUp = val
	}
	if val, ok := patch["PostDown"].(string); ok && len(val) > 0 {
		conf.PostDown = val
	}
	if val, ok := patch["Peers"].([]interface{}); ok && len(val) > 0 {
		conf.Peers = make([]PeerConfig, len(val))
		for i, v := range val {
			peerMap, ok := v.(map[string]interface{})
			if !ok {
				return fmt.Errorf("failed to parse peer: %v", v)
			}
			peer := PeerConfig{}
			// decode manually without peerstructure
			if val, ok := peerMap["Hostname"].(string); ok && val != "" {
				peer.Hostname = val
			} else {
				return fmt.Errorf("failed to parse peer hostname: %v", peerMap)
			}
			if val, ok := peerMap["Description"].(string); ok && val != "" {
				peer.Description = val
			} else {
				return fmt.Errorf("failed to parse peer description: %v", peerMap)
			}
			if val, ok := peerMap["Owner"].(string); ok && val != "" {
				peer.Owner = val
			} else {
				return fmt.Errorf("failed to parse peer owner: %v", peerMap)
			}
			if val, ok := peerMap["IP"].(string); ok && len(val) > 0 {
				peer.IP = net.ParseIP(val)
			} else {
				return fmt.Errorf("failed to parse peer IP: %v", peerMap)
			}
			if val, ok := peerMap["IP6"].(string); ok && len(val) > 0 {
				peer.IP6 = net.ParseIP(val)
			} else {
				return fmt.Errorf("failed to parse peer IP6: %v", peerMap)
			}
			if val, ok := peerMap["Added"].(string); ok && len(val) > 0 {
				t, err := time.Parse(time.RFC3339, val)
				if err != nil {
					return fmt.Errorf("failed to parse peer Added: %w", err)
				}
				peer.Added = t
			} else {
				return fmt.Errorf("failed to parse peer Added: %v", peerMap)
			}
			if val, ok := peerMap["Networks"].([]interface{}); ok && len(val) > 0 {
				peer.Networks = make([]lib.JSONIPNet, len(val))
				for j, v := range val {
					net, err := lib.ParseJSONIPNet(v.(string))
					if err != nil {
						return fmt.Errorf("failed to parse peer network: %w", err)
					}
					peer.Networks[j] = net
				}
			} else {
				return fmt.Errorf("failed to parse peer networks: %v", peerMap)
			}
			if val, ok := peerMap["PublicKey"].(string); ok && len(val) > 0 {
				b64Key := strings.Trim(val, "\"")
				key, err := wgtypes.ParseKey(b64Key)
				if err != nil {
					return fmt.Errorf("failed to parse peer public key: %w", err)
				}
				peer.PublicKey.Key = key
			} else {
				return fmt.Errorf("failed to parse peer public key: %v", peerMap)
			}

			if val, ok := peerMap["PresharedKey"].(string); ok && len(val) > 0 {
				b64Key := strings.Trim(val, "\"")
				key, err := wgtypes.ParseKey(b64Key)
				if err != nil {
					return fmt.Errorf("failed to parse peer preshared key: %w", err)
				}
				peer.PresharedKey.Key = key
			} else {
				return fmt.Errorf("failed to parse peer preshared key: %v", peerMap)
			}

			conf.Peers[i] = peer
		}
	}

	// Validate the updated configuration
	return validator.New().Struct(conf)
}
