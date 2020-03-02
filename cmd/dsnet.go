package main

import (
	"os"
	"flag"
	"fmt"
	"github.com/naggie/dsnet"
)

func main() {
	var cmd string

	addCmd := flag.NewFlagSet("add", flag.ExitOnError)



	if len(os.Args) == 1 {
		cmd = "help"
	} else {
		cmd = os.Args[1]
	}

	switch cmd {
	case "init":
		dsnet.Init()

	case "up":

	case "add":
		dsnet.Add()

	case "report":

	case "down":

	default:
		help()
	}
}

func help() {
	fmt.Printf(`dsnet is a simple tool to manage a wireguard VPN.

Usage: dsnet <cmd>

Available commands:

	init   : Create %s containing default configuration + new keys without loading. Edit to taste.
	add    : Generate configuration for a new peer, adding to %s. Send with passworded ffsend.
	sync   : Synchronise wireguard configuration with %s, creating and activating interface if necessary.
	report : Generate a JSON status report to the location configured in %s.

To remove an interface or bring it down, use standard tools such as iproute2.
To modify or remove peers, edit %s and then run sync.

`, dsnet.CONFIG_FILE, dsnet.CONFIG_FILE, dsnet.CONFIG_FILE, dsnet.CONFIG_FILE, dsnet.CONFIG_FILE)
}
