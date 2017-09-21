package cmd

import (
	"github.com/spf13/cobra"

	"go.mikenewswanger.com/service-router-configurator/haproxyconfigurator"
)

// viewCmd represents the view command
var viewCmd = &cobra.Command{
	Use:   "view",
	Short: "View the dynamically generated configuration",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		haproxyconfigurator.Run(false)
	},
}

func init() {
	RootCmd.AddCommand(viewCmd)
}
