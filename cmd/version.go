package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var Version = "0.0.0"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "prints this shipyardctl version",
	Run: func(cmd *cobra.Command, args []string) {
    fmt.Printf("Version: %s\n", Version)
  },
}

func init() {
	RootCmd.AddCommand(versionCmd)
}
