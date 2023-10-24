package k8s

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
	"time"

	corev1 "k8s.io/api/core/v1"
	v1c "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

const Namespace = "colonies"
const DeploymentName = "executor-deployment"
const JobName = "executorjob"

type K8sHandler struct {
	client       dynamic.Interface
	clientset    *kubernetes.Clientset
	namespace    string
	executorName string
	jobCounter   int
}

type ContainerSpec struct {
	Args           []string
	Name           string
	ContainerImage string
}

func CreateK8sHandler(executorName string, namespace string) (*K8sHandler, error) {
	handler := &K8sHandler{}
	handler.namespace = namespace
	handler.executorName = executorName
	handler.jobCounter = 0

	var err error
	handler.client, handler.clientset, err = handler.setupK8sClient()
	if err != nil {
		return nil, err
	}

	return handler, nil
}

func (handler *K8sHandler) setupK8sClient() (dynamic.Interface, *kubernetes.Clientset, error) {
	home := homedir.HomeDir()
	kubeconfig := filepath.Join(home, ".kube", "config")

	var config *rest.Config
	var err error

	config, err = rest.InClusterConfig()
	if err != nil {
		if os.Getenv("KUBECONFIG") != "" {
			config, err = clientcmd.BuildConfigFromFlags("", os.Getenv("KUBECONFIG"))
			if err != nil {
				return nil, nil, err
			}
		} else {
			config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
			if err != nil {
				return nil, nil, err
			}
		}
	}

	client, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, nil, err
	}

	clientset, err := kubernetes.NewForConfig(config)

	return client, clientset, nil
}

func (handler K8sHandler) CreateNamespace() error {
	nsSpec := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: handler.namespace,
		},
	}

	_, err := handler.clientset.CoreV1().Namespaces().Create(context.Background(), nsSpec, metav1.CreateOptions{})
	return err
}

func (handler K8sHandler) GetNamespaces() ([]string, error) {
	nsInterface := handler.clientset.CoreV1().Namespaces()

	ns, err := nsInterface.List(context.TODO(), v1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var namespaceNames []string
	for _, d := range ns.Items {
		namespaceNames = append(namespaceNames, d.ObjectMeta.Name)
	}

	return namespaceNames, nil
}

func (handler K8sHandler) DeleteNamespace() error {
	err := handler.clientset.CoreV1().Namespaces().Delete(context.TODO(), handler.namespace, metav1.DeleteOptions{})
	if err != nil {
		return err
	}

	return nil
}

func (handler K8sHandler) ComposeDeploymentYAML(spec DeploymentSpec, deploymentName string) (string, error) {
	fmap := template.FuncMap{
		"Iterate": func(count int) []uint {
			var i uint
			var Items []uint
			for i = 0; i < (uint(count)); i++ {
				Items = append(Items, i)
			}
			return Items
		},
	}

	spec.DeploymentName = deploymentName
	spec.Namespace = handler.namespace

	t, err := template.New("spec").Funcs(fmap).Parse(deploymentTemplate)
	if err != nil {
		return "", err
	}

	var buffer bytes.Buffer
	err = t.Execute(&buffer, spec)
	if err != nil {
		return "", err
	}

	return buffer.String(), nil
}

func (handler K8sHandler) ComposeJobYAML(spec JobSpec) (string, string, error) {
	fmap := template.FuncMap{
		"Iterate": func(count int) []uint {
			var i uint
			var Items []uint
			for i = 0; i < (uint(count)); i++ {
				Items = append(Items, i)
			}
			return Items
		},
	}

	spec.JobName = handler.executorName + "-" + JobName + "-" + strconv.Itoa(handler.jobCounter)
	handler.jobCounter++
	spec.Namespace = handler.namespace

	t, err := template.New("spec").Funcs(fmap).Parse(jobTemplate)
	if err != nil {
		return "", spec.JobName, err
	}

	var buffer bytes.Buffer
	err = t.Execute(&buffer, spec)
	if err != nil {
		return "", spec.JobName, err
	}

	return buffer.String(), spec.JobName, nil
}

func (handler K8sHandler) CreateDockerRegistrySecret(dockerSecret *DockerRegistrySecret) error {
	dockerSecretSerialized, err := dockerSecret.Serialize()
	if err != nil {
		return err
	}

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "prvdockerreg",
			Namespace: handler.namespace,
		},
		Type:       "kubernetes.io/dockerconfigjson",
		StringData: map[string]string{".dockerconfigjson": dockerSecretSerialized},
	}

	_, err = handler.clientset.CoreV1().Secrets(handler.namespace).Create(context.Background(), secret, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	return nil
}

func (handler K8sHandler) CreateDeployment(deploymentYAML string) error {
	deployment := &unstructured.Unstructured{}
	dec := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
	_, _, err := dec.Decode([]byte(deploymentYAML), nil, deployment)
	if err != nil {
		return err
	}

	resource := schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"}
	_, err = handler.client.Resource(resource).Namespace(handler.namespace).Create(context.TODO(), deployment, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	return nil
}

func (handler K8sHandler) CreateJob(jobYAML string, jobName string, jobSpec *JobSpec) ([]string, error) {
	job := &unstructured.Unstructured{}
	dec := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
	_, _, err := dec.Decode([]byte(jobYAML), nil, job)
	if err != nil {
		return nil, err
	}

	resource := schema.GroupVersionResource{Group: "batch", Version: "v1", Resource: "jobs"}
	_, err = handler.client.Resource(resource).Namespace(handler.namespace).Create(context.TODO(), job, metav1.CreateOptions{})
	if err != nil {
		return nil, err
	}

	maxRetries := 600
	retries := 0
	var podNames []string
	for {
		if retries == maxRetries {
			return nil, errors.New("Pods failed to start")
		}
		podNames, err = handler.GetPodNames()
		if err != nil {
			return nil, err
		}
		if len(podNames) != jobSpec.Parallelism {
			time.Sleep(1 * time.Second)
			retries++
			continue
		} else {
			break
		}
	}

	var jobPodNames []string
	for _, podName := range podNames {
		if strings.HasPrefix(podName, jobName) {
			jobPodNames = append(jobPodNames, podName)
		}
	}

	return jobPodNames, nil
}

func (handler *K8sHandler) DeleteDeployment(deploymentName string) error {
	client := handler.clientset.AppsV1().Deployments(handler.namespace)
	if client == nil {
		return errors.New("failed to delete deployment")
	}

	err := client.Delete(context.TODO(), deploymentName, metav1.DeleteOptions{})
	if err != nil {
		return err
	}

	return nil
}

func (handler K8sHandler) GetDeploymentNames() ([]string, error) {
	var names []string
	listOptions := metav1.ListOptions{}

	resource := schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"}
	deployments, err := handler.client.Resource(resource).Namespace(handler.namespace).List(context.TODO(), listOptions)
	if err != nil {
		return names, err
	}

	for _, d := range deployments.Items {
		metadata := d.Object["metadata"].(map[string]interface{})
		name := metadata["name"].(string)
		names = append(names, name)
	}

	return names, err
}

func (handler K8sHandler) GetJobNames() ([]string, error) {
	var names []string
	listOptions := metav1.ListOptions{}

	resource := schema.GroupVersionResource{Group: "batch", Version: "v1", Resource: "jobs"}
	deployments, err := handler.client.Resource(resource).Namespace(handler.namespace).List(context.TODO(), listOptions)
	if err != nil {
		return names, err
	}

	for _, d := range deployments.Items {
		metadata := d.Object["metadata"].(map[string]interface{})
		name := metadata["name"].(string)
		names = append(names, name)
	}

	return names, err
}

func (handler *K8sHandler) GetPodNames() ([]string, error) {
	podInterface := handler.clientset.CoreV1().Pods(handler.namespace)
	pods, err := podInterface.List(context.TODO(), v1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var podNames []string
	for _, pod := range pods.Items {
		//fmt.Println(string(pod.Status.Phase))
		podNames = append(podNames, pod.ObjectMeta.Name)
	}

	return podNames, nil
}

func (handler *K8sHandler) GetContainerNames(podName string) ([]string, error) {
	pod, err := handler.clientset.CoreV1().Pods(handler.namespace).Get(context.TODO(), podName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	var containerNames []string
	for _, container := range pod.Spec.Containers {
		containerNames = append(containerNames, container.Name)
	}

	return containerNames, nil
}

func (handler *K8sHandler) RestartPod(podName string) error {
	err := handler.clientset.CoreV1().Pods(handler.namespace).Delete(context.TODO(), podName, metav1.DeleteOptions{})
	if err != nil {
		return err
	}

	return nil
}

func (handler *K8sHandler) GetScale(deploymentName string) (int, error) {
	s, err := handler.clientset.AppsV1().
		Deployments(handler.namespace).
		GetScale(context.TODO(), deploymentName, metav1.GetOptions{})
	if err != nil {
		return -1, err
	}

	return int(s.Spec.Replicas), nil
}

func (handler *K8sHandler) SetScale(replicas int, deploymentName string) error {
	s, err := handler.clientset.AppsV1().
		Deployments(handler.namespace).
		GetScale(context.TODO(), deploymentName, metav1.GetOptions{})
	if err != nil {
		return err
	}

	sc := *s
	sc.Spec.Replicas = int32(replicas)

	_, err = handler.clientset.AppsV1().
		Deployments(handler.namespace).
		UpdateScale(context.TODO(),
			deploymentName, &sc, metav1.UpdateOptions{})
	if err != nil {
		return err
	}

	return nil
}

func (handler *K8sHandler) WaitForPod(podName string, containerName string) error {
	for {
		pod, err := handler.clientset.CoreV1().Pods(handler.namespace).Get(context.TODO(), podName, metav1.GetOptions{})
		if err != nil {
			return err
		}

		if pod.Status.Phase == "Running" || pod.Status.Phase == "Succeeded" {
			return nil
		}

		time.Sleep(1000 * time.Millisecond)
	}
}

func (handler *K8sHandler) HasPodFinished(podName string) (bool, error) {
	pod, err := handler.clientset.CoreV1().Pods(handler.namespace).Get(context.TODO(), podName, metav1.GetOptions{})
	if err != nil {
		return false, err
	}

	if pod.Status.Phase == corev1.PodSucceeded || pod.Status.Phase == corev1.PodFailed {
		return true, nil
	}

	return false, nil
}

func (handler *K8sHandler) HasContainerFinished(podName string, containerName string) (bool, error) {
	pod, err := handler.clientset.CoreV1().Pods(handler.namespace).Get(context.TODO(), podName, metav1.GetOptions{})
	if err != nil {
		return false, err
	}

	for _, containerStatus := range pod.Status.ContainerStatuses {
		if containerStatus.Name == containerName {
			if containerStatus.State.Terminated != nil {
				if containerStatus.State.Terminated.ExitCode == 0 {
					return true, nil
				} else {
					return true, fmt.Errorf("container %s terminated with error: %s", containerName, containerStatus.State.Terminated.Message)
				}
			}

			return false, nil
		}
	}

	return false, fmt.Errorf("container %s not found in pod %s", containerName, podName)
}

type Log struct {
	MsgChan chan string
	EofChan chan bool
	ErrChan chan error
}

func (handler *K8sHandler) PrintAllLogs(podName string, follow bool) error {
	containerNames, err := handler.GetContainerNames(podName)
	if err != nil {
		return err
	}

outerloop:
	for _, containerName := range containerNames {
		log, err := handler.GetLog(podName, containerName, follow)
		if err != nil {
			return err
		}

	innerloop:
		for {
			select {
			case msg := <-log.MsgChan:
				fmt.Println(msg)
			case err := <-log.ErrChan:
				fmt.Println(err)
				break outerloop
			case <-log.EofChan:
				if len(log.MsgChan) > 0 {
					msg := <-log.MsgChan
					fmt.Println(msg)
				}
				break innerloop
			}
		}
	}

	return nil
}

func (handler *K8sHandler) PrintLogs(podName string, containerName string, follow bool) error {
	log, err := handler.GetLog(podName, containerName, follow)
	if err != nil {
		return err
	}

	for {
		select {
		case msg := <-log.MsgChan:
			fmt.Println(msg)
		case err := <-log.ErrChan:
			fmt.Println(err)
			return err
		case <-log.EofChan:
			if len(log.MsgChan) > 0 {
				msg := <-log.MsgChan
				fmt.Println(msg)
			}
			return nil
		}
	}
}

func (handler *K8sHandler) GetLog(podName string, containerName string, follow bool) (*Log, error) {
	log := &Log{MsgChan: make(chan string, 100), EofChan: make(chan bool, 100), ErrChan: make(chan error, 100)}
	count := int64(100)
	podLogOptions := v1c.PodLogOptions{
		Container: containerName,
		Follow:    follow,
		TailLines: &count,
	}

	podNames, err := handler.GetPodNames()
	if err != nil {
		return nil, err
	}
	found := false
	for _, p := range podNames {
		if podName == p {
			found = true
		}
	}
	if !found {
		return nil, errors.New("Pod with name " + podName + " does not exists")
	}

	containerNames, err := handler.GetContainerNames(podName)
	if err != nil {
		return nil, err
	}
	found = false
	for _, c := range containerNames {
		if containerName == c {
			found = true
		}
	}
	if !found {
		return nil, errors.New("Container with name " + podName + " does not exists in pod " + podName)
	}

	podLogRequest := handler.clientset.CoreV1().Pods(handler.namespace).GetLogs(podName, &podLogOptions)

	func() {
		var stream io.ReadCloser
		var err error
		retries := 0
		maxRetries := 600 // Wait max 600s for a Pod to start
		for {
			stream, err = podLogRequest.Stream(context.TODO())
			if err != nil {
				if retries == maxRetries {
					log.ErrChan <- errors.New("Exceeded maxRetries: " + err.Error())
					break
				}
				time.Sleep(1 * time.Second)
				retries++
				continue
			} else {
				break
			}
		}

		defer stream.Close()

		for {
			buf := make([]byte, 2000)
			numBytes, err := stream.Read(buf)
			if err == io.EOF {
				containerFinished, err := handler.HasContainerFinished(podName, containerName)
				if err != nil {
					log.ErrChan <- err
					break
				}
				if containerFinished {
					log.EofChan <- true
					break
				}
				continue
			}
			if err != nil {
				log.ErrChan <- errors.New("Failed to read log buffer: " + err.Error())
				break
			}
			if numBytes == 0 {
				continue
			}

			message := string(buf[:numBytes])
			log.MsgChan <- message
		}
	}()

	return log, nil
}

func (handler *K8sHandler) GetStdOut(podName string, containerName string) (string, error) {
	count := int64(100)
	podLogOptions := v1c.PodLogOptions{
		Container: containerName,
		Follow:    false,
		TailLines: &count,
	}

	podLogRequest := handler.clientset.CoreV1().Pods(handler.namespace).GetLogs(podName, &podLogOptions)

	stream, err := podLogRequest.Stream(context.TODO())
	if err != nil {
		return "", err
	}

	defer stream.Close()

	allLines := ""
	for {
		buf := make([]byte, 2000)
		numBytes, err := stream.Read(buf)
		if numBytes == 0 {
			continue
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}

		message := string(buf[:numBytes])
		allLines += message

	}

	return allLines, nil
}
