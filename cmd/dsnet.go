package main

import (
	"fmt"
	"os"

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

	case "remove":
		dsnet.Remove()

	case "down":
		dsnet.Down()

	default:
		help()
	}
}

func help() {
	fmt.Printf(`dsnet is a simple tool to manage a centralised wireguard VPN.

Usage: dsnet <cmd>

Available commands:

	init   : Create %s containing default configuration + new keys without loading. Edit to taste.
	add    : Add a new peer + sync
	up     : Create the interface, run pre/post up, sync
	report : Generate a JSON status report to the location configured in %s.
	remove : Remove a peer by hostname provided as argument + sync
	down   : Destroy the interface, run pre/post down
	sync   : Update wireguard configuration from %s after validating

`, dsnet.CONFIG_FILE, dsnet.CONFIG_FILE, dsnet.CONFIG_FILE)
}
