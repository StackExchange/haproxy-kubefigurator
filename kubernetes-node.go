package main

type kubernetesNodeList struct {
	Items []kubernetesNodeItem `json:"items"`
}

type kubernetesNodeItem struct {
	Metadata kubernetesNodeMetadata `json:"metadata"`
}

type kubernetesNodeMetadata struct {
	Name string `json:"name"`
}
