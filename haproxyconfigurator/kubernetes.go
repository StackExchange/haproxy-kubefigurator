package haproxyconfigurator

import (
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type kubernetesNodeIPs map[string]string

// getAllKubernetesNodes loads the nodes in the target kubernetes cluster
func getAllKubernetesNodes() (kubernetesNodeIPs, error) {
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigFile)
	if err != nil {
		return nil, err
	}

	nodeIPs := kubernetesNodeIPs{}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	nodes, err := clientset.CoreV1().Nodes().List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	for _, node := range nodes.Items {
		for _, address := range node.Status.Addresses {
			if address.Type == "InternalIP" {
				nodeIPs[node.Name] = address.Address
			}
		}
	}
	return nodeIPs, nil
}

func getProxiedKubernetesServices() ([]v1.Service, error) {
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigFile)
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	proxiedServices := []v1.Service{}
	services, err := clientset.CoreV1().Services(v1.NamespaceAll).List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	for _, service := range services.Items {
		if service.Labels["haproxy-kubefigurator.enabled"] == "yes" {
			proxiedServices = append(proxiedServices, service)
		}
	}
	return proxiedServices, nil
}

func waitForChanges() error {
	// // use the current context in kubeconfig
	// config, err := clientcmd.BuildConfigFromFlags("", kubeconfigFile)
	// if err != nil {
	// 	return err
	// }

	// clientset, err := kubernetes.NewForConfig(config)
	// if err != nil {
	// 	return err
	// }

	// w, err := clientset.CoreV1().Services("").Watch(metav1.ListOptions{})
	// return err
	return nil
}
