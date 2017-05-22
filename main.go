package main

import (
	"strconv"

	"go.mikenewswanger.com/service-router-configurator/haproxy-configurator"
	"go.mikenewswanger.com/service-router-configurator/kubernetes-lightweight"
)

type configData struct {
	KubernetesMaster string
	EtcdHost         string
	HaproxyEtcdKey   string
}

func main() {
	var config = configData{
		KubernetesMaster: "https://master.kubernetes.home.mikenewswanger.com:6443",
		EtcdHost:         "http://etcd.kubernetes.home.mikenewswanger.com:2379",
		HaproxyEtcdKey:   "/service-router/haproxy-config",
	}

	var k = kube.Kube{}
	k.Initialize(config.KubernetesMaster)
	buildHaproxyConfig(config, k.GetAllNodes().GetIPs(), k.GetAllServices())
}

func buildHaproxyConfig(config configData, nodes map[string]string, services kube.KubernetesServiceList) {
	var configurator = haproxyConfigurator.HaproxyConfigurator{}
	configurator.Initialize()

	for _, service := range services.Items {
		for _, port := range service.Spec.Ports {
			if service.Metadata.Labels["service-router.enabled"] != "yes" || port.NodePort == 0 {
				continue
			}

			var targets = []haproxyConfigurator.HaproxyBackendTarget{}
			for hostname, ip := range nodes {
				targets = append(targets, haproxyConfigurator.HaproxyBackendTarget{
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
				haproxyConfigurator.HaproxyListenerConfig{
					Name:           "kube-service_" + ipLabel + "_" + strconv.Itoa(int(haproxyListenPort)) + "_listen",
					ListenIP:       listenIP,
					ListenPort:     haproxyListenPort,
					Mode:           haproxyMode,
					Hostname:       service.Metadata.Annotations["service-router."+port.Name+".hostname"],
					SslCertificate: sslCertificate,
					Backend: haproxyConfigurator.HaproxyBackend{
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
	publish(config.EtcdHost, config.HaproxyEtcdKey, configurator.Render())
}
