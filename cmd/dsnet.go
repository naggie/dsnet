package main

import (
	"os"
	"flag"
	"fmt"
	"github.com/naggie/dsnet"
)

func main() {
	addCmd := flag.NewFlagSet("add", flag.ExitOnError)


	switch os.Args[1] {
		case "init":
			dsnet.Init()

		case "up":

		case "add":

		case "report":

		case "down":

		default:
			help();
	}
}

func help() {
	fmt.Printf(`Usage: dsnet <cmd>

Available commands:

	init : Create %s containing default configuration + new keys without loading. Edit to taste.
	add  : Generate configuration for a new peer. (Send with passworded ffsend)
	sync : Synchronise wireguard configuration with %s, creating and activating interface if necessary

To remove an interface or bring it down, use standard tools such as iproute2.
To modify or remove peers, edit %s and then run sync.

`, dstask.CONFIG_FILE, dstask.CONFIG_FILE, dstask.CONFIG_FILE)
}
