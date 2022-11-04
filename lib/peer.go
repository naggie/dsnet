package lib

import (
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

// PeerType is what configuration to use when generating
// peer config files
type PeerType int

const (
	// WGQuick is used by wg-quick to set up a peer
	// https://manpages.debian.org/unstable/wireguard-tools/wg-quick.8.en.html
	WGQuick PeerType = iota
	// Vyatta is used by Ubiquiti routers
	// https://github.com/WireGuard/wireguard-vyatta-ubnt/
	Vyatta
	// NixOS is a declartive linux distro
	// https://nixos.wiki/wiki/Wireguard
	NixOS
)

type Peer struct {
	Hostname            string
	Owner               string
	Description         string
	IP                  net.IP
	IP6                 net.IP
	Added               time.Time
	PublicKey           JSONKey
	PrivateKey          JSONKey
	PresharedKey        JSONKey
	Networks            []JSONIPNet
	PersistentKeepalive int
}

// NewPeer generates a peer from the supplied arguments and generates keys if needed.
//   - server is required and provides network information
//   - private is a base64-encoded private key; if the empty string, a new key will be generated
//   - public is a base64-encoded public key. If empty, it will be generated from the private key.
//     If **not** empty, the private key will be included IFF a private key was provided.
//   - owner is the owner name (required)
//   - hostname is the name of the peer (required)
//   - description is the annotation for the peer
func NewPeer(server *Server, private, public, owner, hostname, description string) (Peer, error) {
	if owner == "" {
		return Peer{}, errors.New("missing owner")
	}
	if hostname == "" {
		return Peer{}, errors.New("missing hostname")
	}

	var privateKey JSONKey
	if private != "" {
		userKey := &JSONKey{}
		userKey.UnmarshalJSON([]byte(private))
		privateKey = *userKey
	} else {
		var err error
		privateKey, err = GenerateJSONPrivateKey()
		if err != nil {
			return Peer{}, fmt.Errorf("failed to generate private key: %s", err)
		}
	}

	var publicKey JSONKey
	if public != "" {
		b64Key := strings.Trim(string(public), "\"")
		key, err := wgtypes.ParseKey(b64Key)
		if err != nil {
			return Peer{}, err
		}
		publicKey = JSONKey{Key: key}
		if private == "" {
			privateKey = JSONKey{Key: wgtypes.Key([wgtypes.KeyLen]byte{})}
		} else {
			pubK := privateKey.PublicKey()
			ascK := pubK.Key.String()
			if ascK != public {
				return Peer{}, fmt.Errorf("user-supplied private and public keys are not related")
			}
		}
	} else {
		publicKey = privateKey.PublicKey()
	}

	presharedKey, err := GenerateJSONKey()
	if err != nil {
		return Peer{}, fmt.Errorf("failed to generate private key: %s", err)
	}

	newPeer := Peer{
		Owner:        owner,
		Hostname:     hostname,
		Description:  description,
		Added:        time.Now(),
		PublicKey:    publicKey,
		PrivateKey:   privateKey,
		PresharedKey: presharedKey,
		Networks:     []JSONIPNet{},
		// inherit from server setting, which is derived from config
		PersistentKeepalive: server.PersistentKeepalive,
	}

	if len(server.Network.IPNet.Mask) > 0 {
		newIP, err := server.AllocateIP()
		if err != nil {
			return Peer{}, fmt.Errorf("failed to allocate ipv4 address: %s", err)
		}
		newPeer.IP = newIP
	}

	if len(server.Network6.IPNet.Mask) > 0 {
		newIPV6, err := server.AllocateIP6()
		if err != nil {
			return Peer{}, fmt.Errorf("failed to allocate ipv6 address: %s", err)
		}
		newPeer.IP6 = newIPV6
	}

	if len(server.IP) == 0 && len(server.IP6) == 0 {
		return Peer{}, fmt.Errorf("no IPv4 or IPv6 network defined in config")
	}
	return newPeer, nil
}
