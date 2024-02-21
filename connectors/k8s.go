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

type DeadlineAwareJob struct {
	deadline            time.Time
	id                  string
	NbCPU               int64
	Duration_in_seconds int
}

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
			pendingPod := events.NewPendingPod()
			pendingPod.PodId = pod.ObjectMeta.Name
			nbCpu, _ := pod.Spec.Containers[0].Resources.Requests.Cpu().AsInt64()
			if nbCpu > 0 {
				pendingPod.NbCores = int(nbCpu)
			}
			deadline, err := time.Parse(pod.Labels["deadline"], time.RFC3339)
			if err != nil {
				log.Warnf("Error %s while retrieving CPU request for Pod %v+\n", err, pod)
			}
			if deadline.After(time.Now().Add(time.Minute)) {
				pendingPod.Deadline = deadline
			}
			requestedTime, err := time.ParseDuration(pod.Labels["duration"])
			if err != nil {
				log.Warnf("Error %s while retrieving duration annotation for Pod %v+\n", err, pod)
			} else if requestedTime > time.Minute {
				pendingPod.RequestedTime = requestedTime
			}
			pendingPod.TimeCritical = (pod.Labels["timeCritical"] != "")
			channel <- pendingPod
		case watch.Modified:
			log.Infof("Pod %s/%s modified", pod.ObjectMeta.Namespace, pod.ObjectMeta.Name)
		case watch.Deleted:
			log.Infof("Pod %s/%s deleted", pod.ObjectMeta.Namespace, pod.ObjectMeta.Name)
			channel <- events.PodCompleted{PodId: pod.ObjectMeta.Name}
		}
	}
}

// DEPRECATED
func GetQueueSize() (int, int, []DeadlineAwareJob, error) {
	k8sConfig := K8sConfig{namespace: "default", labelSelector: "", kubeconfigPath: os.Getenv("KUBECONFIG")}
	namespace := k8sConfig.namespace
	selector := k8sConfig.labelSelector
	deadlineAwareJobs := []DeadlineAwareJob{}

	ctx := context.Background()
	config, err := clientcmd.BuildConfigFromFlags("", k8sConfig.kubeconfigPath)
	if err != nil {
		log.Errorf("Error while getting Kubernetes configuration %s", err)
		return -1, -1, nil, err
	}

	clientSet := kubernetes.NewForConfigOrDie(config)

	normalPods, err := GetPendingPods(clientSet, ctx, namespace, selector)
	if err != nil {
		log.Errorf("Error while getting pod state %s", err)
		return -1, -1, nil, err
	} else {
		for _, item := range normalPods {
			log.Debugf("Normal pods: %+v\n", item)
		}
	}
	timeCriticalPods, err := GetPendingPods(clientSet, ctx, namespace, "timeCritical=1")
	if err != nil {
		log.Errorf("Error while getting pod state %s", err)
		return -1, -1, nil, err
	} else {
		for _, item := range timeCriticalPods {
			log.Debugf("Time critical pods: %+v\n", item)
		}
	}
	deadlineAwarePods, err := GetPendingPods(clientSet, ctx, namespace, "deadline")
	if err != nil {
		log.Errorf("Error while getting pod state %s", err)
		return -1, -1, nil, err
	} else {
		for _, pod := range deadlineAwarePods {
			log.Debugf("Deadline aware pods: %+v\n", pod)
			deadline, err := time.Parse(pod.Labels["deadline"], time.RFC3339)
			if err != nil {
				log.Errorf("Error while parsing deadline: %s", err)
			}
			nbCpu, _ := pod.Spec.Containers[0].Resources.Requests.Cpu().AsInt64()
			job := DeadlineAwareJob{
				NbCPU: nbCpu, deadline: deadline, id: pod.Name,
			}
			deadlineAwareJobs = append(deadlineAwareJobs, job)
		}
	}
	return len(normalPods), len(timeCriticalPods), deadlineAwareJobs, nil
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
