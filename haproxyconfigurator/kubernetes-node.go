package haproxyconfigurator

import (
	"net"
	"strings"

	"github.com/ghodss/yaml"
	"go.mikenewswanger.com/utilities/executil"
)

// getAllKubernetesNodes loads the nodes in the target kubernetes cluster
func getAllKubernetesNodes() map[string]string {

	c := executil.Command{
		Name:       "Get Kubernetes Nodes",
		Executable: "kubectl",
		Arguments:  append(getKubernetesContextOptions(), "get", "nodes", "-o", "custom-columns=Name:.metadata.name"),
	}
	c.Run()
	nodes := strings.Split(c.GetStdout(), "\n")
	if len(nodes) < 2 {
		logger.Fatal("No kubernetes nodes were found")
	}
	return getKubernetesNodeIPs(nodes[1:])
}

// getKubernetesNodeIPs returns a map of node names to their IPs
func getKubernetesNodeIPs(nodes []string) map[string]string {
	nodesWithIPs := make(map[string]string)
	for _, n := range nodes {
		if n == "" {
			continue
		}
		ip, err := net.LookupHost(n)
		if err != nil {
			panic(err)
		}
		nodesWithIPs[n] = ip[0]
	}
	return nodesWithIPs
}

func getKubernetesContextOptions() []string {
	return []string{}
}

// KubernetesServiceList contains a list of kubernetes services
type KubernetesServiceList struct {
	Items []kubernetesServiceItem `json:"items"`
}

type kubernetesServiceItem struct {
	Metadata kubernetesServiceMetadata `json:"metadata"`
	Spec     kubernetesServiceSpec     `json:"spec"`
}

type kubernetesServiceMetadata struct {
	Name        string            `json:"name"`
	Namespace   string            `json:"namespace"`
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
}

type kubernetesServiceSpec struct {
	Ports []kubernetesServicePort `json:"ports"`
}

type kubernetesServicePort struct {
	Name       string `json:"name"`
	Protocol   string `json:"protocol"`
	NodePort   uint16 `json:"nodePort"`
	Port       uint16 `json:"port"`
	TargetPort uint16 `json:"targetPort"`
}

// GetAllKubernetesServices loads the services in the target kubernetes cluster
func GetAllKubernetesServices() KubernetesServiceList {
	c := executil.Command{
		Name:       "Get Kubernetes Services",
		Executable: "kubectl",
		Arguments:  append(getKubernetesContextOptions(), "get", "--all-namespaces", "services", "-o", "yaml"),
	}
	c.Run()
	var kubernetesServiceList = KubernetesServiceList{}
	err := yaml.Unmarshal([]byte(c.GetStdout()), &kubernetesServiceList)
	if err != nil {
		panic(err)
	}
	return kubernetesServiceList
}
