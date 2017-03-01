package main

import (
	"encoding/json"
	"io/ioutil"
	"net"
	"net/http"
	"sort"
	"strconv"
)

func main() {
	var config = configData{
		KubernetesMaster: "https://master.kubernetes.home.mikenewswanger.com:6443",
		HaproxyEtcdKey:   "/service-router/haproxy-config",
	}

	resp, err := http.Get(config.KubernetesMaster + "/api/v1/services")
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	var kubernetesServiceList = kubernetesServiceList{}
	err = json.Unmarshal(body, &kubernetesServiceList)
	if err != nil {
		panic(err)
	}

	resp, err = http.Get(config.KubernetesMaster + "/api/v1/nodes")
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	body, _ = ioutil.ReadAll(resp.Body)
	var kubernetesNodeList = kubernetesNodeList{}
	err = json.Unmarshal(body, &kubernetesNodeList)
	if err != nil {
		panic(err)
	}

	var haproxyDesiredConfig = haproxyConfig{
		ports: make(map[uint16]*haproxyListenService),
	}
	var kubelets = make(map[string]string)
	for _, nodeItem := range kubernetesNodeList.Items {
		ip, err := net.LookupHost(nodeItem.Metadata.Name)
		if err != nil {
			panic(err)
		}
		kubelets[nodeItem.Metadata.Name] = ip[0]
	}

	for _, serviceItem := range kubernetesServiceList.Items {
		for _, port := range serviceItem.Spec.Ports {
			if port.NodePort == 0 {
				continue
			}
			var haproxyPort = uint16(443)
			if serviceItem.Metadata.Labels["servicerouter."+serviceItem.Metadata.Name+".listenPort"] != "" {
				listenPort, _ := strconv.Atoi(serviceItem.Metadata.Labels["servicerouter."+serviceItem.Metadata.Name+".listenPort"])
				haproxyPort = uint16(listenPort)
			}
			_, exists := haproxyDesiredConfig.ports[haproxyPort]
			if !exists {
				haproxyDesiredConfig.ports[haproxyPort] = &haproxyListenService{
					mode:             "http",
					hostnameBackends: []haproxyBackend{},
				}
			}
			haproxyDesiredConfig.ports[haproxyPort].hostnameBackends = append(haproxyDesiredConfig.ports[haproxyPort].hostnameBackends, haproxyBackend{
				serviceName: serviceItem.Metadata.Namespace + "_" + serviceItem.Metadata.Name + "_" + port.Name,
				hostname:    serviceItem.Metadata.Labels["servicerouter."+port.Name+".hostname"],
				port:        port.NodePort,
			})
			if serviceItem.Metadata.Labels["servicerouter."+port.Name+".ssl-certificate"] != "" {
				haproxyDesiredConfig.ports[haproxyPort].sslCertificates = append(haproxyDesiredConfig.ports[haproxyPort].sslCertificates, serviceItem.Metadata.Labels["servicerouter."+port.Name+".ssl-certificate"])
			} else {
				haproxyDesiredConfig.ports[haproxyPort].sslCertificates = append(haproxyDesiredConfig.ports[haproxyPort].sslCertificates, serviceItem.Metadata.Labels["servicerouter."+port.Name+".hostname"]+".pem")
			}
			haproxyDesiredConfig.ports[haproxyPort].useSSL = true
		}
	}

	var haproxyConfigFileContents = haproxyDesiredConfig.render(kubelets)
	println(haproxyConfigFileContents)
	publish(config, haproxyConfigFileContents)
}

type configData struct {
	KubernetesMaster string
	HaproxyEtcdKey   string
}

type haproxyConfig struct {
	ports map[uint16]*haproxyListenService
}

type haproxyListenService struct {
	mode             string
	sslCertificates  []string
	hostnameBackends []haproxyBackend
	useSSL           bool
}

type haproxyBackend struct {
	hostname    string
	serviceName string
	port        uint16
}

func (h *haproxyConfig) render(proxyHosts map[string]string) string {
	var config = ""
	for portNumber, haproxyListener := range h.ports {
		var portString = string(strconv.Itoa(int(portNumber)))
		config += "frontend kube_services_https_" + portString + "\n"
		config += "    bind *:" + portString
		if haproxyListener.useSSL == true {
			config += " ssl"
			sort.Strings(haproxyListener.sslCertificates)
			var previous = ""
			for _, sslCertificate := range haproxyListener.sslCertificates {
				if sslCertificate != previous {
					config += " crt /etc/haproxy/ssl/" + sslCertificate
					previous = sslCertificate
				}
			}
		}
		config += "\n"
		config += "    mode " + haproxyListener.mode + "\n"
		if haproxyListener.mode == "http" {
			config += "    reqadd X-Forwarded-Proto:\\ https\n"
		}
		for _, backend := range haproxyListener.hostnameBackends {
			config += "    use_backend kube-service_" + backend.serviceName + " if { hdr(host) -i " + backend.hostname + " }\n"
		}
		config += "\n"
		for _, backend := range haproxyListener.hostnameBackends {
			config += "backend kube-service_" + backend.serviceName + "\n"
			config += "    mode " + haproxyListener.mode + "\n"
			config += "    balance roundrobin\n"
			for hostname, ip := range proxyHosts {
				config += "    server " + hostname + " " + ip + ":" + strconv.Itoa(int(backend.port)) + " check\n"
			}
			config += "\n"
		}
	}
	return config
}
