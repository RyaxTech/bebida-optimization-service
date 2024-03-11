package connectors

import (
	"context"
	"os"
	"time"

	"github.com/RyaxTech/bebida-shaker/events"
	"github.com/apex/log"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type K8sConfig struct {
	namespace      string
	labelSelector  string
	kubeconfigPath string
}

var bebida_prefix = "ryax.tech/"

func WatchQueues(channel chan interface{}) {
	k8sConfig := K8sConfig{namespace: "default", labelSelector: "", kubeconfigPath: os.Getenv("KUBECONFIG")}

	config, err := clientcmd.BuildConfigFromFlags("", k8sConfig.kubeconfigPath)
	if err != nil {
		log.Errorf("Error while getting Kubernetes configuration %s", err)
	}
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Errorf("Error while creating Kubernetes client %s", err)
	}

	ctx := context.Background()
	watcher, err := client.CoreV1().Pods(v1.NamespaceDefault).Watch(ctx, metav1.ListOptions{})
	if err != nil {
		panic(err)
	}

	for event := range watcher.ResultChan() {
		pod := event.Object.(*v1.Pod)

		switch event.Type {
		case watch.Added:
			log.Infof("Pod %s/%s added", pod.ObjectMeta.Namespace, pod.ObjectMeta.Name)
			// Exclude pod if explicitly requested
			if pod.Annotations[bebida_prefix + "bebida"] == "exclude" {
				continue
			}
			log.Infof("Pod %s/%s is a Bebida operator!", pod.ObjectMeta.Namespace, pod.ObjectMeta.Name)
			pendingPod := events.NewPendingPod()
			pendingPod.PodId = pod.ObjectMeta.Name
			nbCpu, _ := pod.Spec.Containers[0].Resources.Requests.Cpu().AsInt64()
			if nbCpu > 0 {
				pendingPod.NbCores = int(nbCpu)
			}
			deadline, err := time.Parse(pod.Annotations[bebida_prefix + "deadline"], time.RFC3339)
			if err != nil {
				log.Warnf("Error %s while retrieving CPU request for Pod %v+\n", err, pod)
			}
			if deadline.After(time.Now().Add(time.Minute)) {
				pendingPod.Deadline = deadline
			}
			requestedTime, err := time.ParseDuration(pod.Annotations[bebida_prefix + "duration"])
			if err != nil {
				log.Warnf("Error %s while retrieving duration annotation for Pod %v+\n", err, pod)
			} else if requestedTime > time.Minute {
				pendingPod.RequestedTime = requestedTime
			}
			pendingPod.TimeCritical = (pod.Annotations[bebida_prefix + "timeCritical"] != "")
			channel <- pendingPod
		case watch.Modified:
			log.Infof("Pod %s/%s modified with status: %s", pod.ObjectMeta.Namespace, pod.ObjectMeta.Name, pod.Status)
			if pod.DeletionTimestamp != nil {
				continue
			}
			switch pod.Status.Phase {
			// case v1.PodRunning:
			// 	channel <- events.PodStarted
			case v1.PodSucceeded:
				channel <- events.PodCompleted{PodId: pod.ObjectMeta.Name}
			case v1.PodFailed:
				channel <- events.PodCompleted{PodId: pod.ObjectMeta.Name}
			}

		case watch.Deleted, watch.Error:
			log.Infof("Pod %s/%s deleted", pod.ObjectMeta.Namespace, pod.ObjectMeta.Name)
			channel <- events.PodCompleted{PodId: pod.ObjectMeta.Name}
		}
	}
}
