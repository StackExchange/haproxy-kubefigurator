package kube

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

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

// GetAllServices loads the services in the target kubernetes cluster
func (k *Kube) GetAllServices() KubernetesServiceList {
	var resp, err = http.Get(k.config.master + "/api/v1/services")
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	var body, _ = ioutil.ReadAll(resp.Body)
	var kubernetesServiceList = KubernetesServiceList{}
	err = json.Unmarshal(body, &kubernetesServiceList)
	if err != nil {
		panic(err)
	}
	return kubernetesServiceList
}
