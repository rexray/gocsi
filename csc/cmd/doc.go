// +build none

package cmd

import (
	"os"

	"github.com/spf13/cobra"
	cdoc "github.com/spf13/cobra/doc"
)

var doc struct {
	docType docTypeArg
}

var docCmd = &cobra.Command{
	Use:   "doc",
	Short: `generates documentation`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := os.MkdirAll(args[0], 0755); err != nil {
			return err
		}
		switch doc.docType.val {
		case "md":
			return cdoc.GenMarkdownTree(RootCmd, args[0])
		case "man":
			return cdoc.GenManTree(RootCmd, &cdoc.GenManHeader{
				Title:   "csc",
				Section: "3",
			}, args[0])
		case "rst":
			return cdoc.GenReSTTree(RootCmd, args[0])
		}
		return nil
	},
}

func init() {
	docCmd.Hidden = true
	RootCmd.AddCommand(docCmd)
	docCmd.Flags().Var(
		&doc.docType,
		"type",
		`the type of documentation to generate`)
}
