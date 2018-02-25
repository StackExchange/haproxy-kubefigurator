package haproxyconfigurator

import (
	"strings"

	"github.com/ghodss/yaml"
	"go.mikenewswanger.com/utilities/executil"
)

// getAllKubernetesNodes loads the nodes in the target kubernetes cluster
func getAllKubernetesNodes() map[string]string {
	nodes := map[string]string{}
	c := executil.Command{
		Name:       "Get Kubernetes Nodes",
		Executable: "kubectl",
		Arguments:  append(getKubernetesContextOptions(), "get", "nodes", "-o", `jsonpath={range .items[*]}{@.metadata.name} {@.status.addresses[?(@.type=="InternalIP")].address}{"\n"}{end}`),
	}
	c.Run()
	kubectlOutput := strings.Split(c.GetStdout(), "\n")
	if len(kubectlOutput) == 0 {
		logger.Fatal("No kubernetes nodes were found")
	}
	for _, nodeRaw := range kubectlOutput {
		if nodeRaw == "" {
			continue
		}
		split := strings.Fields(nodeRaw)
		nodes[split[0]] = split[1]
	}
	return nodes
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
		Arguments: append(
			getKubernetesContextOptions(),
			"get",
			"--all-namespaces",
			"-o",
			"yaml",
			"services",
			"--selector",
			"service-router.enabled=yes",
		),
	}
	c.Run()
	var kubernetesServiceList = KubernetesServiceList{}
	err := yaml.Unmarshal([]byte(c.GetStdout()), &kubernetesServiceList)
	if err != nil {
		panic(err)
	}
	return kubernetesServiceList
}
