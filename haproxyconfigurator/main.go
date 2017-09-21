package haproxyconfigurator

import (
	"strconv"

	"github.com/sirupsen/logrus"
)

var kubernetesMaster = "https://master.kubernetes.home.mikenewswanger.com:6443"
var logger = logrus.New()

// SetLogger sets the logrus logger for use by the configurator
func SetLogger(l *logrus.Logger) {
	logger = l
}

// Run polls the kubernetes configuration and builds out load balancer configurations based on the services in kubernetes
func Run(shouldPublish bool) {
	buildHaproxyConfig(getAllKubernetesNodes(), GetAllKubernetesServices(), shouldPublish)
}

func buildHaproxyConfig(nodes map[string]string, services KubernetesServiceList, shouldPublish bool) {
	var configurator = HaproxyConfigurator{}
	configurator.Initialize()

	for _, service := range services.Items {
		for _, port := range service.Spec.Ports {
			if service.Metadata.Labels["service-router.enabled"] != "yes" || port.NodePort == 0 {
				continue
			}

			var targets = []HaproxyBackendTarget{}
			for hostname, ip := range nodes {
				targets = append(targets, HaproxyBackendTarget{
					Name: hostname,
					IP:   ip,
					Port: port.NodePort,
				})
			}

			var haproxyListenPort = uint16(443)
			if service.Metadata.Annotations["service-router."+port.Name+".listen-port"] != "" {
				var listenPort, _ = strconv.Atoi(service.Metadata.Annotations["service-router."+port.Name+".listen-port"])
				haproxyListenPort = uint16(listenPort)
			}

			var haproxyMode = "http"
			if service.Metadata.Annotations["service-router."+port.Name+".haproxy-mode"] != "" {
				haproxyMode = service.Metadata.Annotations["service-router."+port.Name+".haproxy-mode"]
			}

			var listenIP = "*"
			if service.Metadata.Annotations["service-router."+port.Name+".listen-ip"] != "" {
				listenIP = service.Metadata.Annotations["service-router."+port.Name+".listen-ip"]
			}

			// Default the service to use SSL with <hostname>.pem
			// SSL is enabled by default for HTTP
			var sslCertificate = ""
			useSSL, exists := service.Metadata.Annotations["service-router."+port.Name+".use-ssl"]
			if (haproxyMode == "http" && !exists) || useSSL == "true" {
				if service.Metadata.Annotations["service-router."+port.Name+".ssl-certificate"] != "" {
					sslCertificate = "/etc/haproxy/ssl/" + service.Metadata.Annotations["service-router."+port.Name+".ssl-certificate"]
				} else {
					sslCertificate = "/etc/haproxy/ssl/" + service.Metadata.Annotations["service-router."+port.Name+".hostname"] + ".pem"
				}
			}

			// Default backends to use SSL if SSL is used on the front-end
			var backendsUseSSL = sslCertificate != ""
			backendsUseSSLLabel, exists := service.Metadata.Annotations["service-router."+port.Name+".backends-use-ssl"]
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
			backendsVerifySSLLabel, exists := service.Metadata.Annotations["service-router."+port.Name+".backends-verify-ssl"]
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
			backendBalanceMethodLabel, exists := service.Metadata.Annotations["service-router."+port.Name+".backends-balance-method"]
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
					Hostname:       service.Metadata.Annotations["service-router."+port.Name+".hostname"],
					SslCertificate: sslCertificate,
					Backend: HaproxyBackend{
						Name:          "kube-service_" + service.Metadata.Namespace + "_" + service.Metadata.Name + "_" + port.Name + "_backend",
						Backends:      targets,
						BalanceMethod: backendBalanceMethod,
						UseSSL:        backendsUseSSL,
						VerifySSL:     backendsVerifySSL,
					},
				},
			)
		}
	}

	println(configurator.Render())
	if shouldPublish {
		publish(configurator.Render())
	}
}
