package main

import (
	"os"
	"fmt"
	"github.com/naggie/dsnet"
)

func main() {
	var cmd string

	if len(os.Args) == 1 {
		cmd = "help"
	} else {
		cmd = os.Args[1]
	}

	switch cmd {
	case "init":
		dsnet.Init()

	case "add":
		dsnet.Add()

	case "up":
		dsnet.Up()

	case "sync":
		dsnet.Sync()

	case "report":
		dsnet.Report()

	case "down":
		dsnet.Down()

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
	up     : Create the interface, run pre/post up, sync
	sync   : Update wireguard configuration with %s, adding/removing peers after validating matching config
	report : Generate a JSON status report to the location configured in %s.
	down   : Destroy the interface, run pre/post down

To remove an interface or bring it down, use standard tools such as iproute2.
To modify or remove peers, edit %s and then run sync.

`, dsnet.CONFIG_FILE, dsnet.CONFIG_FILE, dsnet.CONFIG_FILE, dsnet.CONFIG_FILE, dsnet.CONFIG_FILE)
}
