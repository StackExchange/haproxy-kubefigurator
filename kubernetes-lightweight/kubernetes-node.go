package kube

import (
	"encoding/json"
	"io/ioutil"
	"net"
	"net/http"
)

// KubernetesNodeList returns a list of Kubernetes nodes in the target cluster
type KubernetesNodeList struct {
	Items []kubernetesNodeItem `json:"items"`
}

type kubernetesNodeItem struct {
	Metadata kubernetesNodeMetadata `json:"metadata"`
}

type kubernetesNodeMetadata struct {
	Name string `json:"name"`
}

// GetAllNodes loads the nodes in the target kubernetes cluster
func (k *Kube) GetAllNodes() KubernetesNodeList {
	var resp, err = http.Get(k.config.master + "/api/v1/nodes")
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	var body, _ = ioutil.ReadAll(resp.Body)
	var kubernetesNodeList = KubernetesNodeList{}
	err = json.Unmarshal(body, &kubernetesNodeList)
	if err != nil {
		panic(err)
	}
	return kubernetesNodeList
}

// GetIPs gets a map of node names to their IPs
func (n KubernetesNodeList) GetIPs() map[string]string {
	var nodes = make(map[string]string)
	for _, nodeItem := range n.Items {
		ip, err := net.LookupHost(nodeItem.Metadata.Name)
		if err != nil {
			panic(err)
		}
		nodes[nodeItem.Metadata.Name] = ip[0]
	}
	return nodes
}
