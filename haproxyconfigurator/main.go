package haproxyconfigurator

import (
	"strconv"
	"strings"

	"github.com/fatih/color"
	"github.com/sirupsen/logrus"
	"k8s.io/api/core/v1"

	"go.mikenewswanger.com/utilities/executil"
)

var kubeconfigFile string
var logger = logrus.New()

// SetLogger sets the logrus logger for use by the configurator
func SetLogger(l *logrus.Logger) {
	logger = l
}

// SetVerbosity sets the logrus logger for use by the configurator
func SetVerbosity(v uint8) {
	executil.SetVerbosity(v)
}

// Run polls the kubernetes configuration and builds out load balancer configurations based on the services in kubernetes
func Run(kubeconfigFilePath string, clusterFqdn string, etcdHostString string, etcdPathString string, shouldPublish bool) {
	kubeconfigFile = kubeconfigFilePath
	etcdHost = etcdHostString
	etcdPath = etcdPathString
	nodes, err := getAllKubernetesNodes()
	if err != nil {
		logger.Fatal(err)
	}
	services, err := getProxiedKubernetesServices()
	if err != nil {
		logger.Fatal(err)
	}
	buildHaproxyConfig(nodes, services, clusterFqdn, shouldPublish)
}

func buildHaproxyConfig(nodes map[string]string, services []v1.Service, clusterFqdn string, shouldPublish bool) {
	var configurator = HaproxyConfigurator{}
	configurator.Initialize()

	for _, service := range services {
		for _, port := range service.Spec.Ports {
			if port.NodePort == 0 {
				continue
			}

			serviceHostname := strings.Replace(service.Annotations["service-router."+port.Name+".hostname"], "CLUSTER_FQDN", clusterFqdn, 1)

			var targets = []HaproxyBackendTarget{}
			for hostname, ip := range nodes {
				targets = append(targets, HaproxyBackendTarget{
					Name: hostname,
					IP:   ip,
					Port: port.NodePort,
				})
			}

			var haproxyListenPort = uint16(443)
			if service.Annotations["service-router."+port.Name+".listen-port"] != "" {
				var listenPort, _ = strconv.Atoi(service.Annotations["service-router."+port.Name+".listen-port"])
				haproxyListenPort = uint16(listenPort)
			}

			var haproxyMode = "http"
			if service.Annotations["service-router."+port.Name+".haproxy-mode"] != "" {
				haproxyMode = service.Annotations["service-router."+port.Name+".haproxy-mode"]
			}

			var listenIP = "*"
			if service.Annotations["service-router."+port.Name+".listen-ip"] != "" {
				listenIP = service.Annotations["service-router."+port.Name+".listen-ip"]
			}

			// Default the service to use SSL with <hostname>.pem
			// SSL is enabled by default for HTTP
			var sslCertificate = ""
			useSSL, exists := service.Annotations["service-router."+port.Name+".use-ssl"]
			if (haproxyMode == "http" && !exists) || useSSL == "true" {
				if service.Annotations["service-router."+port.Name+".ssl-certificate"] != "" {
					sslCertificate = "/etc/haproxy/ssl/" + service.Annotations["service-router."+port.Name+".ssl-certificate"]
				} else {
					sslCertificate = "/etc/haproxy/ssl/" + serviceHostname + ".pem"
				}
			}

			// Default backends to use SSL if SSL is used on the front-end
			var backendsUseSSL = sslCertificate != ""
			backendsUseSSLLabel, exists := service.Annotations["service-router."+port.Name+".backends-use-ssl"]
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
			backendsVerifySSLLabel, exists := service.Annotations["service-router."+port.Name+".backends-verify-ssl"]
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
			backendBalanceMethodLabel, exists := service.Annotations["service-router."+port.Name+".backends-balance-method"]
			if exists {
				backendBalanceMethod = backendBalanceMethodLabel
			}

			var ipLabel = listenIP
			if listenIP == "*" {
				ipLabel = "all"
			}

			configurator.AddListener(
				HaproxyListenerConfig{
					Name:           "kube-service_" + ipLabel + "_" + strconv.Itoa(int(haproxyListenPort)) + "_listen",
					ListenIP:       listenIP,
					ListenPort:     haproxyListenPort,
					Mode:           haproxyMode,
					Hostname:       serviceHostname,
					SslCertificate: sslCertificate,
					Backend: HaproxyBackend{
						Name:          "kube-service_" + service.Namespace + "_" + service.Name + "_" + port.Name + "_backend",
						Backends:      targets,
						BalanceMethod: backendBalanceMethod,
						UseSSL:        backendsUseSSL,
						VerifySSL:     backendsVerifySSL,
					},
				},
			)
		}
	}

	color.White(configurator.Render())
	if shouldPublish {
		publish(configurator.Render())
	}
}
