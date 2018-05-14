package cmd

import (
	"github.com/spf13/cobra"

	"github.com/stackexchange/haproxy-kubefigurator/haproxyconfigurator"
)

// watchCmd represents the apply command
var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Watch for configuration changes, and save to etcd",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		haproxyconfigurator.Run(commandLineFlags.kubeconfig, commandLineFlags.clusterFqdn, commandLineFlags.etcdHost, commandLineFlags.etcdPath, true, false)
	},
}

func init() {
	RootCmd.AddCommand(watchCmd)
}
