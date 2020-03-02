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
	hostname := addCmd.String("hostname", "", "Hostname of device")
	owner := addCmd.String("owner", "", "Username of owner of device")
	description := addCmd.String("description", "", "Information about device")
	publicKey := addCmd.String("publicKey", "", "Optional existing public key of device")



	if len(os.Args) == 1 {
		cmd = "help"
	} else {
		cmd = os.Args[1]
	}

	switch cmd {
	case "init":
		dsnet.Init()

	case "add":
		addCmd.PrintDefaults()
		addCmd.Parse(os.Args[2:])
		dsnet.Add(*hostname, *owner, *description, *publicKey)

	case "up":

	case "update":

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
	up     : Create the interface, run pre/post up, update
	update : Update wireguard configuration with %s, adding/removing peers after validating matching config
	report : Generate a JSON status report to the location configured in %s.
	down   : Destroy the interface, run pre/post down

To remove an interface or bring it down, use standard tools such as iproute2.
To modify or remove peers, edit %s and then run sync.

`, dsnet.CONFIG_FILE, dsnet.CONFIG_FILE, dsnet.CONFIG_FILE, dsnet.CONFIG_FILE, dsnet.CONFIG_FILE)
}
