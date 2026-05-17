package cli

import "testing"

func TestBytesToSI(t *testing.T) {
	tests := []struct {
		input uint64
		want  string
	}{
		{0, "0 B"},
		{999, "999 B"},
		{1000, "1.0 kB"},
		{1500, "1.5 kB"},
		{1000000, "1.0 MB"},
		{1000000000, "1.0 GB"},
		{1000000000000, "1.0 TB"},
		{2500000000, "2.5 GB"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := BytesToSI(tt.input); got != tt.want {
				t.Fatalf("BytesToSI(%d) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
