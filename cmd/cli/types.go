package cli

import (
	"strings"

	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

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

func GenerateJSONPrivateKey() (JSONKey, error) {
	privateKey, err := wgtypes.GeneratePrivateKey()
	if err != nil {
		return JSONKey{}, err
	}

	return JSONKey{
		Key: privateKey,
	}, nil
}

func GenerateJSONKey() (JSONKey, error) {
	privateKey, err := wgtypes.GenerateKey()

	if err != nil {
		return JSONKey{}, err
	}

	return JSONKey{
		Key: privateKey,
	}, err
}
