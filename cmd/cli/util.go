package cli

// FIXME every function in this file has public scope, but only private references

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/naggie/dsnet/lib"
)

func jsonPeerToDsnetPeer(peers []PeerConfig) []lib.Peer {
	libPeers := make([]lib.Peer, 0, len(peers))
	for _, p := range peers {
		libPeers = append(libPeers, lib.Peer{
			Hostname:     p.Hostname,
			Owner:        p.Owner,
			Description:  p.Description,
			IP:           p.IP,
			IP6:          p.IP6,
			Added:        p.Added,
			PublicKey:    p.PublicKey,
			PrivateKey:   p.PrivateKey,
			PresharedKey: p.PresharedKey,
			Networks:     p.Networks,
		})
	}
	return libPeers
}

func PromptString(prompt string, required bool) (string, error) {
	reader := bufio.NewReader(os.Stdin)
	var text string
	var err error

	for text == "" {
		fmt.Fprintf(os.Stderr, "%s: ", prompt)
		text, err = reader.ReadString('\n')
		if err != nil {
			return "", fmt.Errorf("%w - error getting input", err)
		}
		text = strings.TrimSpace(text)
	}
	return text, nil
}

func ConfirmOrAbort(format string, a ...any) error {
	fmt.Fprintf(os.Stderr, format+" [y/n] ", a...)

	reader := bufio.NewReader(os.Stdin)

	input, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("%w - failed to read confirmation", err)
	}

	if input == "y\n" {
		return nil
	}
	return errors.New("aborted")
}

func BytesToSI(b uint64) string {
	const unit = 1000
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB",
		float64(b)/float64(div), "kMGTPE"[exp])
}
