// +build none

package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

var docCmd = &cobra.Command{
	Use:   "doc",
	Short: `a command to generate documentation`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		os.MkdirAll(args[0], 0755)
		return doc.GenMarkdownTree(RootCmd, args[0])
	},
}

func init() {
	docCmd.Hidden = true
	RootCmd.AddCommand(docCmd)
}
