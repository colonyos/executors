package executor

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/colonyos/colonies/pkg/client"
	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/security/crypto"
	"github.com/colonyos/executors/common/pkg/debug"
	"github.com/colonyos/executors/common/pkg/docker"
	"github.com/colonyos/executors/common/pkg/failure"
	"github.com/colonyos/executors/common/pkg/parsers"
	"github.com/colonyos/executors/common/pkg/sync"
	log "github.com/sirupsen/logrus"
)

type Executor struct {
	verbose            bool
	addDebugLogs       bool
	coloniesServerHost string
	coloniesServerPort int
	coloniesInsecure   bool
	colonyName         string
	colonyPrvKey       string
	executorName       string
	executorID         string
	executorPrvKey     string
	executorType       string
	fsDir              string
	swName             string
	swType             string
	swVersion          string
	hwCPU              string
	hwModel            string
	hwNodes            int
	hwMem              string
	hwStorage          string
	hwGPUCount         int
	hwGPUNodesCount    int
	hwGPUName          string
	hwGPUMem           string
	long               float64
	lat                float64
	locDesc            string
	ctx                context.Context
	cancel             context.CancelFunc
	client             *client.ColoniesClient
	syncHandler        *sync.SyncHandler
	failureHandler     *failure.FailureHandler
	debugHandler       *debug.DebugHandler
	dockerHandler      *docker.DockerHandler
	namespace          string
	pvc                string
}

type ExecutorOption func(*Executor)

func WithVerbose(verbose bool) ExecutorOption {
	return func(e *Executor) {
		e.verbose = verbose
	}
}

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

func WithExecutorType(executorType string) ExecutorOption {
	return func(e *Executor) {
		e.executorType = executorType
	}
}

func WithColoniesInsecure(insecure bool) ExecutorOption {
	return func(e *Executor) {
		e.coloniesInsecure = insecure
	}
}

func WithColonyName(name string) ExecutorOption {
	return func(e *Executor) {
		e.colonyName = name
	}
}

func WithColonyPrvKey(prvkey string) ExecutorOption {
	return func(e *Executor) {
		e.colonyPrvKey = prvkey
	}
}

func WithExecutorName(executorName string) ExecutorOption {
	return func(e *Executor) {
		e.executorName = executorName
	}
}

func WithExecutorID(executorID string) ExecutorOption {
	return func(e *Executor) {
		e.executorID = executorID
	}
}

func WithExecutorPrvKey(key string) ExecutorOption {
	return func(e *Executor) {
		e.executorPrvKey = key
	}
}

func WithFsDir(fsDir string) ExecutorOption {
	return func(e *Executor) {
		e.fsDir = fsDir
	}
}

func WithSoftwareName(swName string) ExecutorOption {
	return func(e *Executor) {
		e.swName = swName
	}
}

func WithSoftwareType(swType string) ExecutorOption {
	return func(e *Executor) {
		e.swType = swType
	}
}

func WithSoftwareVersion(swVersion string) ExecutorOption {
	return func(e *Executor) {
		e.swVersion = swVersion
	}
}

func WithHardwareCPU(hwCPU string) ExecutorOption {
	return func(e *Executor) {
		e.hwCPU = hwCPU
	}
}

func WithHardwareModel(hwModel string) ExecutorOption {
	return func(e *Executor) {
		e.hwModel = hwModel
	}
}

func WithHardwareNodes(hwNodes int) ExecutorOption {
	return func(e *Executor) {
		e.hwNodes = hwNodes
	}
}

func WithHardwareMemory(hwMem string) ExecutorOption {
	return func(e *Executor) {
		e.hwMem = hwMem
	}
}

func WithHardwareStorage(hwStorage string) ExecutorOption {
	return func(e *Executor) {
		e.hwStorage = hwStorage
	}
}

func WithHardwareGPUCount(hwGPUCount int) ExecutorOption {
	return func(e *Executor) {
		e.hwGPUCount = hwGPUCount
	}
}

func WithHardwareGPUNodesCount(hwGPUNodesCount int) ExecutorOption {
	return func(e *Executor) {
		e.hwGPUNodesCount = hwGPUNodesCount
	}
}

func WithHardwareGPUName(hwGPUName string) ExecutorOption {
	return func(e *Executor) {
		e.hwGPUName = hwGPUName
	}
}

func WithHardwareGPUMemory(hwGPUMem string) ExecutorOption {
	return func(e *Executor) {
		e.hwGPUMem = hwGPUMem
	}
}

func WithLong(long float64) ExecutorOption {
	return func(e *Executor) {
		e.long = long
	}
}

func WithLat(lat float64) ExecutorOption {
	return func(e *Executor) {
		e.lat = lat
	}
}

func WithLocDesc(locDesc string) ExecutorOption {
	return func(e *Executor) {
		e.locDesc = locDesc
	}
}

func WithK8sNamespace(namespace string) ExecutorOption {
	return func(e *Executor) {
		e.namespace = namespace
	}
}

func WithK8sPVC(pvc string) ExecutorOption {
	return func(e *Executor) {
		e.pvc = pvc
	}
}

func WithAddDebugLogs(addDebugLogs bool) ExecutorOption {
	return func(e *Executor) {
		e.addDebugLogs = addDebugLogs
	}
}

func (e *Executor) createColoniesExecutorWithKey(colonyName string) (*core.Executor, string, string, error) {
	crypto := crypto.CreateCrypto()
	executorPrvKey, err := crypto.GeneratePrivateKey()
	if err != nil {
		return nil, "", "", err
	}

	executorID, err := crypto.GenerateID(executorPrvKey)
	if err != nil {
		return nil, "", "", err
	}

	executor := core.CreateExecutor(executorID, e.executorType, e.executorName, colonyName, time.Now(), time.Now())
	executor.Capabilities.Software.Name = e.swName
	executor.Capabilities.Software.Type = e.swType
	executor.Capabilities.Software.Version = e.swVersion
	executor.Capabilities.Hardware.CPU = e.hwCPU
	executor.Capabilities.Hardware.Model = e.hwModel
	executor.Capabilities.Hardware.Nodes = e.hwNodes
	executor.Capabilities.Hardware.Storage = e.hwStorage
	executor.Capabilities.Hardware.Memory = e.hwMem
	executor.Capabilities.Hardware.GPU.Count = e.hwGPUCount
	executor.Capabilities.Hardware.GPU.NodeCount = e.hwGPUNodesCount
	executor.Capabilities.Hardware.GPU.Name = e.hwGPUName
	executor.Capabilities.Hardware.GPU.Memory = e.hwGPUMem
	executor.Location.Description = e.locDesc
	executor.Location.Long = e.long
	executor.Location.Lat = e.lat

	return executor, executorID, executorPrvKey, nil
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
		spec, executorID, executorPrvKey, err := e.createColoniesExecutorWithKey(e.colonyName)
		if err != nil {
			return nil, err
		}
		e.executorID = executorID
		e.executorPrvKey = executorPrvKey

		_, err = e.client.AddExecutor(spec, e.colonyPrvKey)
		if err != nil {
			return nil, err
		}
		err = e.client.ApproveExecutor(e.colonyName, e.executorName, e.colonyPrvKey)
		if err != nil {
			return nil, err
		}

		log.WithFields(log.Fields{"ColonyName": e.colonyName, "ExecutorName": e.executorName}).Info("Self-registered")
	}

	function := &core.Function{ExecutorName: e.executorName, ColonyName: e.colonyName, FuncName: "execute"}
	e.client.AddFunction(function, e.executorPrvKey)

	var err error
	e.failureHandler, err = failure.CreateFailureHandler(e.executorPrvKey, e.client)
	if err != nil {
		return nil, err
	}

	e.debugHandler, err = debug.CreateDebugHandler(e.executorPrvKey, e.client, e.addDebugLogs)
	if err != nil {
		return nil, err
	}

	e.syncHandler, err = sync.CreateSyncHandler(e.colonyName, e.executorPrvKey, e.client, e.fsDir, e.failureHandler, e.debugHandler)
	if err != nil {
		return nil, err
	}

	e.dockerHandler, err = docker.CreateDockerHandler()
	if err != nil {
		return nil, err
	}

	log.WithFields(log.Fields{
		"Verbose":               e.verbose,
		"ColoniesServerHost":    e.coloniesServerHost,
		"ColoniesServerPort":    e.coloniesServerPort,
		"ColoniesInsecure":      e.coloniesInsecure,
		"K8sPVC":                e.pvc,
		"K8sNamespace":          e.namespace,
		"FsDir":                 e.fsDir,
		"ColonyName":            e.colonyName,
		"ColonyPrvKey":          "***********************",
		"ExecutorId":            e.executorID,
		"ExecutorName":          e.executorName,
		"ExecutorPrvKey":        "***********************",
		"Longitude":             e.long,
		"Latitude":              e.lat,
		"LocationDesc":          e.locDesc,
		"HardwareModel":         e.hwModel,
		"HardwareNodes":         e.hwNodes,
		"HardwareCPU":           e.hwCPU,
		"HardwareMemory":        e.hwMem,
		"HardwareStorage":       e.hwStorage,
		"HardwareGPUName":       e.hwGPUName,
		"HardwareGPUCount":      e.hwGPUCount,
		"HardwareGPUNodesCount": e.hwGPUNodesCount,
		"HardwareGPUMemory":     e.hwGPUMem,
		"SoftwareName":          e.swName,
		"SoftwareVersion":       e.swVersion,
		"SoftwareType":          e.swType,
		"ExecutorType":          e.executorType}).
		Info("Docker Executor started")

	return e, nil
}

func (e *Executor) Shutdown() error {
	log.Info("Shutting down")
	if e.colonyPrvKey != "" {
		err := e.client.RemoveExecutor(e.colonyName, e.executorName, e.colonyPrvKey)
		if err != nil {
			log.WithFields(log.Fields{
				"ExecutorID":   e.executorID,
				"ExecutorName": e.executorName,
				"ColonyName":   e.colonyName}).
				Warning("Failed to deregistered")
		}

		log.WithFields(log.Fields{
			"ExecutorID":   e.executorID,
			"ExecutorName": e.executorName,
			"ColonyName":   e.colonyName}).
			Info("Deregistered")
	}
	e.cancel()
	return nil
}

func (e *Executor) FetchJobLogs(process *core.Process, containerID string) error {
	aggregatedLogsChan := make(chan string, 1000)
	eofChan := make(chan string, 1000)
	errChan := make(chan error, 1000)
	e.dockerHandler.GetContainerLogs(containerID, aggregatedLogsChan, eofChan, errChan)

	for {
		select {
		case msg := <-aggregatedLogsChan:
			fmt.Println("len(msg): ", len(msg))
			log.WithFields(log.Fields{"Log": string(msg)}).Info("Adding log")
			err := e.client.AddLog(process.ID, string(msg), e.executorPrvKey)
			if err != nil {
				log.WithFields(log.Fields{"Error": err}).Error("Failed to add log")
				return err
			}
		case err := <-errChan:
			log.WithFields(log.Fields{"Error": err}).Error("Failed to get logs")
			return err
		case msg := <-eofChan:
			log.WithFields(log.Fields{"Log": string(msg)}).Info("Adding log")
			fmt.Println("len(msg): ", len(msg))
			err := e.client.AddLog(process.ID, string(msg), e.executorPrvKey)
			if err != nil {
				log.WithFields(log.Fields{"Error": err}).Error("Failed to add log")
				return err
			}
			return nil
		}
	}
}

func (e *Executor) executeDocker(process *core.Process) bool {
	err := parsers.ValidateFuncSpec(&process.FunctionSpec)
	if err != nil {
		e.failureHandler.HandleError(process, err, "Failed to validate funcspec")
		return false
	}

	kwArgs, err := parsers.ParseKwArgs(process, e.failureHandler, e.debugHandler)
	if err != nil {
		e.failureHandler.HandleError(process, err, "Failed to parse kwArgs")
		return false
	}

	err = e.syncHandler.PreSync(process, e.debugHandler, e.failureHandler)
	if err != nil {
		e.failureHandler.HandleError(process, err, "Failed to pre sync")
		return false
	}

	cmd := kwArgs.Cmd
	if kwArgs.InitCmd != "" {
		cmd = kwArgs.InitCmd + ";" + kwArgs.Cmd
	}

	logChan := make(chan string, 100)
	doneChan := make(chan string, 100)
	errChan := make(chan error, 1)

	err = e.dockerHandler.PullImage(kwArgs.Image, logChan, doneChan)
	if err != nil {
		e.failureHandler.HandleError(process, err, "Failed to pull image")
		return false
	}

otherloop:
	for {
		select {
		case msg := <-logChan:
			err := e.client.AddLog(process.ID, msg, e.executorPrvKey)
			if err != nil {
				e.failureHandler.HandleError(process, err, "Failed to add log")
				return false
			}
		case err := <-errChan:
			e.failureHandler.HandleError(process, err, "Failed to pull image")
			return false
		case msg := <-doneChan:
			for len(doneChan) > 0 {
				err := e.client.AddLog(process.ID, msg, e.executorPrvKey)
				if err != nil {
					e.failureHandler.HandleError(process, err, "Failed to add log")
					return false
				}
				log.WithFields(log.Fields{"Image": kwArgs.Image}).Info("Image pulled")
			}
			break otherloop
		}
	}

	err = e.client.AddLog(process.ID, "\n", e.executorPrvKey)
	if err != nil {
		e.failureHandler.HandleError(process, err, "Failed to add log")
		return false
	}

	containerID, err := e.dockerHandler.StartContainer(kwArgs.Image, cmd, []string{kwArgs.Args}, e.fsDir)
	if err != nil {
		e.failureHandler.HandleError(process, err, "Failed to start container")
		return false
	}

	err = e.FetchJobLogs(process, containerID)
	if err != nil {
		e.failureHandler.HandleError(process, err, "Failed to get logs")
		return false
	}

	err = e.syncHandler.PostSync(process, e.debugHandler, e.failureHandler, e.fsDir, e.client, e.colonyName, e.executorPrvKey)
	if err != nil {
		e.failureHandler.HandleError(process, err, "Failed to post sync")
		return false
	}

	return true
}

func (e *Executor) ServeForEver() error {
	for {
		process, err := e.client.AssignWithContext(e.colonyName, 100, e.ctx, "", "", e.executorPrvKey)
		if err != nil {
			log.WithFields(log.Fields{"Error": err}).Error("Failed to assign process to executor")
			var coloniesError *core.ColoniesError
			if errors.As(err, &coloniesError) {
				if coloniesError.Status == 404 { // No processes can be selected for executor
					log.Info(err)
					continue
				}
			}

			log.WithFields(log.Fields{"Error": err}).Error("Failed to assign process to executor")
			log.Error("Retrying in 5 seconds ...")
			time.Sleep(5 * time.Second)
			continue
		}

		log.WithFields(log.Fields{
			"ProcessID":    process.ID,
			"ExecutorID":   e.executorID,
			"ExecutorName": e.executorName}).
			Info("Assigned process to executor")

		if process.FunctionSpec.FuncName == "execute" {
			if err != nil {
				log.WithFields(log.Fields{"Error": err}).Error("Failed to assign process to executor")
				return err
			}
			go func() {
				ok := e.executeDocker(process)
				if ok {
					e.client.Close(process.ID, e.executorPrvKey)
				}
			}()
		} else if process.FunctionSpec.FuncName == "sync" {
			err = e.syncHandler.PreSync(process, e.debugHandler, e.failureHandler)
			if err != nil {
				e.failureHandler.HandleError(process, err, "Failed to pre-sync")
				continue
			}
			err = e.client.Close(process.ID, e.executorPrvKey)
			if err != nil {
				e.failureHandler.HandleError(process, err, "Failed to close process, processID="+process.ID)
				continue
			}
		} else {
			log.WithFields(log.Fields{"FuncName": process.FunctionSpec.FuncName}).Error("Unsupported funcname")
			err := e.client.Fail(process.ID, []string{"Unsupported funcname"}, e.executorPrvKey)
			if err != nil {
				log.WithFields(log.Fields{"ProcessId": process.ID, "Error": err}).Error("Failed to close process as failed")
			}
			continue
		}
	}
}
