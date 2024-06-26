package connectors

import (
	"context"
	"os"
	"strconv"
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

func WatchQueues(channel chan interface{}, defaultPunchNbCore int, defaultPunchDuration int) {
	defer panic("The kubernetes watcher stopped for unknown reason. Quit!")

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
			if pod.Annotations[bebida_prefix+"bebida"] == "exclude" {
				log.Infof("Found exclusion annotation %s, do exclude this pod", bebida_prefix+"bebida: exclude")
				continue
			}
			pendingPod := events.NewPendingPod(defaultPunchNbCore, defaultPunchDuration)
			pendingPod.PodId = pod.ObjectMeta.Name

			nbCpu := 0
			if pod.Annotations[bebida_prefix+"resources.cores"] != "" {
				nbCpu, err = strconv.Atoi(pod.Annotations[bebida_prefix+"resources.cores"])
				if err != nil {
					log.Warnf("Error %s while parsing resources.cores annotation for Pod %v+\n", err, pod)
				}
			}
			if nbCpu > 0 {
				pendingPod.NbCores = int(nbCpu)
			}

			deadline, err := time.Parse(time.RFC3339, pod.Annotations[bebida_prefix+"deadline"])
			if err != nil {
				log.Warnf("Error %s while retrieving deadline for Pod %v+\n", err, pod)
			}
			if deadline.After(time.Now().Add(time.Minute)) {
				pendingPod.Deadline = deadline
			}
			requestedTime, err := time.ParseDuration(pod.Annotations[bebida_prefix+"duration"])
			if err != nil {
				log.Warnf("Error %s while retrieving duration annotation for Pod %v+\n", err, pod)
			} else if requestedTime > time.Minute {
				pendingPod.RequestedTime = requestedTime
			}
			pendingPod.TimeCritical = (pod.Annotations[bebida_prefix+"timeCritical"] != "")
			channel <- pendingPod
		case watch.Modified:
			log.Infof("Pod %s/%s modified with status: %s", pod.ObjectMeta.Namespace, pod.ObjectMeta.Name, pod.Status.Phase)
			if pod.DeletionTimestamp != nil {
				continue
			}
			switch pod.Status.Phase {
			// case v1.PodRunning:
			// 	channel <- events.PodStarted
			case v1.PodSucceeded, v1.PodFailed:
				// Exclude pod if explicitly requested
				if pod.Annotations[bebida_prefix+"bebida"] == "exclude" {
					log.Infof("Found exclusion annotation %s, do exclude this pod", bebida_prefix+"bebida: exclude")
					continue
				}
				nbCores, err := strconv.Atoi(pod.Annotations[bebida_prefix+"resources.cores"])
				if err != nil {
					log.Warnf("Error %s while parsing resources.cores annotation for Pod %v+\n", err, pod)
					channel <- events.PodCompleted{PodId: pod.ObjectMeta.Name}
				} else {
					// Only set time critical if the number of cores is set	
					timeCritical := (pod.Annotations[bebida_prefix+"timeCritical"] != "")
					channel <- events.PodCompleted{PodId: pod.ObjectMeta.Name, TimeCritical: timeCritical, NbCores: nbCores}
				}
			}

		case watch.Deleted, watch.Error:
			log.Infof("Pod %s/%s deleted", pod.ObjectMeta.Namespace, pod.ObjectMeta.Name)
			channel <- events.PodCompleted{PodId: pod.ObjectMeta.Name}
		}
	}
}
