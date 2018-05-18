package haproxyconfigurator

import (
	"time"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type kubernetesNodeIPs map[string]string

func kubeClient(kubeConfigPath string) (*kubernetes.Clientset, error) {
	config, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	if err != nil {
		return nil, err
	}
	return kubernetes.NewForConfig(config)
}

// getAllKubernetesNodes loads the nodes in the target kubernetes cluster
func getAllKubernetesNodes(client *kubernetes.Clientset) (kubernetesNodeIPs, error) {
	nodes, err := client.CoreV1().Nodes().List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	nodeIPs := kubernetesNodeIPs{}
	for _, node := range nodes.Items {
		for _, address := range node.Status.Addresses {
			if address.Type == "InternalIP" {
				nodeIPs[node.Name] = address.Address
			}
		}
	}
	return nodeIPs, nil
}

func getProxiedKubernetesServices(client *kubernetes.Clientset) ([]v1.Service, error) {
	proxiedServices := []v1.Service{}
	services, err := client.CoreV1().Services(v1.NamespaceAll).List(metav1.ListOptions{})
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

func watchForServiceChanges(client *kubernetes.Clientset, ch chan<- bool) {
	for {
		start := time.Now()
		logger.Debug("Watching for service changes")
		w, err := client.CoreV1().Services("").Watch(metav1.ListOptions{})
		if err != nil {
			logger.Error(err)
			time.Sleep(time.Second)
			continue
		}
		var timer *time.Timer
		const quietTime = time.Second * 2
		for ev := range w.ResultChan() {
			logger.Infof("Detected change to service %s (%s)", ev.Object.(*v1.Service).Name, ev.Type)
			if timer != nil {
				timer.Stop()
			}
			timer = time.AfterFunc(quietTime, func() {
				select {
				case ch <- true: // if can't send, there is already a pending update.
				default:
					logger.Infof("Queue full. Already a pending config update.")
				}
			})
		}
		logger.Infof("Watch closed after %s", time.Now().Sub(start))
	}
}
