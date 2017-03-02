package main

import (
	"strconv"

	"gitlab.home.mikenewswanger.com/golang/haproxy-configurator"
	"gitlab.home.mikenewswanger.com/golang/kubernetes-lightweight"
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
			if port.NodePort == 0 {
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
			if service.Metadata.Labels["service-router."+service.Metadata.Name+".listen-port"] != "" {
				var listenPort, _ = strconv.Atoi(service.Metadata.Labels["service-router."+service.Metadata.Name+".listen-port"])
				haproxyListenPort = uint16(listenPort)
			}

			var haproxyMode = "http"
			if service.Metadata.Labels["service-router."+service.Metadata.Name+".haproxy-mode"] != "" {
				haproxyMode = service.Metadata.Labels["service-router."+service.Metadata.Name+".haproxy-mode"]
			}

			var listenIP = "*"
			if service.Metadata.Labels["service-router."+service.Metadata.Name+".listen-ip"] != "" {
				listenIP = service.Metadata.Labels["service-router."+service.Metadata.Name+".listen-ip"]
			}

			var sslCertificate = ""
			var useSSL, exists = service.Metadata.Labels["service-router."+service.Metadata.Name+".use-ssl"]
			if !exists || useSSL != "false" {
				if service.Metadata.Labels["service-router."+port.Name+".ssl-certificate"] != "" {
					sslCertificate = "/etc/haproxy/ssl/" + service.Metadata.Labels["service-router."+port.Name+".ssl-certificate"]
				} else {
					sslCertificate = "/etc/haproxy/ssl/" + service.Metadata.Labels["service-router."+port.Name+".hostname"] + ".pem"
				}
			}

			var ipLabel = listenIP
			if listenIP == "*" {
				ipLabel = "all"
			}

			configurator.AddListener(
				"kube-service_"+ipLabel+"_"+strconv.Itoa(int(haproxyListenPort))+"_listen",
				listenIP,
				haproxyListenPort,
				haproxyMode,
				service.Metadata.Labels["service-router."+port.Name+".hostname"],
				sslCertificate,
				haproxyConfigurator.HaproxyBackend{
					Name:     "kube-service_" + service.Metadata.Namespace + "_" + service.Metadata.Name + "_" + port.Name + "_backend",
					Backends: targets,
				},
			)
		}
	}

	println(configurator.Render())
	publish(config.EtcdHost, config.HaproxyEtcdKey, configurator.Render())
}
