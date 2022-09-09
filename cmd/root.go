package main

import (
	"errors"
	"fmt"
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
		Run: func(cmd *cobra.Command, args []string) {
			cli.Init()
		},
	}

	upCmd = &cobra.Command{
		Use:   "up",
		Short: "Create the interface, run pre/post up, sync",
		Run: func(cmd *cobra.Command, args []string) {
			config := cli.MustLoadConfigFile()
			server := cli.GetServer(config)
			server.Up()
			utils.ShellOut(config.PostUp, "PostUp")
		},
	}

	downCmd = &cobra.Command{
		Use:   "down",
		Short: "Destroy the interface, run pre/post down",
		Run: func(cmd *cobra.Command, args []string) {
			config := cli.MustLoadConfigFile()
			server := cli.GetServer(config)
			server.DeleteLink()
			utils.ShellOut(config.PostDown, "PostDown")
		},
	}

	addCmd = &cobra.Command{
		Use:   "add [hostname]",
		Short: "Add a new peer + sync",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			// Make sure we have the hostname
			if len(args) != 1 {
				return errors.New("Missing hostname argument")
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			cli.Add(args[0], owner, description, confirm)
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
		Run: func(cmd *cobra.Command, args []string) {
			cli.Regenerate(args[0], confirm)
		},
	}

	syncCmd = &cobra.Command{
		Use:   "sync",
		Short: fmt.Sprintf("Update wireguard configuration from %s after validating", viper.GetString("config_file")),
		Run: func(cmd *cobra.Command, args []string) {
			cli.Sync()
		},
	}

	reportCmd = &cobra.Command{
		Use:   "report",
		Short: "Generate a JSON status report to stdout",
		Run: func(cmd *cobra.Command, args []string) {
			cli.GenerateReport()
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
		Run: func(cmd *cobra.Command, args []string) {
			cli.Remove(args[0], confirm)
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
	removeCmd.Flags().BoolVar(&confirm, "confirm", false, "confirm")

	// Environment variable handling.
	viper.AutomaticEnv()
	viper.SetEnvPrefix("DSNET")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := viper.BindPFlag("output", rootCmd.PersistentFlags().Lookup("output")); err != nil {
		cli.ExitFail(err.Error())
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
		cli.ExitFail(err.Error())
	}
}
