package k8s

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"text/template"
	"time"

	"github.com/google/uuid"
	corev1 "k8s.io/api/core/v1"
	v1c "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
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
const timeout = 30 * time.Second

type K8sHandler struct {
	client       dynamic.Interface
	clientset    *kubernetes.Clientset
	namespace    string
	executorName string
	pvcName      string
}

type ContainerSpec struct {
	Args           []string
	Name           string
	ContainerImage string
}

func CreateK8sHandler(executorName string, namespace string, pvcName string) (*K8sHandler, error) {
	handler := &K8sHandler{}
	handler.executorName = executorName
	handler.namespace = namespace
	handler.pvcName = pvcName

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

func (handler K8sHandler) SetupPVC(storageClass string, diskSize string) error {
	pvcExists, err := handler.DoesPVCExist(handler.pvcName)
	if err != nil {
		return err
	}

	if !pvcExists {
		spec := &PVCSpec{
			PVCName:      handler.pvcName,
			StorageClass: storageClass,
			DiskSize:     diskSize,
		}

		yaml, err := handler.ComposePVCYAML(spec)
		if err != nil {
			return err
		}

		err = handler.CreatePVC(yaml)
		if err != nil {
			return err
		}
	}

	return nil
}

func (handler K8sHandler) GetNamespaces() ([]string, error) {
	nsInterface := handler.clientset.CoreV1().Namespaces()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	ns, err := nsInterface.List(ctx, v1.ListOptions{})
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
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	err := handler.clientset.CoreV1().Namespaces().Delete(ctx, handler.namespace, metav1.DeleteOptions{})
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

func (handler K8sHandler) ComposePVCYAML(spec *PVCSpec) (string, error) {
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

	t, err := template.New("spec").Funcs(fmap).Parse(pvcTemplate)
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

func CreateUniqueJobName(baseName string) string {
	uniqueID := uuid.New()
	return fmt.Sprintf("%s-%s", baseName, uniqueID.String())
}

func (handler *K8sHandler) ComposeJobYAML(spec *JobSpec) (string, error) {
	spec.PVCName = handler.pvcName

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

	spec.JobContainerName = JobName + "-" + "container"
	spec.Namespace = handler.namespace

	t, err := template.New("spec").Funcs(fmap).Parse(jobTemplate)
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

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	_, err = handler.client.Resource(resource).Namespace(handler.namespace).Create(ctx, deployment, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	return nil
}

func (handler *K8sHandler) WaitForPods(jobName string, numberOfPods int) ([]string, error) {
	labelSelector := metav1.ListOptions{
		LabelSelector: labels.Set{"job-name": jobName}.AsSelector().String(),
	}

	if jobName == "" {
		return nil, errors.New("jobName must be provided")
	}

	maxRetries := 600
	retries := 0
	var podNames []string
	for {
		if retries >= maxRetries {
			return nil, errors.New("Failed to create job pods")
		}

		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		pods, err := handler.clientset.CoreV1().Pods(handler.namespace).List(ctx, labelSelector)
		if err != nil {
			return nil, err
		}
		if len(pods.Items) == numberOfPods {
			for _, pod := range pods.Items {
				podNames = append(podNames, pod.GetName())
			}
			break
		} else {
			time.Sleep(1 * time.Second)
			retries++
		}
	}

	retries = 0
	for {
		podCounter := 0
		if retries == maxRetries {
			return nil, errors.New("Pods failed to start")
		}
		for _, podName := range podNames {
			hasStarted, err := handler.HasPodStarted(podName)
			if err != nil {
				return nil, err
			}
			if hasStarted {
				podCounter++
			}
		}
		if podCounter == numberOfPods {
			break
		} else {
			time.Sleep(1 * time.Second)
		}
	}

	return podNames, nil
}

func (handler K8sHandler) CreatePVC(pvcYAML string) error {
	decUnstructured := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
	obj := &unstructured.Unstructured{}
	_, _, err := decUnstructured.Decode([]byte(pvcYAML), nil, obj)
	if err != nil {
		return err
	}

	pvcClient := handler.clientset.CoreV1().PersistentVolumeClaims(handler.namespace)

	var pvc = new(corev1.PersistentVolumeClaim)
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(obj.Object, pvc)
	if err != nil {
		return err
	}

	_, err = pvcClient.Create(context.TODO(), pvc, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	return nil
}

func (handler *K8sHandler) DoesPVCExist(pvcName string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	_, err := handler.clientset.CoreV1().PersistentVolumeClaims(handler.namespace).Get(ctx, pvcName, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return false, nil // PVC does not exist
		}
		return false, err
	}

	return true, nil
}

func (handler K8sHandler) CreateJob(jobYAML string, jobSpec *JobSpec) ([]string, error) {
	job := &unstructured.Unstructured{}
	dec := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
	_, _, err := dec.Decode([]byte(jobYAML), nil, job)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	resource := schema.GroupVersionResource{Group: "batch", Version: "v1", Resource: "jobs"}
	_, err = handler.client.Resource(resource).Namespace(handler.namespace).Create(ctx, job, metav1.CreateOptions{})
	if err != nil {
		return nil, err
	}

	return handler.WaitForPods(jobSpec.JobName, jobSpec.Parallelism)
}

func (handler *K8sHandler) DeleteDeployment(deploymentName string) error {
	client := handler.clientset.AppsV1().Deployments(handler.namespace)
	if client == nil {
		return errors.New("failed to delete deployment")
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	err := client.Delete(ctx, deploymentName, metav1.DeleteOptions{})
	if err != nil {
		return err
	}

	return nil
}

func (handler K8sHandler) GetDeploymentNames() ([]string, error) {
	var names []string
	listOptions := metav1.ListOptions{}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	resource := schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"}
	deployments, err := handler.client.Resource(resource).Namespace(handler.namespace).List(ctx, listOptions)
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

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	deployments, err := handler.client.Resource(resource).Namespace(handler.namespace).List(ctx, listOptions)
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
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	podInterface := handler.clientset.CoreV1().Pods(handler.namespace)
	pods, err := podInterface.List(ctx, v1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var podNames []string
	for _, pod := range pods.Items {
		podNames = append(podNames, pod.ObjectMeta.Name)
	}

	return podNames, nil
}

func (handler *K8sHandler) DeleteJob(jobName string) error {
	labelSelector := metav1.ListOptions{
		LabelSelector: labels.Set{"job-name": jobName}.AsSelector().String(), // this assumes your Pods are labeled with the name of the Job
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if err := handler.clientset.CoreV1().Pods(handler.namespace).DeleteCollection(ctx, metav1.DeleteOptions{}, labelSelector); err != nil {
		return err
	}

	propagationPolicy := metav1.DeletePropagationForeground
	deleteOptions := metav1.DeleteOptions{
		PropagationPolicy: &propagationPolicy,
	}

	ctx2, cancel2 := context.WithTimeout(context.Background(), timeout)
	defer cancel2()

	if err := handler.clientset.BatchV1().Jobs(handler.namespace).Delete(ctx2, jobName, deleteOptions); err != nil {
		return err
	}

	return nil
}

func (handler *K8sHandler) GetContainerNames(podName string) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	pod, err := handler.clientset.CoreV1().Pods(handler.namespace).Get(ctx, podName, metav1.GetOptions{})
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
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	err := handler.clientset.CoreV1().Pods(handler.namespace).Delete(ctx, podName, metav1.DeleteOptions{})
	if err != nil {
		return err
	}

	return nil
}

func (handler *K8sHandler) GetScale(deploymentName string) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	s, err := handler.clientset.AppsV1().
		Deployments(handler.namespace).
		GetScale(ctx, deploymentName, metav1.GetOptions{})
	if err != nil {
		return -1, err
	}

	return int(s.Spec.Replicas), nil
}

func (handler *K8sHandler) SetScale(replicas int, deploymentName string) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	s, err := handler.clientset.AppsV1().
		Deployments(handler.namespace).
		GetScale(ctx, deploymentName, metav1.GetOptions{})
	if err != nil {
		return err
	}

	sc := *s
	sc.Spec.Replicas = int32(replicas)

	ctx2, cancel2 := context.WithTimeout(context.Background(), timeout)
	defer cancel2()

	_, err = handler.clientset.AppsV1().
		Deployments(handler.namespace).
		UpdateScale(ctx2,
			deploymentName, &sc, metav1.UpdateOptions{})
	if err != nil {
		return err
	}

	return nil
}

func (handler *K8sHandler) WaitForPod(podName string, containerName string) error {
	for {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		pod, err := handler.clientset.CoreV1().Pods(handler.namespace).Get(ctx, podName, metav1.GetOptions{})
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
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	pod, err := handler.clientset.CoreV1().Pods(handler.namespace).Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		return false, err
	}

	if pod.Status.Phase == corev1.PodSucceeded || pod.Status.Phase == corev1.PodFailed {
		return true, nil
	}

	return false, nil
}

func (handler *K8sHandler) HasPodStarted(podName string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	pod, err := handler.clientset.CoreV1().Pods(handler.namespace).Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		return false, err
	}

	if pod.Status.Phase == corev1.PodSucceeded || pod.Status.Phase == corev1.PodFailed || pod.Status.Phase == corev1.PodRunning {
		return true, nil
	}

	finished, err := handler.HasPodFinished(podName)
	if err != nil {
		return false, err
	}

	if finished {
		return true, nil
	}

	return false, nil
}

func (handler *K8sHandler) HasContainerFinished(podName string, containerName string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	pod, err := handler.clientset.CoreV1().Pods(handler.namespace).Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		return false, err
	}

	for _, containerStatus := range pod.Status.ContainerStatuses {
		if containerStatus.Name == containerName {
			if containerStatus.State.Waiting != nil {
				return false, nil
			} else if containerStatus.State.Terminated != nil {
				return true, nil
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

func (handler *K8sHandler) HandleContainerLog(podName string, containerName string, aggregatedLogsChan chan string, eofChan chan bool, errChan chan error) {
	go func() {
		log, err := handler.GetLog(podName, containerName, true)
		if err != nil {
			errChan <- err
			return
		}

		for {
			select {
			case msg := <-log.MsgChan:
				aggregatedLogsChan <- msg
			case err := <-log.ErrChan:
				errChan <- err
				return
			case <-log.EofChan:
				if len(log.MsgChan) > 0 {
					msg := <-log.MsgChan
					aggregatedLogsChan <- msg
				}
				eofChan <- true
				return
			}
		}
	}()
}

func (handler *K8sHandler) HandlePodLog(podName string, aggregatedLogsChan chan string, eofChan chan bool, errChan chan error) {
	func() {
		containernames, err := handler.GetContainerNames(podName)
		if err != nil {
			errChan <- err
			return
		}

		for _, containername := range containernames {
			handler.HandleContainerLog(podName, containername, aggregatedLogsChan, eofChan, errChan)
		}
	}()
}

func (handler *K8sHandler) HandleJobLog(podNames []string, aggregatedLogsChan chan string, eofChan chan bool, errChan chan error) {
	func() {
		for _, podName := range podNames {
			handler.HandlePodLog(podName, aggregatedLogsChan, eofChan, errChan)
		}
	}()
}

func (handler *K8sHandler) PrintJobLogs(podNames []string, containers int) error {
	aggregatedLogsChan := make(chan string)
	eofChan := make(chan bool)
	errChan := make(chan error)
	handler.HandleJobLog(podNames, aggregatedLogsChan, eofChan, errChan)

	eofCounter := 0
	for {
		select {
		case msg := <-aggregatedLogsChan:
			fmt.Println(msg)
		case err := <-errChan:
			return err
		case <-eofChan:
			eofCounter++
			if eofCounter == len(podNames)*containers {
				return nil
			}
		}
	}
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

	log := &Log{MsgChan: make(chan string, 100), EofChan: make(chan bool, 100), ErrChan: make(chan error, 100)}
	count := int64(100)
	podLogOptions := v1c.PodLogOptions{
		Container: containerName,
		Follow:    follow,
		TailLines: &count,
	}

	podLogRequest := handler.clientset.CoreV1().Pods(handler.namespace).GetLogs(podName, &podLogOptions)

	go func() {
		var stream io.ReadCloser
		var err error
		retries := 0
		maxRetries := 600 // Wait max 600s for a Pod to start
		for {
			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()

			stream, err = podLogRequest.Stream(ctx)
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
			if numBytes > 0 {
				message := string(buf[:numBytes])
				log.MsgChan <- message
			} else {
				time.Sleep(500 * time.Millisecond)
			}
			if err != nil {
				containerFinished, err := handler.HasContainerFinished(podName, containerName)
				if err != nil {
					log.ErrChan <- err
					return
				}
				if containerFinished {
					log.EofChan <- true
					return
				}
			}
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

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	stream, err := podLogRequest.Stream(ctx)
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
