package connectors

import (
	"context"
	"os"

	"github.com/apex/log"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type K8sConfig struct {
	namespace      string
	labelSelector  string
	kubeconfigPath string
}

func GetQueueSize() (int, error) {
	k8sConfig := K8sConfig{namespace: "default", labelSelector: "", kubeconfigPath: os.Getenv("KUBECONFIG")}
	namespace := k8sConfig.namespace
	selector := k8sConfig.labelSelector

	ctx := context.Background()
	config, err := clientcmd.BuildConfigFromFlags("", k8sConfig.kubeconfigPath)
	if err != nil {
		log.Errorf("Error while getting Kubernetes configuration %s", err)
		return -1, err
	}

	clientSet := kubernetes.NewForConfigOrDie(config)

	items, err := GetPendingPods(clientSet, ctx, namespace, selector)
	if err != nil {
		log.Errorf("Error while getting pod state %s", err)
		return -1, err
	} else {
		for _, item := range items {
			log.Debugf("%+v\n", item)
		}
	}
	return len(items), nil
}

func GetPendingPods(clientSet *kubernetes.Clientset, ctx context.Context, namespace string, selector string) ([]v1.Pod, error) {

	list, err := clientSet.CoreV1().Pods(namespace).
		List(ctx, metav1.ListOptions{LabelSelector: selector, FieldSelector: "status.phase=Pending"})
	if err != nil {
		return nil, err
	}
	return list.Items, nil
}

func GetAllPods(clientSet *kubernetes.Clientset, ctx context.Context, namespace string, selector string) ([]v1.Pod, error) {

	list, err := clientSet.CoreV1().Pods(namespace).
		List(ctx, metav1.ListOptions{LabelSelector: selector})
	if err != nil {
		return nil, err
	}
	return list.Items, nil
}

func GetNbRunningApp() (int, error) {
	k8sConfig := K8sConfig{namespace: "default", labelSelector: "", kubeconfigPath: os.Getenv("KUBECONFIG")}
	namespace := k8sConfig.namespace
	selector := k8sConfig.labelSelector

	ctx := context.Background()
	config, err := clientcmd.BuildConfigFromFlags("", k8sConfig.kubeconfigPath)
	if err != nil {
		log.Errorf("Error while getting Kubernetes configuration %s", err)
		return -1, err
	}

	clientSet := kubernetes.NewForConfigOrDie(config)

	items, err := GetAllPods(clientSet, ctx, namespace, selector)
	if err != nil {
		log.Errorf("Error while getting pod state %s", err)
		return -1, err
	} else {
		for _, item := range items {
			log.Debugf("%+v\n", item)
		}
	}
	return len(items), nil
}
