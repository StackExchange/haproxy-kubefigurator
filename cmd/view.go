package cmd

import (
	"github.com/spf13/cobra"

	"github.com/stackexchange/haproxy-kubefigurator/haproxyconfigurator"
)

// viewCmd represents the view command
var viewCmd = &cobra.Command{
	Use:   "view",
	Short: "View the dynamically generated configuration",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		haproxyconfigurator.Run(commandLineFlags.kubeconfig, commandLineFlags.clusterName, commandLineFlags.etcdOptions, false, false)
	},
}

func init() {
	RootCmd.AddCommand(viewCmd)
}
