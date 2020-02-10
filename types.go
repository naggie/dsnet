package dsnet

import (
	"net"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

// see https://github.com/WireGuard/wgctrl-go/blob/master/wgtypes/types.go for definitions
type Peer struct {
	Name string
	Description string
	PublicKey wgtypes.Key
	PresharedKey wgtypes.Key
	Endpoint *net.UDPAddr
	LastHandshakeTime time.Time
	ReceiveBytes int64
	TransmitBytes int64
	AllowedIPs []net.IPNet
}
