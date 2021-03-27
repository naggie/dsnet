package main

import (
	"errors"
	"fmt"
	"strings"

	"github.com/naggie/dsnet"
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
		Run: func(cmd *cobra.Command, args []string) {
			dsnet.Init()
		},
		Use: "init",
		Short: fmt.Sprintf(
			"Create %s containing default configuration + new keys without loading. Edit to taste.",
			dsnet.CONFIG_FILE,
		),
	}

	addCmd = &cobra.Command{
		PreRunE: func(cmd *cobra.Command, args []string) error {
			// Make sure we have the hostname
			if len(args) != 1 {
				return errors.New("Missing hostname argument")
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			dsnet.Add(args[0], owner, description, confirm)
		},
		Use:   "add [hostname]",
		Short: "Add a new peer + sync",
	}

	regenerateCmd = &cobra.Command{
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.New("Missing hostname argument")
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			dsnet.Regenerate(args[0], confirm)
		},
		Use:   "regenerate [hostname]",
		Short: "Regenerate keys and config for peer",
	}

	upCmd = &cobra.Command{
		Run: func(cmd *cobra.Command, args []string) {
			dsnet.Up()
		},
		Use:   "up",
		Short: "Create the interface, run pre/post up, sync",
	}

	syncCmd = &cobra.Command{
		Run: func(cmd *cobra.Command, args []string) {
			dsnet.Sync()
		},
		Use:   "sync",
		Short: fmt.Sprintf("Update wireguard configuration from %s after validating", dsnet.CONFIG_FILE),
	}

	reportCmd = &cobra.Command{
		Run: func(cmd *cobra.Command, args []string) {
			dsnet.Report()
		},
		Use:   "report",
		Short: fmt.Sprintf("Generate a JSON status report to the location configured in %s.", dsnet.CONFIG_FILE),
	}

	removeCmd = &cobra.Command{
		PreRunE: func(cmd *cobra.Command, args []string) error {
			// Make sure we have the hostname
			if len(args) != 1 {
				return errors.New("Missing hostname argument")
			}
			viper.Set("hostname", args[0])

			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			dsnet.Remove(args[0], confirm)
		},
		Use:   "remove [hostname]",
		Short: "Remove a peer by hostname provided as argument + sync",
	}

	downCmd = &cobra.Command{
		Run: func(cmd *cobra.Command, args []string) {
			dsnet.Down()
		},
		Use:   "down",
		Short: "Destroy the interface, run pre/post down",
	}

	versionCmd = &cobra.Command{
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("dsnet version %s\ncommit %s\nbuilt %s", dsnet.VERSION, dsnet.GIT_COMMIT, dsnet.BUILD_DATE)
		},
		Use:   "version",
		Short: "Print version",
	}
)

func main() {
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
		dsnet.ExitFail(err.Error())
	}

	// Adds subcommands.
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(regenerateCmd)
	rootCmd.AddCommand(upCmd)
	rootCmd.AddCommand(syncCmd)
	rootCmd.AddCommand(reportCmd)
	rootCmd.AddCommand(removeCmd)
	rootCmd.AddCommand(downCmd)
	rootCmd.AddCommand(versionCmd)

	if err := rootCmd.Execute(); err != nil {
		dsnet.ExitFail(err.Error())
	}
}
