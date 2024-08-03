package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/naggie/dsnet"
	"github.com/naggie/dsnet/cmd/cli"
	"github.com/naggie/dsnet/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	// Flags.
	owner       string
	description string
	confirm     bool

	// Commands.
	rootCmd = &cobra.Command{}

	initCmd = &cobra.Command{
		Use: "init",
		Short: fmt.Sprintf(
			"Create %s containing default configuration + new keys without loading. Edit to taste.",
			viper.GetString("config_file"),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			return cli.Init()
		},
	}

	upCmd = &cobra.Command{
		Use:   "up",
		Short: "Create the interface, run pre/post up, sync",
		RunE: func(cmd *cobra.Command, args []string) error {
			config, err := cli.LoadConfigFile()
			if err != nil {
				return fmt.Errorf("%w - failure to load config file", err)
			}
			server := cli.GetServer(config)
			if e := server.Up(); e != nil {
				return e
			}
			if e := utils.ShellOut(config.PostUp, "PostUp"); e != nil {
				return e
			}
			return nil
		},
	}

	downCmd = &cobra.Command{
		Use:   "down",
		Short: "Destroy the interface, run pre/post down",
		RunE: func(cmd *cobra.Command, args []string) error {
			config, err := cli.LoadConfigFile()
			if err != nil {
				return fmt.Errorf("%w - failure to load config file", err)
			}
			server := cli.GetServer(config)
			if e := server.DeleteLink(); e != nil {
				return e
			}
			if e := utils.ShellOut(config.PostDown, "PostDown"); e != nil {
				return e
			}
			return nil
		},
	}

	addCmd = &cobra.Command{
		Use:   "add <hostname>",
		Short: "Add a new peer + sync, optionally using a provided WireGuard private key",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			// Make sure we have the hostname
			if len(args) != 1 {
				return errors.New("Missing hostname argument")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) {
			privKey, err := cmd.PersistentFlags().GetBool("private-key")
			pubKey, err := cmd.PersistentFlags().GetBool("public-key")
			if err != nil {
				cli.ExitFail("%w - error processing key flag", err)
			}
			cli.Add(args[0], privKey, pubKey, owner, description, confirm)
		},
	}

	regenerateCmd = &cobra.Command{
		Use:   "regenerate [hostname]",
		Short: "Regenerate keys and config for peer",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.New("Missing hostname argument")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return cli.Regenerate(args[0], confirm)
		},
	}

	syncCmd = &cobra.Command{
		Use:   "sync",
		Short: fmt.Sprintf("Update wireguard configuration from %s after validating", viper.GetString("config_file")),
		RunE: func(cmd *cobra.Command, args []string) error {
			return cli.Sync()
		},
	}

	reportCmd = &cobra.Command{
		Use:   "report",
		Short: "Generate a JSON status report to stdout",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cli.GenerateReport()
		},
	}

	removeCmd = &cobra.Command{
		Use:   "remove [hostname]",
		Short: "Remove a peer by hostname provided as argument + sync",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			// Make sure we have the hostname
			if len(args) != 1 {
				return errors.New("Missing hostname argument")
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return cli.Remove(args[0], confirm)
		},
	}

	versionCmd = &cobra.Command{
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("dsnet version %s\ncommit %s\nbuilt %s", dsnet.VERSION, dsnet.GIT_COMMIT, dsnet.BUILD_DATE)
		},
		Use:   "version",
		Short: "Print version",
	}
)

func init() {
	// Flags.
	rootCmd.PersistentFlags().String("output", "wg-quick", "config file format: vyatta/wg-quick/nixos")
	addCmd.Flags().StringVar(&owner, "owner", "", "owner of the new peer")
	addCmd.Flags().StringVar(&description, "description", "", "description of the new peer")
	addCmd.Flags().BoolVar(&confirm, "confirm", false, "confirm")
	addCmd.PersistentFlags().BoolP("private-key", "r", false, "Accept user-supplied private key. If supplied, dsnet will generate a public key.")
	addCmd.PersistentFlags().BoolP("public-key", "u", false, "Accept user-supplied public key. If supplied, the user must add the private key to the generated config (or provide it with --private-key).")
	removeCmd.Flags().BoolVar(&confirm, "confirm", false, "confirm")
	regenerateCmd.Flags().BoolVar(&confirm, "confirm", false, "confirm")

	// Environment variable handling.
	viper.AutomaticEnv()
	viper.SetEnvPrefix("DSNET")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := viper.BindPFlag("output", rootCmd.PersistentFlags().Lookup("output")); err != nil {
		fmt.Fprintf(os.Stderr, "\033[31m%s\033[0m\n", err.Error())
		os.Exit(1)
	}

	viper.SetDefault("config_file", "/etc/dsnetconfig.json")
	viper.SetDefault("fallback_wg_bing", "wireguard-go")
	viper.SetDefault("listen_port", 51820)
	viper.SetDefault("interface_name", "dsnet")

	// if last handshake (different from keepalive, see https://www.wireguard.com/protocol/)
	viper.SetDefault("peer_timeout", 3*time.Minute)

	// when is a peer considered gone forever? (could remove)
	viper.SetDefault("peer_expiry", 28*time.Hour*24)

	// Adds subcommands.
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(regenerateCmd)
	rootCmd.AddCommand(syncCmd)
	rootCmd.AddCommand(reportCmd)
	rootCmd.AddCommand(removeCmd)
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(upCmd)
	rootCmd.AddCommand(downCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		// Because of side effects in viper, this gets printed twice
		fmt.Fprintf(os.Stderr, "\033[31m%s\033[0m\n", err.Error())
		os.Exit(1)
	}
	os.Exit(0)
}
