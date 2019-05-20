package cmd

import (
	"log"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "flakbase",
	Short: "Flakbase is a realtime database server compatible with Firebase APIsxs",
	Args:  cobra.NoArgs,
}

// Main defines the entry for flakbase.
func Main() {
	rootCmd.AddCommand(cmdServe)

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
