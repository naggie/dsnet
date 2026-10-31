package lib

import (
	"fmt"
	"math/rand"
	"net"
	"time"

	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

type Server struct {
	ExternalHostname    string
	ExternalIP          net.IP
	ExternalIP6         net.IP
	ListenPort          int
	Domain              string
	InterfaceName       string
	Network             JSONIPNet
	Network6            JSONIPNet
	IP                  net.IP
	IP6                 net.IP
	DNS                 net.IP
	PrivateKey          JSONKey
	PostUp              string
	PostDown            string
	FallbackWGBin       string
	Peers               []Peer
	Networks            []JSONIPNet
	PersistentKeepalive int
	MTU                 int
	ClientMTU           int
}

func (s *Server) GetPeers() []wgtypes.PeerConfig {
	wgPeers := make([]wgtypes.PeerConfig, 0, len(s.Peers))

	for _, peer := range s.Peers {
		// omitting a paused peer makes ConfigureDevice remove it from the
		// interface, blocking its traffic
		if peer.Paused {
			continue
		}

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

// AllocateIP finds a free IPv4 for a new Peer (sequential allocation)
func (s *Server) AllocateIP() (net.IP, error) {
	network := s.Network.IPNet
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

		if !s.IPAllocated(IP) {
			return IP, nil
		}
	}

	return nil, fmt.Errorf("IP range exhausted")
}

// AllocateIP6 finds a free IPv6 for a new Peer (pseudorandom allocation)
func (s *Server) AllocateIP6() (net.IP, error) {
	network := s.Network6.IPNet
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

		if !s.IPAllocated(IP) {
			return IP, nil
		}
	}

	return nil, fmt.Errorf("Could not allocate random IPv6 after 10000 tries. This was highly unlikely!")
}

// checkFreeIP returns an error unless a user-supplied IP is a usable host
// address in network and not already in use
func (s *Server) checkFreeIP(IP net.IP, network net.IPNet) error {
	if len(network.Mask) == 0 || !network.Contains(IP) {
		return fmt.Errorf("%s is not in network %s", IP, network.String())
	}
	if IP.Equal(network.IP.Mask(network.Mask)) {
		return fmt.Errorf("%s is the network address of %s", IP, network.String())
	}
	if IP.To4() != nil {
		broadcast := make(net.IP, net.IPv4len)
		for i, b := range network.IP.To4() {
			broadcast[i] = b | ^network.Mask[len(network.Mask)-net.IPv4len+i]
		}
		if IP.Equal(broadcast) {
			return fmt.Errorf("%s is the broadcast address of %s", IP, network.String())
		}
	}
	if s.IPAllocated(IP) {
		return fmt.Errorf("%s is already allocated", IP)
	}
	return nil
}

// IPAllocated checks the existing used ips and returns bool
// depending on if the IP is in use
func (s *Server) IPAllocated(IP net.IP) bool {
	if IP.Equal(s.IP) || IP.Equal(s.IP6) {
		return true
	}

	for _, peer := range s.Peers {
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
