package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var cfgFile string

var RootCmd = &cobra.Command{
	Use:   "service-router-configurator",
	Short: "Dynamically configure load balancers for Kubernetes services",
	Long:  ``,
}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {

	// RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.tmp.yaml)")

	// RootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
