package lib

import (
	"net"
	"time"
)

type Peer struct {
	Hostname     string
	Owner        string
	Description  string
	IP           net.IP
	IP6          net.IP
	Added        time.Time
	PublicKey    JSONKey
	PrivateKey   JSONKey
	PresharedKey JSONKey
	Networks     []JSONIPNet
}
