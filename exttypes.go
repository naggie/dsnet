package dsnet

import (
	"net"
	"strings"

	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

type JSONIPNet struct {
	IPNet net.IPNet
}

func (n JSONIPNet) MarshalJSON() ([]byte, error) {
	return []byte("\"" + n.IPNet.String() + "\""), nil
}

func (n *JSONIPNet) UnmarshalJSON(b []byte) error {
	cidr := strings.Trim(string(b), "\"")
	IP, IPNet, err := net.ParseCIDR(cidr)
	IPNet.IP = IP
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
