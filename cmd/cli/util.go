package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/naggie/dsnet/lib"
)

func check(e error, optMsg ...string) {
	if e != nil {
		if len(optMsg) > 0 {
			ExitFail("%s - %s", e, strings.Join(optMsg, " "))
		}
		ExitFail("%s", e)
	}
}

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

func ExitFail(format string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, "\033[31m"+format+"\033[0m\n", a...)
	os.Exit(1)
}

func MustPromptString(prompt string, required bool) string {
	reader := bufio.NewReader(os.Stdin)
	var text string
	var err error

	for text == "" {
		fmt.Fprintf(os.Stderr, "%s: ", prompt)
		text, err = reader.ReadString('\n')
		check(err)
		text = strings.TrimSpace(text)
	}
	return text
}

func ConfirmOrAbort(format string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, format+" [y/n] ", a...)

	reader := bufio.NewReader(os.Stdin)

	input, err := reader.ReadString('\n')
	if err != nil {
		panic(err)
	}

	if input == "y\n" {
		return
	} else {
		ExitFail("Aborted.")
	}
}
