package cmd

import (
	"github.com/paulgmiller/wg-sync/nethelpers"
	"github.com/spf13/cobra"
)

// addCmd represents the add command
var debugCmd = &cobra.Command{
	Use:   "debug",
	Short: "dump some info",
	Long:  `all sorts`,
	RunE: func(cmd *cobra.Command, args []string) error {
		_, err := nethelpers.FindPublicIp()
		return err
	},
}

var server *bool

func init() {
	rootCmd.AddCommand(debugCmd)
}
