package cli

import (
	"net"
	"testing"
)

func TestGetPrivateNet(t *testing.T) {
	n, err := getPrivateNet()
	if err != nil {
		t.Fatalf("getPrivateNet error: %v", err)
	}

	// Should be a 10.x.x.x /22 network
	if n.IPNet.IP[0] != 10 {
		t.Fatalf("expected 10.x.x.x, got %s", n.IPNet.IP)
	}

	ones, bits := n.IPNet.Mask.Size()
	if ones != 22 {
		t.Fatalf("expected /22, got /%d", ones)
	}
	if bits != 32 {
		t.Fatalf("expected 32 total bits, got %d", bits)
	}

	// Last byte should be 0 (network address)
	if n.IPNet.IP[3] != 0 {
		t.Fatalf("expected last octet 0, got %d", n.IPNet.IP[3])
	}

	// Third octet should be aligned to /22 boundary (multiple of 4)
	if n.IPNet.IP[2]%4 != 0 {
		t.Fatalf("expected third octet aligned to /22 (multiple of 4), got %d", n.IPNet.IP[2])
	}
}

func TestGetPrivateNetRandomness(t *testing.T) {
	// Generate multiple and check they're not all identical
	// (probabilistically this should pass)
	seen := make(map[string]bool)
	for i := 0; i < 10; i++ {
		n, err := getPrivateNet()
		if err != nil {
			t.Fatalf("getPrivateNet error: %v", err)
		}
		seen[n.IPNet.String()] = true
	}
	if len(seen) < 2 {
		t.Fatal("getPrivateNet should produce varying results")
	}
}

func TestGetULANet(t *testing.T) {
	n, err := getULANet()
	if err != nil {
		t.Fatalf("getULANet error: %v", err)
	}

	// Should be an fd00::/8 ULA prefix
	if n.IPNet.IP[0] != 0xfd {
		t.Fatalf("expected fd prefix, got %x", n.IPNet.IP[0])
	}

	// Second byte should be 0 (per the implementation)
	if n.IPNet.IP[1] != 0 {
		t.Fatalf("expected second byte 0, got %x", n.IPNet.IP[1])
	}

	ones, bits := n.IPNet.Mask.Size()
	if ones != 64 {
		t.Fatalf("expected /64, got /%d", ones)
	}
	if bits != 128 {
		t.Fatalf("expected 128 total bits, got %d", bits)
	}

	// Should be a valid IPv6 address (16 bytes)
	if len(n.IPNet.IP) != net.IPv6len {
		t.Fatalf("expected 16 byte IPv6, got %d bytes", len(n.IPNet.IP))
	}

	// Host portion (bytes 8-15) should be zero
	for i := 8; i < 16; i++ {
		if n.IPNet.IP[i] != 0 {
			t.Fatalf("expected zero in host portion at byte %d, got %x", i, n.IPNet.IP[i])
		}
	}
}

func TestGetULANetRandomness(t *testing.T) {
	seen := make(map[string]bool)
	for i := 0; i < 10; i++ {
		n, err := getULANet()
		if err != nil {
			t.Fatalf("getULANet error: %v", err)
		}
		seen[n.IPNet.String()] = true
	}
	if len(seen) < 2 {
		t.Fatal("getULANet should produce varying results")
	}
}

func TestGetULANetGlobalID(t *testing.T) {
	n, err := getULANet()
	if err != nil {
		t.Fatalf("getULANet error: %v", err)
	}

	// Bytes 2-6 are the 40-bit global ID (should be random, at least some non-zero)
	allZero := true
	for i := 2; i < 7; i++ {
		if n.IPNet.IP[i] != 0 {
			allZero = false
			break
		}
	}
	// Run a few times to avoid flaky false positive on all-zero random
	if allZero {
		n2, err := getULANet()
		if err != nil {
			t.Fatalf("getULANet error: %v", err)
		}
		allZero2 := true
		for i := 2; i < 7; i++ {
			if n2.IPNet.IP[i] != 0 {
				allZero2 = false
				break
			}
		}
		if allZero2 {
			t.Fatal("global ID should not be all zeros repeatedly")
		}
	}
}
