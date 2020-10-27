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
	if len(n.IPNet.IP) == 0 {
		return []byte("\"\""), nil
	} else {
		return []byte("\"" + n.IPNet.String() + "\""), nil
	}
}

func (n *JSONIPNet) UnmarshalJSON(b []byte) error {
	cidr := strings.Trim(string(b), "\"")

	if cidr == "" {
		// Leave as empty/uninitialised IPNet. A bit like omitempty behaviour,
		// but we can leave the field there and blank which is useful if the
		// user wishes to add the cidr manually.
		return nil
	}

	IP, IPNet, err := net.ParseCIDR(cidr)

	if err == nil {
		IPNet.IP = IP
		n.IPNet = *IPNet
	}

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
