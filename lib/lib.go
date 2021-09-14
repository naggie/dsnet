package lib

import (
	"fmt"

	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

func (s *Server) Up() error {
	if err := s.CreateLink(); err != nil {
		return err
	}
	return s.ConfigureDevice()
}

// ConfigureDevice sets up the WG interface
func (s *Server) ConfigureDevice() error {
	wg, err := wgctrl.New()
	if err != nil {
		return err
	}
	defer wg.Close()

	dev, err := wg.Device(s.InterfaceName)

	if err != nil {
		return fmt.Errorf("could not retrieve device '%s' (%v)", s.InterfaceName, err)
	}

	peers := s.GetPeers()

	// compare peers to see if any exist on the device and not the config. If
	// so, they should be removed by appending a dummy peer with Remove:true + pubkey.
	knownKeys := make(map[wgtypes.Key]bool)

	for _, peer := range peers {
		knownKeys[peer.PublicKey] = true
	}

	// find deleted peers, and append dummy "remove" peers
	for _, peer := range dev.Peers {
		if !knownKeys[peer.PublicKey] {
			peers = append(peers, wgtypes.PeerConfig{
				PublicKey: peer.PublicKey,
				Remove:    true,
			})
		}
	}

	wgConfig := wgtypes.Config{
		PrivateKey: &s.PrivateKey.Key,
		ListenPort: &s.ListenPort,
		// ReplacePeers with the same peers results in those peers losing
		// connection, so it's not possible to do declarative configuration
		// idempotently with ReplacePeers like I had assumed. Instead, peers
		// must be removed imperatively with Remove:true. Peers can still be
		// added/updated with ConfigureDevice declaratively.
		ReplacePeers: false,
		Peers:        peers,
	}

	err = wg.ConfigureDevice(s.InterfaceName, wgConfig)

	if err != nil {
		return fmt.Errorf("could not configure device '%s' (%v)", s.InterfaceName, err)
	}
	return nil
}
