package cmd

import (
	"github.com/spf13/cobra"

	"go.mikenewswanger.com/proxy-konfigurator/haproxyconfigurator"
)

// viewCmd represents the view command
var viewCmd = &cobra.Command{
	Use:   "view",
	Short: "View the dynamically generated configuration",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		haproxyconfigurator.Run(commandLineFlags.kubernetesContext, commandLineFlags.clusterFqdn, commandLineFlags.etcdHost, commandLineFlags.etcdPath, false)
	},
}

func init() {
	RootCmd.AddCommand(viewCmd)
}
