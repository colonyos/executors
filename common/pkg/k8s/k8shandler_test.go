package k8s

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const deploymentWaittime = 10 // 5 seconds

func TestK8sHandlerComposeDeployment(t *testing.T) {
	namespace := "testnamespace0"
	handler, err := CreateK8sHandler("testexecutor", namespace, namespace+"kubeexecutor-pvc")
	assert.Nil(t, err)

	spec := createTestDeploymentSpec()
	yaml, err := handler.ComposeDeploymentYAML(spec, DeploymentName)
	assert.Nil(t, err)
	fmt.Println(yaml)
}

func TestK8sHandlerComposeJob(t *testing.T) {
	namespace := "testnamespace1"
	handler, err := CreateK8sHandler("testexecutor", namespace, namespace+"kubeexecutor-pvc")
	assert.Nil(t, err)

	spec := createTestJobSpec()
	fmt.Println(spec)
	yaml, err := handler.ComposeJobYAML(&spec)
	assert.Nil(t, err)
	fmt.Println(yaml)
}

func TestK8sHandlerComposePVC(t *testing.T) {
	namespace := "testnamespace2"
	handler, err := CreateK8sHandler("testexecutor", namespace, namespace+"kubeexecutor-pvc")
	assert.Nil(t, err)

	pvcSpec := createTestPVCSpec()
	yaml, err := handler.ComposePVCYAML(pvcSpec)
	assert.Nil(t, err)
	fmt.Println(yaml)
}

func TestK8sHandlerCreateJob(t *testing.T) {
	namespace := "testnamespace3"
	handler, err := CreateK8sHandler("testexecutor", namespace, namespace+"kubeexecutor-pvc")
	assert.Nil(t, err)

	err = handler.CreateNamespace()
	assert.Nil(t, err)

	err = handler.SetupPVC("longhorn", "1Gi")
	assert.Nil(t, err)

	spec := createTestJobSpec()
	fmt.Println(spec)
	yaml, err := handler.ComposeJobYAML(&spec)
	assert.Nil(t, err)
	fmt.Println(yaml)

	jobPodNames, err := handler.CreateJob(yaml, &spec)
	assert.Nil(t, err)
	fmt.Println(jobPodNames)

	jobNames, err := handler.GetJobNames()
	assert.Nil(t, err)
	assert.Len(t, jobNames, 1)
	assert.Equal(t, jobNames[0], spec.JobName)

	err = handler.PrintJobLogs(jobPodNames, spec.ContainersPerPod)
	assert.Nil(t, err)

	err = handler.DeleteNamespace()
	assert.Nil(t, err)
}

func TestK8sHandlerDeleteJob(t *testing.T) {
	namespace := "testnamespace4"
	handler, err := CreateK8sHandler("testexecutor", namespace, namespace+"kubeexecutor-pvc")
	assert.Nil(t, err)

	err = handler.CreateNamespace()
	assert.Nil(t, err)

	err = handler.SetupPVC("longhorn", "1Gi")
	assert.Nil(t, err)

	spec := createTestJobSpec()
	fmt.Println(spec)
	yaml, err := handler.ComposeJobYAML(&spec)
	assert.Nil(t, err)
	fmt.Println(yaml)

	jobPodNames, err := handler.CreateJob(yaml, &spec)
	assert.Nil(t, err)
	fmt.Println(jobPodNames)

	jobNames, err := handler.GetJobNames()
	assert.Nil(t, err)
	assert.Len(t, jobNames, 1)

	err = handler.DeleteJob(spec.JobName)
	assert.Nil(t, err)

	jobNames, err = handler.GetJobNames()
	assert.Nil(t, err)
	assert.Len(t, jobNames, 0)

	err = handler.DeleteNamespace()
	assert.Nil(t, err)
}

func TestK8sHandlerGetLogsInvalidContainer(t *testing.T) {
	namespace := "testnamespace5"
	handler, err := CreateK8sHandler("testexecutor", namespace, namespace+"kubeexecutor-pvc")
	assert.Nil(t, err)

	err = handler.CreateNamespace()
	assert.Nil(t, err)

	err = handler.SetupPVC("longhorn", "1Gi")
	assert.Nil(t, err)

	spec := createTestJobSpec()
	fmt.Println(spec)
	yaml, err := handler.ComposeJobYAML(&spec)
	assert.Nil(t, err)
	fmt.Println(yaml)

	jobPodNames, err := handler.CreateJob(yaml, &spec)
	assert.Nil(t, err)
	assert.True(t, len(jobPodNames) > 1)

	_, err = handler.GetLog(jobPodNames[0]+"error", spec.JobContainerName, true)
	assert.NotNil(t, err)

	_, err = handler.GetLog(jobPodNames[0], spec.JobContainerName+"error", true)
	assert.NotNil(t, err)

	err = handler.DeleteNamespace()
	assert.Nil(t, err)
}

func TestK8sHandlerNamespace(t *testing.T) {
	namespace := "testnamespace6"
	handler, err := CreateK8sHandler("testexecutor", namespace, namespace+"kubeexecutor-pvc")
	assert.Nil(t, err)

	err = handler.CreateNamespace()
	assert.Nil(t, err)

	err = handler.SetupPVC("longhorn", "1Gi")
	assert.Nil(t, err)

	namespaceNames, err := handler.GetNamespaces()
	assert.Nil(t, err)

	found := false
	for _, namespaceName := range namespaceNames {
		if namespaceName == "testnamespace6" {
			found = true
		}
	}

	assert.True(t, found)

	err = handler.DeleteNamespace()
	assert.Nil(t, err)
}

func TestK8sHandlerDeploymentNames(t *testing.T) {
	spec := createTestDeploymentSpec()
	spec.DeploymentName = "executor-deployment"

	namespace := "testnamespace7"
	handler, err := CreateK8sHandler("testexecutor", namespace, namespace+"kubeexecutor-pvc")
	assert.Nil(t, err)

	err = handler.CreateNamespace()
	assert.Nil(t, err)

	err = handler.SetupPVC("longhorn", "1Gi")
	assert.Nil(t, err)

	deploymentYAML, err := handler.ComposeDeploymentYAML(spec, DeploymentName)
	assert.Nil(t, err)

	err = handler.CreateDeployment(deploymentYAML)
	assert.Nil(t, err)

	time.Sleep(deploymentWaittime * time.Second)

	deploymentNames, err := handler.GetDeploymentNames()
	assert.Nil(t, err)

	assert.Len(t, deploymentNames, 1)
	assert.Equal(t, deploymentNames[0], spec.DeploymentName)

	err = handler.DeleteNamespace()
	assert.Nil(t, err)
}

func TestK8sHandlerGetPodNames(t *testing.T) {
	spec := createTestDeploymentSpec()

	namespace := "testnamespace8"
	handler, err := CreateK8sHandler("testexecutor", namespace, namespace+"kubeexecutor-pvc")
	assert.Nil(t, err)

	err = handler.CreateNamespace()
	assert.Nil(t, err)

	err = handler.SetupPVC("longhorn", "1Gi")
	assert.Nil(t, err)

	deploymentYAML, err := handler.ComposeDeploymentYAML(spec, DeploymentName)
	assert.Nil(t, err)

	err = handler.CreateDeployment(deploymentYAML)
	assert.Nil(t, err)

	time.Sleep(deploymentWaittime * time.Second)

	podNames, err := handler.GetPodNames()
	assert.Nil(t, err)
	assert.Len(t, podNames, spec.NumberOfPods)

	err = handler.DeleteNamespace()
	assert.Nil(t, err)
}

func TestK8sHandlerGetContainerNames(t *testing.T) {
	spec := createTestDeploymentSpec()

	namespace := "testnamespace9"
	handler, err := CreateK8sHandler("testexecutor", namespace, namespace+"kubeexecutor-pvc")
	assert.Nil(t, err)

	err = handler.CreateNamespace()
	assert.Nil(t, err)

	err = handler.SetupPVC("longhorn", "1Gi")
	assert.Nil(t, err)

	deploymentYAML, err := handler.ComposeDeploymentYAML(spec, DeploymentName)
	assert.Nil(t, err)

	err = handler.CreateDeployment(deploymentYAML)
	assert.Nil(t, err)

	time.Sleep(deploymentWaittime * time.Second)

	podNames, err := handler.GetPodNames()
	assert.Nil(t, err)

	for _, podName := range podNames {
		containerNames, err := handler.GetContainerNames(podName)
		assert.Nil(t, err)
		assert.Len(t, containerNames, spec.ExecutorsPerPod)
	}

	err = handler.DeleteNamespace()
	assert.Nil(t, err)
}

func TestK8sHandlerGetContainerOutput(t *testing.T) {
	spec := createTestDeploymentSpec()
	spec.NumberOfPods = 1

	namespace := "testnamespace10"
	handler, err := CreateK8sHandler("testexecutor", namespace, namespace+"kubeexecutor-pvc")
	assert.Nil(t, err)

	err = handler.CreateNamespace()
	assert.Nil(t, err)

	err = handler.SetupPVC("longhorn", "1Gi")
	assert.Nil(t, err)

	deploymentYAML, err := handler.ComposeDeploymentYAML(spec, DeploymentName)
	assert.Nil(t, err)

	err = handler.CreateDeployment(deploymentYAML)
	assert.Nil(t, err)

	time.Sleep(deploymentWaittime * time.Second)

	podNames, err := handler.GetPodNames()
	assert.Nil(t, err)
	assert.Len(t, podNames, spec.NumberOfPods)

	podName := podNames[0]
	err = handler.WaitForPod(podName, "modelresource-downloader")
	assert.Nil(t, err)

	go func() {
		handler.PrintLogs(podName, "modelresource-downloader", false)
	}()

	time.Sleep(10 * time.Second)

	err = handler.DeleteNamespace()
	assert.Nil(t, err)
}

func TestK8sHandlerRestartContainer(t *testing.T) {
	spec := createTestDeploymentSpec()
	spec.NumberOfPods = 1

	namespace := "testnamespace11"
	handler, err := CreateK8sHandler("testexecutor", namespace, namespace+"kubeexecutor-pvc")
	assert.Nil(t, err)

	err = handler.CreateNamespace()
	assert.Nil(t, err)

	err = handler.SetupPVC("longhorn", "1Gi")
	assert.Nil(t, err)

	deploymentYAML, err := handler.ComposeDeploymentYAML(spec, DeploymentName)
	assert.Nil(t, err)

	err = handler.CreateDeployment(deploymentYAML)
	assert.Nil(t, err)

	time.Sleep(deploymentWaittime * time.Second)

	podNames, err := handler.GetPodNames()
	assert.Nil(t, err)
	assert.Len(t, podNames, spec.NumberOfPods)

	podName := podNames[0]
	err = handler.WaitForPod(podName, "executor-0")
	assert.Nil(t, err)

	err = handler.RestartPod(podName)

	err = handler.DeleteNamespace()
	assert.Nil(t, err)
}

func TestK8sHandlerScale(t *testing.T) {
	spec := createTestDeploymentSpec()
	spec.DeploymentName = "executor-deployment"
	spec.NumberOfPods = 1

	namespace := "testnamespace12"
	handler, err := CreateK8sHandler("testexecutor", namespace, namespace+"kubeexecutor-pvc")
	assert.Nil(t, err)

	err = handler.CreateNamespace()
	assert.Nil(t, err)

	err = handler.SetupPVC("longhorn", "1Gi")
	assert.Nil(t, err)

	deploymentYAML, err := handler.ComposeDeploymentYAML(spec, DeploymentName)
	assert.Nil(t, err)

	err = handler.CreateDeployment(deploymentYAML)
	assert.Nil(t, err)

	time.Sleep(deploymentWaittime * time.Second)

	replicas, err := handler.GetScale(spec.DeploymentName)
	assert.Nil(t, err)
	assert.Equal(t, replicas, spec.NumberOfPods)

	err = handler.SetScale(2, spec.DeploymentName)
	assert.Nil(t, err)

	replicas, err = handler.GetScale(spec.DeploymentName)
	assert.Nil(t, err)
	assert.Equal(t, replicas, 2)

	err = handler.DeleteNamespace()
	assert.Nil(t, err)
}

func TestK8sHandlerDockerReg(t *testing.T) {
	namespace := "testnamespace13"
	handler, err := CreateK8sHandler("testexecutor", namespace, namespace+"kubeexecutor-pvc")
	assert.Nil(t, err)

	err = handler.CreateNamespace()
	assert.Nil(t, err)

	err = handler.SetupPVC("longhorn", "1Gi")
	assert.Nil(t, err)

	username := "user"
	password := "secret"
	regURL := "https://registry.colonyos.io"

	secret := CreateDockerRegistrySecret(username, password, regURL)
	err = handler.CreateDockerRegistrySecret(secret)
	assert.Nil(t, err)

	err = handler.DeleteNamespace()
	assert.Nil(t, err)
}

func TestK8sHandlerDeleteDeployment(t *testing.T) {
	namespace := "testnamespace14"
	handler, err := CreateK8sHandler("testexecutor", namespace, namespace+"kubeexecutor-pvc")
	assert.Nil(t, err)

	err = handler.CreateNamespace()
	assert.Nil(t, err)

	err = handler.SetupPVC("longhorn", "1Gi")
	assert.Nil(t, err)

	spec := createTestDeploymentSpec()
	spec.DeploymentName = "executor-deployment"

	deploymentYAML, err := handler.ComposeDeploymentYAML(spec, DeploymentName)
	assert.Nil(t, err)

	err = handler.CreateDeployment(deploymentYAML)
	assert.Nil(t, err)

	time.Sleep(deploymentWaittime * time.Second)

	deploymentNames, err := handler.GetDeploymentNames()
	assert.Nil(t, err)
	assert.Len(t, deploymentNames, 1)

	err = handler.DeleteDeployment(DeploymentName)
	assert.Nil(t, err)

	time.Sleep(deploymentWaittime * time.Second)

	deploymentNames, err = handler.GetDeploymentNames()
	assert.Nil(t, err)
	assert.Len(t, deploymentNames, 0)

	err = handler.DeleteNamespace()
	assert.Nil(t, err)
}

func TestK8sHandlerCreatePVC(t *testing.T) {
	namespace := "testnamespace15"
	handler, err := CreateK8sHandler("testexecutor", namespace, namespace+"kubeexecutor-pvc")
	assert.Nil(t, err)

	err = handler.CreateNamespace()
	assert.Nil(t, err)

	err = handler.SetupPVC("longhorn", "1Gi")
	assert.Nil(t, err)

	spec := createTestPVCSpec()
	fmt.Println(spec)
	yaml, err := handler.ComposePVCYAML(spec)
	assert.Nil(t, err)
	fmt.Println(yaml)

	pvcExists, err := handler.DoesPVCExist(spec.PVCName)
	assert.Nil(t, err)
	assert.False(t, pvcExists)

	err = handler.CreatePVC(yaml)
	assert.Nil(t, err)

	pvcExists, err = handler.DoesPVCExist(spec.PVCName)
	assert.Nil(t, err)
	assert.True(t, pvcExists)

	err = handler.DeleteNamespace()
	assert.Nil(t, err)
}
