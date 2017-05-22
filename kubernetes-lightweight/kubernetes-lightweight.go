package kube

type Kube struct {
	config kubeConfig
}

type kubeConfig struct {
	master string
}

// Initialize a new kubernetes connector
func (k *Kube) Initialize(master string) {
	k.config.master = master
}
