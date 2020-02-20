package main

import (
	"os"
	"flag"
	"github.com/naggie/dsnet"
)

func main() {
	addCmd := flag.NewFlagSet("add", flag.ExitOnError)


	switch os.Args[1] {
		case "init":

		case "up":

		case "add":

		case "report":

		case "down":

		default:
			help();
	}
}
