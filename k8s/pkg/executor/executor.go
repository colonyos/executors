package executor

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/colonyos/colonies/pkg/client"
	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/security/crypto"
	"github.com/colonyos/executors/k8s/pkg/k8s"
	log "github.com/sirupsen/logrus"
)

type Executor struct {
	coloniesServerHost string
	coloniesServerPort int
	coloniesInsecure   bool
	colonyID           string
	colonyPrvKey       string
	executorID         string
	executorPrvKey     string
	executorNamespace  string
	ctx                context.Context
	cancel             context.CancelFunc
	client             *client.ColoniesClient
}

type ExecutorOption func(*Executor)

func WithColoniesServerHost(host string) ExecutorOption {
	return func(e *Executor) {
		e.coloniesServerHost = host
	}
}

func WithColoniesServerPort(port int) ExecutorOption {
	return func(e *Executor) {
		e.coloniesServerPort = port
	}
}

func WithColoniesInsecure(insecure bool) ExecutorOption {
	return func(e *Executor) {
		e.coloniesInsecure = insecure
	}
}

func WithColonyID(id string) ExecutorOption {
	return func(e *Executor) {
		e.colonyID = id
	}
}

func WithColonyPrvKey(prvkey string) ExecutorOption {
	return func(e *Executor) {
		e.colonyPrvKey = prvkey
	}
}

func WithExecutorID(id string) ExecutorOption {
	return func(e *Executor) {
		e.executorID = id
	}
}

func WithExecutorPrvKey(key string) ExecutorOption {
	return func(e *Executor) {
		e.executorPrvKey = key
	}
}

func WithExecutorNamespace(namespace string) ExecutorOption {
	return func(e *Executor) {
		e.executorNamespace = namespace
	}
}

func createExecutorWithKey(colonyID string) (*core.Executor, string, string, error) {
	crypto := crypto.CreateCrypto()
	executorPrvKey, err := crypto.GeneratePrivateKey()
	if err != nil {
		return nil, "", "", err
	}

	executorID, err := crypto.GenerateID(executorPrvKey)
	if err != nil {
		return nil, "", "", err
	}

	return core.CreateExecutor(executorID, "k8s", core.GenerateRandomID(), colonyID, time.Now(), time.Now()), executorID, executorPrvKey, nil
}

func CreateExecutor(opts ...ExecutorOption) (*Executor, error) {
	e := &Executor{}
	for _, opt := range opts {
		opt(e)
	}

	ctx, cancel := context.WithCancel(context.Background())
	e.ctx = ctx
	e.cancel = cancel

	sigc := make(chan os.Signal)
	signal.Notify(sigc, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT, syscall.SIGSEGV)
	go func() {
		<-sigc
		e.Shutdown()
		os.Exit(1)
	}()

	e.client = client.CreateColoniesClient(e.coloniesServerHost, e.coloniesServerPort, e.coloniesInsecure, false)

	if e.colonyPrvKey != "" {
		spec, executorID, executorPrvKey, err := createExecutorWithKey(e.colonyID)
		if err != nil {
			return nil, err
		}
		e.executorID = executorID
		e.executorPrvKey = executorPrvKey

		_, err = e.client.AddExecutor(spec, e.colonyPrvKey)
		if err != nil {
			return nil, err
		}
		err = e.client.ApproveExecutor(e.executorID, e.colonyPrvKey)
		if err != nil {
			return nil, err
		}

		function := &core.Function{ExecutorID: e.executorID, ColonyID: e.colonyID, FuncName: "deploy", Desc: "Deploy executors", Args: []string{"deploymentname::string, pods::int, executorperpod::int, ramdisk::bool, containerimage::string"}}
		_, err = e.client.AddFunction(function, e.executorPrvKey)
		if err != nil {
			return nil, err
		}

		function = &core.Function{ExecutorID: e.executorID, ColonyID: e.colonyID, FuncName: "undeploy", Desc: "Undeploy executors", Args: []string{"deploymentname::string, pods::int"}}
		_, err = e.client.AddFunction(function, e.executorPrvKey)
		if err != nil {
			return nil, err
		}

		log.WithFields(log.Fields{"ExecutorID": e.executorID}).Info("Self-registered")
	}

	return e, nil
}

func (e *Executor) Shutdown() error {
	log.Info("Shutting down")
	if e.colonyPrvKey != "" {
		err := e.client.DeleteExecutor(e.executorID, e.colonyPrvKey)
		if err != nil {
			log.WithFields(log.Fields{"ExecutorID": e.executorID}).Warning("Failed to deregistered")
		}

		log.WithFields(log.Fields{"ExecutorID": e.executorID}).Info("Deregistered")
	}
	e.cancel()
	return nil
}

func (e *Executor) deployExecutor(process *core.Process) {
	var err error
	if len(process.FunctionSpec.KwArgs) != 5 {
		log.Info(err)
		err = e.client.Fail(process.ID, []string{"Invalid argument"}, e.executorPrvKey)
		return
	}
	deploymentNameIf := process.FunctionSpec.KwArgs["deployment_name"]
	deploymentName, ok := deploymentNameIf.(string)
	if !ok {
		log.Warning(err)
		err = e.client.Fail(process.ID, []string{"Invalid argument, deploymentName is not a string"}, e.executorPrvKey)
		return
	}

	podsIf := process.FunctionSpec.KwArgs["pods"]
	podsFloat, ok := podsIf.(float64)
	if !ok {
		log.Warning("pods is not a int")
		err = e.client.Fail(process.ID, []string{"Invalid argument, pods is not a int"}, e.executorPrvKey)
		return
	}
	pods := int(podsFloat)

	executorsIf := process.FunctionSpec.KwArgs["executors_per_pod"]
	executorsFloat, ok := executorsIf.(float64)
	if !ok {
		log.Warning("executors is not a int")
		err = e.client.Fail(process.ID, []string{"Invalid argument, executors is not a int"}, e.executorPrvKey)
		return
	}
	executors := int(executorsFloat)

	ramdiskIf := process.FunctionSpec.KwArgs["ramdisk"]
	ramdisk, ok := ramdiskIf.(bool)
	if !ok {
		log.Warning("ramdisk is not a bool")
		err = e.client.Fail(process.ID, []string{"Invalid argument, ramdisk is not a bool"}, e.executorPrvKey)
		return
	}

	dockerImageIf := process.FunctionSpec.KwArgs["image"]
	dockerImage, ok := dockerImageIf.(string)
	if !ok {
		log.Warning("dockerImage is not a string")
		err = e.client.Fail(process.ID, []string{"Invalid argument, dockerImage is not a string"}, e.executorPrvKey)
		return
	}

	log.WithFields(log.Fields{
		"ProcessID":      process.ID,
		"ExecutorID":     e.executorID,
		"DockerImage":    dockerImage,
		"Namespace":      e.executorNamespace,
		"Ramdisk":        ramdisk,
		"Pods":           pods,
		"Executors":      executors,
		"DeploymentName": deploymentName}).
		Info("Executing Deploy function")

	deploymentSpec := k8s.DeploymentSpec{
		TestMode:               false,
		DeploymentName:         deploymentName,
		Namespace:              e.executorNamespace,
		NumberOfPods:           pods,
		ExecutorsPerPod:        executors,
		ColoniesTLS:            !e.coloniesInsecure,
		ColoniesServerHost:     e.coloniesServerHost,
		ColoniesServerPort:     e.coloniesServerPort,
		ColoniesColonyID:       e.colonyID,
		ColoniesColonyPrvKey:   e.colonyPrvKey,
		ColoniesExecutorID:     e.executorID,
		ColoniesExecutorPrvKey: e.executorPrvKey,
		EnableRamdisk:          false,
		RamdiskSize:            "",
		DockerImage:            dockerImage,
		DockerRegistryURL:      "",
		DockerRegistryUsername: "",
		DockerRegistryPassword: "",
	}

	handler, err := k8s.CreateK8sHandler(e.executorNamespace)
	if err != nil {
		log.Warning("failed to create k8s handler")
		err = e.client.Fail(process.ID, []string{"Failed to create k8s handler"}, e.executorPrvKey)
		return
	}

	deploymentYAML, err := handler.ComposeDeploymentYAML(deploymentSpec, deploymentName)
	if err != nil {
		log.Warning("failed to create k8s handler")
		err = e.client.Fail(process.ID, []string{"Failed to create k8s handler"}, e.executorPrvKey)
		return
	}

	err = handler.CreateDeployment(deploymentYAML)
	if err != nil {
		log.Warning("failed to create deployment")
		log.Warning(err)
		err = e.client.Fail(process.ID, []string{"Failed to create deployment"}, e.executorPrvKey)
		return
	}

	err = e.client.Close(process.ID, e.executorPrvKey)
	log.Info("Closing process")
}

func (e *Executor) undeploy(process *core.Process) {
	var err error
	if len(process.FunctionSpec.Args) != 1 {
		log.Info(err)
		err = e.client.Fail(process.ID, []string{"Invalid argument"}, e.executorPrvKey)
		return
	}

	deploymentNameIf := process.FunctionSpec.Args[0]
	deploymentName, ok := deploymentNameIf.(string)
	if !ok {
		log.Warning(err)
		err = e.client.Fail(process.ID, []string{"Invalid argument, deploymentName is not a string"}, e.executorPrvKey)
		return
	}

	handler, err := k8s.CreateK8sHandler(e.executorNamespace)
	if err != nil {
		log.Warning("failed to create k8s handler")
		err = e.client.Fail(process.ID, []string{"Failed to create k8s handler"}, e.executorPrvKey)
		return
	}

	log.WithFields(log.Fields{
		"ProcessID":      process.ID,
		"ExecutorID":     e.executorID,
		"Namespace":      e.executorNamespace,
		"DeploymentName": deploymentName}).
		Info("Executing Undeploy function")

	err = handler.DeleteDeployment(deploymentName)
	if err != nil {
		log.Warning("failed to create k8s handler")
		log.Warning(err)
		err = e.client.Fail(process.ID, []string{"Failed to create k8s handler"}, e.executorPrvKey)
		return
	}

	err = e.client.Close(process.ID, e.executorPrvKey)
	log.Info("Closing process")
}

func (e *Executor) scale(process *core.Process) {
	var err error
	if len(process.FunctionSpec.Args) != 2 {
		log.Info(err)
		err = e.client.Fail(process.ID, []string{"Invalid argument"}, e.executorPrvKey)
		return
	}

	deploymentNameIf := process.FunctionSpec.Args[0]
	deploymentName, ok := deploymentNameIf.(string)
	if !ok {
		log.Warning(err)
		err = e.client.Fail(process.ID, []string{"Invalid argument, deploymentName is not a string"}, e.executorPrvKey)
		return
	}

	podsIf := process.FunctionSpec.Args[1]
	podsFloat, ok := podsIf.(float64)
	if !ok {
		log.Warning("pods is not a int")
		err = e.client.Fail(process.ID, []string{"Invalid argument, pods is not a int"}, e.executorPrvKey)
		return
	}
	pods := int(podsFloat)

	handler, err := k8s.CreateK8sHandler(e.executorNamespace)
	if err != nil {
		log.Warning("failed to create k8s handler")
		err = e.client.Fail(process.ID, []string{"Failed to create k8s handler"}, e.executorPrvKey)
		return
	}

	log.WithFields(log.Fields{
		"ProcessID":      process.ID,
		"ExecutorID":     e.executorID,
		"Namespace":      e.executorNamespace,
		"Pods":           pods,
		"DeploymentName": deploymentName}).
		Info("Executing Scale function")

	err = handler.SetScale(pods, deploymentName)
	if err != nil {
		log.Warning("failed to create k8s handler")
		log.Warning(err)
		err = e.client.Fail(process.ID, []string{"Failed to create k8s handler"}, e.executorPrvKey)
		return
	}

	err = e.client.Close(process.ID, e.executorPrvKey)
	log.Info("Closing process")
}

func (e *Executor) getScale(process *core.Process) {
	var err error
	if len(process.FunctionSpec.Args) != 1 {
		log.Info(err)
		err = e.client.Fail(process.ID, []string{"Invalid argument"}, e.executorPrvKey)
		return
	}

	deploymentNameIf := process.FunctionSpec.Args[0]
	deploymentName, ok := deploymentNameIf.(string)
	if !ok {
		log.Warning(err)
		err = e.client.Fail(process.ID, []string{"Invalid argument, deploymentName is not a string"}, e.executorPrvKey)
		return
	}

	handler, err := k8s.CreateK8sHandler(e.executorNamespace)
	if err != nil {
		log.Warning("failed to create k8s handler")
		err = e.client.Fail(process.ID, []string{"Failed to create k8s handler"}, e.executorPrvKey)
		return
	}

	log.WithFields(log.Fields{
		"ProcessID":      process.ID,
		"ExecutorID":     e.executorID,
		"Namespace":      e.executorNamespace,
		"DeploymentName": deploymentName}).
		Info("Executing GetScale function")

	scale, err := handler.GetScale(deploymentName)
	if err != nil {
		log.Warning("failed to create k8s handler")
		log.Warning(err)
		err = e.client.Fail(process.ID, []string{"Failed to create k8s handler"}, e.executorPrvKey)
		return
	}

	output := make([]interface{}, 1)
	output[0] = scale
	err = e.client.CloseWithOutput(process.ID, output, e.executorPrvKey)
	log.Info("Closing process")
}

func (e *Executor) getDeployments(process *core.Process) {
	var err error
	if len(process.FunctionSpec.Args) != 0 {
		log.Info(err)
		err = e.client.Fail(process.ID, []string{"Invalid argument"}, e.executorPrvKey)
		return
	}

	handler, err := k8s.CreateK8sHandler(e.executorNamespace)
	if err != nil {
		log.Warning("failed to create k8s handler")
		err = e.client.Fail(process.ID, []string{"Failed to create k8s handler"}, e.executorPrvKey)
		return
	}

	log.WithFields(log.Fields{
		"ProcessID":  process.ID,
		"ExecutorID": e.executorID,
		"Namespace":  e.executorNamespace}).
		Info("Executing GetDeploymentNames function")

	deploymentNames, err := handler.GetDeploymentNames()
	if err != nil {
		log.Warning("failed to create k8s handler")
		log.Warning(err)
		err = e.client.Fail(process.ID, []string{"Failed to create k8s handler"}, e.executorPrvKey)
		return
	}
	json, err := json.Marshal(deploymentNames)
	if err != nil {
		log.Warning("failed to marshal json")
		log.Warning(err)
		err = e.client.Fail(process.ID, []string{"Failed to marshaljson "}, e.executorPrvKey)
		return
	}

	output := make([]interface{}, 1)
	output[0] = string(json)
	err = e.client.CloseWithOutput(process.ID, output, e.executorPrvKey)
	log.Info("Closing process")
}

func (e *Executor) getPods(process *core.Process) {
	var err error
	if len(process.FunctionSpec.Args) != 0 {
		log.Info(err)
		err = e.client.Fail(process.ID, []string{"Invalid argument"}, e.executorPrvKey)
		return
	}

	handler, err := k8s.CreateK8sHandler(e.executorNamespace)
	if err != nil {
		log.Warning("failed to create k8s handler")
		err = e.client.Fail(process.ID, []string{"Failed to create k8s handler"}, e.executorPrvKey)
		return
	}

	log.WithFields(log.Fields{
		"ProcessID":  process.ID,
		"ExecutorID": e.executorID,
		"Namespace":  e.executorNamespace}).
		Info("Executing GetPodNames function")

	podNames, err := handler.GetPodNames()
	if err != nil {
		log.Warning("failed to create k8s handler")
		log.Warning(err)
		err = e.client.Fail(process.ID, []string{"Failed to create k8s handler"}, e.executorPrvKey)
		return
	}
	json, err := json.Marshal(podNames)
	if err != nil {
		log.Warning("failed to marshal json")
		log.Warning(err)
		err = e.client.Fail(process.ID, []string{"Failed to marshaljson "}, e.executorPrvKey)
		return
	}

	output := make([]interface{}, 1)
	output[0] = string(json)
	err = e.client.CloseWithOutput(process.ID, output, e.executorPrvKey)
	log.Info("Closing process")
}

func (e *Executor) getNumberOfPods(process *core.Process) {
	var err error
	if len(process.FunctionSpec.Args) != 0 {
		log.Info(err)
		err = e.client.Fail(process.ID, []string{"Invalid argument"}, e.executorPrvKey)
		return
	}

	handler, err := k8s.CreateK8sHandler(e.executorNamespace)
	if err != nil {
		log.Warning("failed to create k8s handler")
		err = e.client.Fail(process.ID, []string{"Failed to create k8s handler"}, e.executorPrvKey)
		return
	}

	log.WithFields(log.Fields{
		"ProcessID":  process.ID,
		"ExecutorID": e.executorID,
		"Namespace":  e.executorNamespace}).
		Info("Executing GetPodNames function")

	podNames, err := handler.GetPodNames()
	if err != nil {
		log.Warning("failed to create k8s handler")
		log.Warning(err)
		err = e.client.Fail(process.ID, []string{"Failed to create k8s handler"}, e.executorPrvKey)
		return
	}
	count := len(podNames)
	json, err := json.Marshal(count)
	if err != nil {
		log.Warning("failed to marshal json")
		log.Warning(err)
		err = e.client.Fail(process.ID, []string{"Failed to marshaljson "}, e.executorPrvKey)
		return
	}

	output := make([]interface{}, 1)
	output[0] = string(json)
	err = e.client.CloseWithOutput(process.ID, output, e.executorPrvKey)
	log.Info("Closing process")
}

func (e *Executor) getContainers(process *core.Process) {
	var err error
	if len(process.FunctionSpec.Args) != 1 {
		log.Info(err)
		err = e.client.Fail(process.ID, []string{"Invalid argument"}, e.executorPrvKey)
		return
	}

	podNameIf := process.FunctionSpec.Args[0]
	podName, ok := podNameIf.(string)
	if !ok {
		log.Warning(err)
		err = e.client.Fail(process.ID, []string{"Invalid argument, podname is not a string"}, e.executorPrvKey)
		return
	}

	handler, err := k8s.CreateK8sHandler(e.executorNamespace)
	if err != nil {
		log.Warning("failed to create k8s handler")
		err = e.client.Fail(process.ID, []string{"Failed to create k8s handler"}, e.executorPrvKey)
		return
	}

	log.WithFields(log.Fields{
		"ProcessID":  process.ID,
		"ExecutorID": e.executorID,
		"PodName":    podName,
		"Namespace":  e.executorNamespace}).
		Info("Executing GetContainerNames function")

	containerNames, err := handler.GetContainerNames(podName)
	if err != nil {
		log.Warning("failed to create k8s handler")
		log.Warning(err)
		err = e.client.Fail(process.ID, []string{"Failed to create k8s handler"}, e.executorPrvKey)
		return
	}
	json, err := json.Marshal(containerNames)
	if err != nil {
		log.Warning("failed to marshal json")
		log.Warning(err)
		err = e.client.Fail(process.ID, []string{"Failed to marshaljson "}, e.executorPrvKey)
		return
	}

	output := make([]interface{}, 1)
	output[0] = string(json)
	err = e.client.CloseWithOutput(process.ID, output, e.executorPrvKey)
	log.Info("Closing process")
}

func (e *Executor) getNumberOfContainers(process *core.Process) {
	var err error
	if len(process.FunctionSpec.Args) != 1 {
		log.Info(err)
		err = e.client.Fail(process.ID, []string{"Invalid argument"}, e.executorPrvKey)
		return
	}

	podNameIf := process.FunctionSpec.Args[0]
	podName, ok := podNameIf.(string)
	if !ok {
		log.Warning(err)
		err = e.client.Fail(process.ID, []string{"Invalid argument, podname is not a string"}, e.executorPrvKey)
		return
	}

	handler, err := k8s.CreateK8sHandler(e.executorNamespace)
	if err != nil {
		log.Warning("failed to create k8s handler")
		err = e.client.Fail(process.ID, []string{"Failed to create k8s handler"}, e.executorPrvKey)
		return
	}

	log.WithFields(log.Fields{
		"ProcessID":  process.ID,
		"ExecutorID": e.executorID,
		"PodName":    podName,
		"Namespace":  e.executorNamespace}).
		Info("Executing GetContainerNames function")

	containerNames, err := handler.GetContainerNames(podName)
	if err != nil {
		log.Warning("failed to create k8s handler")
		log.Warning(err)
		err = e.client.Fail(process.ID, []string{"Failed to create k8s handler"}, e.executorPrvKey)
		return
	}

	output := make([]interface{}, 1)
	output[0] = len(containerNames)
	err = e.client.CloseWithOutput(process.ID, output, e.executorPrvKey)
	log.Info("Closing process")
}

func (e *Executor) restart(process *core.Process) {
	var err error
	if len(process.FunctionSpec.Args) != 1 {
		log.Info(err)
		err = e.client.Fail(process.ID, []string{"Invalid argument"}, e.executorPrvKey)
		return
	}
	podNameIf := process.FunctionSpec.Args[0]
	podName, ok := podNameIf.(string)
	if !ok {
		log.Warning(err)
		err = e.client.Fail(process.ID, []string{"Invalid argument, deploymentName is not a string"}, e.executorPrvKey)
		return
	}

	log.WithFields(log.Fields{
		"ProcessID":  process.ID,
		"ExecutorID": e.executorID,
		"Namespace":  e.executorNamespace,
		"PodName":    podName}).
		Info("Executing Restart function")

	handler, err := k8s.CreateK8sHandler(e.executorNamespace)
	if err != nil {
		log.Warning("failed to create k8s handler")
		err = e.client.Fail(process.ID, []string{"Failed to create k8s handler"}, e.executorPrvKey)
		return
	}

	err = handler.RestartPod(podName)
	if err != nil {
		log.Warning("failed to restart pod")
		log.Warning(err)
		err = e.client.Fail(process.ID, []string{"Failed to restart pod"}, e.executorPrvKey)
		return
	}

	err = e.client.Close(process.ID, e.executorPrvKey)
	log.Info("Closing process")
}

func (e *Executor) ServeForEver() error {
	for {
		process, err := e.client.AssignWithContext(e.colonyID, 100, e.ctx, e.executorPrvKey)
		if err != nil {
			var coloniesError *core.ColoniesError
			if errors.As(err, &coloniesError) {
				if coloniesError.Status == 404 { // No processes can be selected for executor
					log.Info(err)
					continue
				}
			}

			log.Error(err)
			log.Error("Retrying in 5 seconds ...")
			time.Sleep(5 * time.Second)

			continue
		}

		log.WithFields(log.Fields{"ProcessID": process.ID, "ExecutorID": e.executorID}).Info("Assigned process to executor")

		switch funcName := process.FunctionSpec.FuncName; funcName {
		case "deploy_executor":
			e.deployExecutor(process)
		case "undeploy":
			e.undeploy(process)
		case "scale":
			e.scale(process)
		case "get_scale":
			e.getScale(process)
		case "get_deployments":
			e.getDeployments(process)
		case "get_pods":
			e.getPods(process)
		case "pods":
			e.getNumberOfPods(process)
		case "get_containers":
			e.getContainers(process)
		case "containers":
			e.getNumberOfContainers(process)
		case "restart":
			e.restart(process)
		default:
			log.WithFields(log.Fields{"ProcessID": process.ID, "ExecutorID": e.executorID, "FuncName": funcName}).Info("Unsupported function")
			err = e.client.Fail(process.ID, []string{"Unsupported function: " + funcName}, e.executorPrvKey)
			log.Info(err)
		}
	}
}
