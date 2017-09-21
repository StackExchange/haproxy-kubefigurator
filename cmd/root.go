package cmd

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"go.mikenewswanger.com/service-router-configurator/haproxyconfigurator"
)

var commandLineFlags = struct {
	etcdHost          string
	etcdPath          string
	kubernetesContext string
	verbosity         int
}{}
var logger = logrus.New()

var RootCmd = &cobra.Command{
	Use:   "service-router-configurator",
	Short: "Dynamically configure load balancers for Kubernetes services",
	Long:  ``,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		switch commandLineFlags.verbosity {
		case 0:
			logger.Level = logrus.ErrorLevel
			break
		case 1:
			logger.Level = logrus.WarnLevel
			break
		case 2:
			fallthrough
		case 3:
			logger.Level = logrus.InfoLevel
			break
		default:
			logger.Level = logrus.DebugLevel
			break
		}

		haproxyconfigurator.SetLogger(logger)
		haproxyconfigurator.SetVerbosity(uint8(commandLineFlags.verbosity))

		logger.Debug("Pre-run complete")
	},
}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	RootCmd.PersistentFlags().CountVarP(&commandLineFlags.verbosity, "verbosity", "v", "Output verbosity")
	RootCmd.PersistentFlags().StringVarP(&commandLineFlags.kubernetesContext, "kubectl-context", "", "", "Kubectl Context")
	RootCmd.PersistentFlags().StringVarP(&commandLineFlags.etcdHost, "etcd-host", "", "", "etcd Host")
	RootCmd.PersistentFlags().StringVarP(&commandLineFlags.etcdPath, "etcd-path", "", "", "etcd Path")
}
