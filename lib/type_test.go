package lib

import (
	"encoding/json"
	"net"
	"testing"

	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

func TestJSONIPNetMarshalEmpty(t *testing.T) {
	n := JSONIPNet{}
	b, err := n.MarshalJSON()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(b) != `""` {
		t.Fatalf("expected empty string, got %s", b)
	}
}

func TestJSONIPNetMarshalCIDR(t *testing.T) {
	_, ipnet, _ := net.ParseCIDR("10.0.0.0/24")
	n := JSONIPNet{IPNet: *ipnet}
	b, err := n.MarshalJSON()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(b) != `"10.0.0.0/24"` {
		t.Fatalf("expected \"10.0.0.0/24\", got %s", b)
	}
}

func TestJSONIPNetUnmarshalValid(t *testing.T) {
	var n JSONIPNet
	err := n.UnmarshalJSON([]byte(`"10.1.2.0/22"`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n.IPNet.IP.String() != "10.1.2.0" {
		t.Fatalf("expected IP 10.1.2.0, got %s", n.IPNet.IP)
	}
	ones, _ := n.IPNet.Mask.Size()
	if ones != 22 {
		t.Fatalf("expected /22, got /%d", ones)
	}
}

func TestJSONIPNetUnmarshalEmpty(t *testing.T) {
	var n JSONIPNet
	err := n.UnmarshalJSON([]byte(`""`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(n.IPNet.IP) != 0 {
		t.Fatalf("expected empty IP, got %s", n.IPNet.IP)
	}
}

func TestJSONIPNetUnmarshalInvalid(t *testing.T) {
	var n JSONIPNet
	err := n.UnmarshalJSON([]byte(`"not-a-cidr"`))
	if err == nil {
		t.Fatal("expected error for invalid CIDR")
	}
}

func TestJSONIPNetString(t *testing.T) {
	_, ipnet, _ := net.ParseCIDR("192.168.1.0/24")
	n := &JSONIPNet{IPNet: *ipnet}
	s := n.String()
	if s != "192.168.1.0/24" {
		t.Fatalf("expected 192.168.1.0/24, got %s", s)
	}
}

func TestJSONIPNetRoundTrip(t *testing.T) {
	original := JSONIPNet{}
	if err := original.UnmarshalJSON([]byte(`"10.5.4.0/22"`)); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	b, err := original.MarshalJSON()
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var decoded JSONIPNet
	err = decoded.UnmarshalJSON(b)
	if err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if original.IPNet.String() != decoded.IPNet.String() {
		t.Fatalf("round trip mismatch: %s != %s", original.IPNet.String(), decoded.IPNet.String())
	}
}

func TestJSONIPNetIPv6(t *testing.T) {
	var n JSONIPNet
	err := n.UnmarshalJSON([]byte(`"fd00:1234::/64"`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	b, err := n.MarshalJSON()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(b) != `"fd00:1234::/64"` {
		t.Fatalf("expected \"fd00:1234::/64\", got %s", b)
	}
}

func TestJSONKeyMarshalUnmarshal(t *testing.T) {
	// Generate a real key
	privKey, err := wgtypes.GeneratePrivateKey()
	if err != nil {
		t.Fatalf("failed to generate key: %v", err)
	}

	k := JSONKey{Key: privKey}
	b, err := k.MarshalJSON()
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var k2 JSONKey
	err = k2.UnmarshalJSON(b)
	if err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if k.Key != k2.Key {
		t.Fatalf("round trip mismatch: keys differ")
	}
}

func TestJSONKeyPublicKey(t *testing.T) {
	privKey, err := wgtypes.GeneratePrivateKey()
	if err != nil {
		t.Fatalf("failed to generate key: %v", err)
	}

	k := JSONKey{Key: privKey}
	pub := k.PublicKey()

	expected := privKey.PublicKey()
	if pub.Key != expected {
		t.Fatalf("public key mismatch")
	}
}

func TestGenerateJSONPrivateKey(t *testing.T) {
	k1, err := GenerateJSONPrivateKey()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	k2, err := GenerateJSONPrivateKey()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Two generated keys should differ
	if k1.Key == k2.Key {
		t.Fatal("two generated private keys should not be identical")
	}

	// Should produce valid public key
	pub := k1.PublicKey()
	if pub.Key == k1.Key {
		t.Fatal("public key should differ from private key")
	}
}

func TestGenerateJSONKey(t *testing.T) {
	k, err := GenerateJSONKey()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Key should not be zero
	zero := wgtypes.Key{}
	if k.Key == zero {
		t.Fatal("generated key should not be zero")
	}
}

func TestJSONKeyInStruct(t *testing.T) {
	// Test that JSONKey works properly when embedded in a JSON struct
	type testStruct struct {
		Name string  `json:"name"`
		Key  JSONKey `json:"key"`
	}

	privKey, _ := wgtypes.GeneratePrivateKey()
	original := testStruct{
		Name: "test",
		Key:  JSONKey{Key: privKey},
	}

	b, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var decoded testStruct
	err = json.Unmarshal(b, &decoded)
	if err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if original.Key.Key != decoded.Key.Key {
		t.Fatal("key round trip through struct failed")
	}
}
