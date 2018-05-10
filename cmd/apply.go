package cmd

import (
	"github.com/spf13/cobra"

	"go.mikenewswanger.com/proxy-konfigurator/haproxyconfigurator"
)

// applyCmd represents the apply command
var applyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Apply the dynamic configuration to the load balancers",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		haproxyconfigurator.Run(commandLineFlags.kubeconfig, commandLineFlags.clusterFqdn, commandLineFlags.etcdHost, commandLineFlags.etcdPath, true)
	},
}

func init() {
	RootCmd.AddCommand(applyCmd)
}
