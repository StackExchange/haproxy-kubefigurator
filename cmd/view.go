package cmd

import (
	"github.com/spf13/cobra"

	"github.com/StackExchange/haproxy-kubefigurator/haproxyconfigurator"
)

// viewCmd represents the view command
var viewCmd = &cobra.Command{
	Use:   "view",
	Short: "View the dynamically generated configuration",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		haproxyconfigurator.Run(commandLineFlags.kubeconfig, commandLineFlags.clusterName, commandLineFlags.haproxyConfig, false, false, "")
	},
}

func init() {
	RootCmd.AddCommand(viewCmd)
}
