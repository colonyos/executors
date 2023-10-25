package k8s

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const deploymentWaittime = 10 // 5 seconds

func TestK8sHandlerComposeDeployment(t *testing.T) {
	handler, err := CreateK8sHandler("testexecutor", "testnamespace0")
	assert.Nil(t, err)

	spec := createTestDeploymentSpec()
	yaml, err := handler.ComposeDeploymentYAML(spec, DeploymentName)
	assert.Nil(t, err)
	fmt.Println(yaml)
}

func TestK8sHandlerComposeJob(t *testing.T) {
	handler, err := CreateK8sHandler("testexecutor", "testnamespace0")
	assert.Nil(t, err)

	spec := createTestJobSpec()
	fmt.Println(spec)
	yaml, podName, err := handler.ComposeJobYAML(spec)
	fmt.Println(err)
	assert.Nil(t, err)
	fmt.Println(yaml)
	fmt.Println(podName)
}

func TestK8sHandlerCreateJob(t *testing.T) {
	handler, err := CreateK8sHandler("testexecutor", "testnamespace-job299")
	assert.Nil(t, err)

	err = handler.CreateNamespace()
	assert.Nil(t, err)

	spec := createTestJobSpec()
	fmt.Println(spec)
	yaml, jobName, err := handler.ComposeJobYAML(spec)
	assert.Nil(t, err)
	fmt.Println(yaml)
	fmt.Println(jobName)

	jobPodNames, err := handler.CreateJob(yaml, jobName, &spec)
	assert.Nil(t, err)
	fmt.Println(jobPodNames)

	jobNames, err := handler.GetJobNames()
	assert.Nil(t, err)
	assert.Len(t, jobNames, 1)
	assert.Equal(t, jobNames[0], jobName)

	for _, jobPodName := range jobPodNames {
		err = handler.PrintAllLogs(jobPodName, true)
		assert.Nil(t, err)
	}

	err = handler.DeleteNamespace()
	assert.Nil(t, err)
}

func TestK8sHandlerGetLogsInvalidContainer(t *testing.T) {
	handler, err := CreateK8sHandler("testexecutor", "testnamespace-invalidlog8")
	assert.Nil(t, err)

	err = handler.CreateNamespace()
	assert.Nil(t, err)

	spec := createTestJobSpec()
	fmt.Println(spec)
	yaml, jobName, err := handler.ComposeJobYAML(spec)
	assert.Nil(t, err)
	fmt.Println(yaml)
	fmt.Println(jobName)

	jobPodNames, err := handler.CreateJob(yaml, jobName, &spec)
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
	handler, err := CreateK8sHandler("testexecutor", "testnamespace1")
	assert.Nil(t, err)

	err = handler.CreateNamespace()
	assert.Nil(t, err)

	namespaceNames, err := handler.GetNamespaces()
	assert.Nil(t, err)

	found := false
	for _, namespaceName := range namespaceNames {
		if namespaceName == "testnamespace1" {
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

	handler, err := CreateK8sHandler("testexecutor", "testnamespace2")
	assert.Nil(t, err)

	err = handler.CreateNamespace()
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

	handler, err := CreateK8sHandler("testexecutor", "testnamespace3")
	assert.Nil(t, err)

	err = handler.CreateNamespace()
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

	handler, err := CreateK8sHandler("testexecutor", "testnamespace4")
	assert.Nil(t, err)

	err = handler.CreateNamespace()
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

	handler, err := CreateK8sHandler("testexecutor", "testnamespace5")
	assert.Nil(t, err)

	err = handler.CreateNamespace()
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

	handler, err := CreateK8sHandler("testexecutor", "testnamespace6")
	assert.Nil(t, err)

	err = handler.CreateNamespace()
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

	handler, err := CreateK8sHandler("testexecutor", "testnamespace7")
	assert.Nil(t, err)

	err = handler.CreateNamespace()
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
	handler, err := CreateK8sHandler("testexecutor", "testnamespace8")
	assert.Nil(t, err)

	err = handler.CreateNamespace()
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
	handler, err := CreateK8sHandler("testexecutor", "testnamespace9")
	assert.Nil(t, err)

	err = handler.CreateNamespace()
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
