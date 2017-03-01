package main

type kubernetesServiceList struct {
	Items []kubernetesServiceItem `json:"items"`
}

type kubernetesServiceItem struct {
	Metadata kubernetesServiceMetadata `json:"metadata"`
	Spec     kubernetesServiceSpec     `json:"spec"`
}

type kubernetesServiceMetadata struct {
	Name      string            `json:"name"`
	Namespace string            `json:"namespace"`
	Labels    map[string]string `json:"labels"`
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
