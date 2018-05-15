package haproxyconfigurator

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/fatih/color"
	"github.com/sirupsen/logrus"
	"k8s.io/api/core/v1"
)

var (
	kubeconfigFile string
	logger         = logrus.New()
)

// SetLogger sets the logrus logger for use by the configurator
func SetLogger(l *logrus.Logger) {
	logger = l
}

// Run polls the kubernetes configuration and builds out load balancer configurations based on the services in kubernetes
func Run(kubeconfigFilePath string, clusterName string, etcdOptions EtcdOptions, watch bool, shouldPublish bool) {
	kubeconfigFile = kubeconfigFilePath
	lastConfigPushed := ""
	for {
		nodes, err := getAllKubernetesNodes()
		if err != nil {
			logger.Fatal(err)
		}
		services, err := getProxiedKubernetesServices()
		if err != nil {
			logger.Fatal(err)
		}
		config, err := buildHaproxyConfig(nodes, services, clusterName, shouldPublish)
		if err != nil {
			logger.Fatal(err)
		}
		anyChange := config != lastConfigPushed
		if anyChange {
			color.White(config)
			if shouldPublish {
				publish(config, etcdOptions)
				lastConfigPushed = config
			}
		} else {
			color.Green("No change")
		}
		if !watch {
			break
		}
		// Look for changes in either services or nodes
		waitForChanges()
	}
}

type servicePortWrapper v1.ServicePort

func (s servicePortWrapper) annoName(name string) string {
	return fmt.Sprintf("haproxy-kubefigurator.%s.%s", s.Name, name)
}

type serviceWrapper v1.Service

func (s serviceWrapper) anno(p servicePortWrapper, name string) string {
	return s.Annotations[p.annoName(name)]
}
func (s serviceWrapper) annoExists(p servicePortWrapper, name string) (string, bool) {
	str, ok := s.Annotations[p.annoName(name)]
	return str, ok
}

func buildHaproxyConfig(nodes map[string]string, services []v1.Service, clusterName string, shouldPublish bool) (string, error) {
	var configurator = HaproxyConfigurator{}
	configurator.Initialize()

	for _, svc := range services {
		service := serviceWrapper(svc)
		for _, p := range service.Spec.Ports {
			port := servicePortWrapper(p)
			if port.NodePort == 0 {
				continue
			}

			serviceHostname := strings.Replace(service.anno(port, "hostname"), "CLUSTER", clusterName, 1)

			var targets = []HaproxyBackendTarget{}
			for hostname, ip := range nodes {
				targets = append(targets, HaproxyBackendTarget{
					Name: hostname,
					IP:   ip,
					Port: port.NodePort,
				})
			}

			var haproxyListenPort = uint16(443)
			if lp := service.anno(port, "listen-port"); lp != "" {
				var listenPort, _ = strconv.Atoi(lp)
				haproxyListenPort = uint16(listenPort)
			}

			var haproxyMode = "http"
			if mode := service.anno(port, "haproxy-mode"); mode != "" {
				haproxyMode = mode
			}

			var listenIP = "*"
			if lIP := service.anno(port, "listen-ip"); lIP != "" {
				listenIP = lIP
			}

			// Default the service to use SSL with <hostname>.pem
			// SSL is enabled by default for HTTP
			var sslCertificate = ""
			useSSL, exists := service.annoExists(port, "use-ssl")
			if (haproxyMode == "http" && !exists) || useSSL == "true" {
				if cert := service.anno(port, "ssl-certificate"); cert != "" {
					sslCertificate = "/etc/haproxy/ssl/" + cert
				} else {
					sslCertificate = "/etc/haproxy/ssl/" + serviceHostname + ".pem"
				}
			}

			// Default backends to use SSL if SSL is used on the front-end
			var backendsUseSSL = sslCertificate != ""
			backendsUseSSLLabel, exists := service.annoExists(port, "backends-use-ssl")
			if exists {
				if backendsUseSSLLabel == "false" {
					backendsUseSSL = false
				}
				if backendsUseSSLLabel == "true" {
					backendsUseSSL = true
				}
			}

			// Default backends to use SSL if SSL is used on the front-end
			var backendsVerifySSL = false
			backendsVerifySSLLabel, exists := service.annoExists(port, "backends-verify-ssl")
			if exists {
				if backendsVerifySSLLabel == "false" {
					backendsVerifySSL = false
				}
				if backendsVerifySSLLabel == "true" {
					backendsVerifySSL = true
				}
			}

			// Default balance method to roundrobin
			var backendBalanceMethod = "roundrobin"
			backendBalanceMethodLabel, exists := service.annoExists(port, "backends-balance-method")
			if exists {
				backendBalanceMethod = backendBalanceMethodLabel
			}

			var ipLabel = listenIP
			if listenIP == "*" {
				ipLabel = "all"
			}

			configurator.AddListener(
				HaproxyListenerConfig{
					Name:           "k8s-service_" + ipLabel + "_" + strconv.Itoa(int(haproxyListenPort)) + "_listen",
					ListenIP:       listenIP,
					ListenPort:     haproxyListenPort,
					Mode:           haproxyMode,
					Hostname:       serviceHostname,
					SslCertificate: sslCertificate,
					Backend: HaproxyBackend{
						Name:          "k8s-service_" + service.Namespace + "_" + service.Name + "_" + port.Name + "_backend",
						Backends:      targets,
						BalanceMethod: backendBalanceMethod,
						UseSSL:        backendsUseSSL,
						VerifySSL:     backendsVerifySSL,
					},
				},
			)
		}
	}

	return configurator.Render(), nil
}
