package executor

import (
	"context"
	"errors"
	"fmt"
	"math"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/colonyos/colonies/pkg/client"
	"github.com/colonyos/colonies/pkg/core"
	colparser "github.com/colonyos/colonies/pkg/parsers"
	"github.com/colonyos/colonies/pkg/security/crypto"
	"github.com/colonyos/executors/common/pkg/debug"
	"github.com/colonyos/executors/common/pkg/failure"
	"github.com/colonyos/executors/common/pkg/k8s"
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
	k8sName            string
	k8sNamespace       string
	ctx                context.Context
	cancel             context.CancelFunc
	client             *client.ColoniesClient
	syncHandler        *sync.SyncHandler
	failureHandler     *failure.FailureHandler
	debugHandler       *debug.DebugHandler
	k8sHandler         *k8s.K8sHandler
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

func WithK8sName(k8sName string) ExecutorOption {
	return func(e *Executor) {
		e.k8sName = k8sName
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

	e.k8sHandler, err = k8s.CreateK8sHandler(e.k8sName, e.namespace, e.pvc)
	if err != nil {
		return nil, err
	}

	log.WithFields(log.Fields{
		"Verbose":               e.verbose,
		"ColoniesServerHost":    e.coloniesServerHost,
		"ColoniesServerPort":    e.coloniesServerPort,
		"ColoniesInsecure":      e.coloniesInsecure,
		"K8sName":               e.k8sName,
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
		Info("Kube Executor started")

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

func (e *Executor) FetchJobLogs(process *core.Process, podNames []string, containers int) error {
	aggregatedLogsChan := make(chan string, 1000)
	eofChan := make(chan bool, 1000)
	errChan := make(chan error, 1000)
	e.k8sHandler.HandleJobLog(podNames, aggregatedLogsChan, eofChan, errChan)

	eofCounter := 0
	for {
		select {
		case msg := <-aggregatedLogsChan:
			fmt.Println("Adding log:", msg)
			err := e.client.AddLog(process.ID, msg, e.executorPrvKey)
			if err != nil {
				fmt.Println("error adding log:", err)
				return err
			}
		case err := <-errChan:
			fmt.Println("errchan error adding log:", err)
			return err
		case <-eofChan:
			eofCounter++
			if eofCounter == len(podNames)*containers {
				for i := 0; i < len(aggregatedLogsChan); i++ {
					msg := <-aggregatedLogsChan
					err := e.client.AddLog(process.ID, msg, e.executorPrvKey)
					if err != nil {
						fmt.Println("error adding log:", err)
						return err
					}
				}
				return nil
			}
		}
	}
}

func (e *Executor) executeK8s(process *core.Process) bool {
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

	spec := &k8s.JobSpec{
		JobName:           k8s.CreateUniqueJobName("kubexexecutor"),
		JobContainerImage: kwArgs.Image,
		ExecCmd:           cmd,
		ArgsStr:           kwArgs.Args,
		MountPath:         process.FunctionSpec.Filesystem.Mount,
		Parallelism:       process.FunctionSpec.Conditions.Nodes,
		ContainersPerPod:  process.FunctionSpec.Conditions.ProcessesPerNode,
		CPU:               process.FunctionSpec.Conditions.CPU,
		Memory:            process.FunctionSpec.Conditions.Memory,
		UseGPU:            process.FunctionSpec.Conditions.GPU.Count > 0,
		GPUCount:          process.FunctionSpec.Conditions.GPU.Count,
		GPUName:           process.FunctionSpec.Conditions.GPU.Name,
		ProcessID:         process.ID,
		Walltime:          process.FunctionSpec.Conditions.WallTime,
		EnvMap:            process.FunctionSpec.Env,
	}

	yaml, err := e.k8sHandler.ComposeJobYAML(spec)
	if err != nil {
		e.failureHandler.HandleError(process, err, "Failed to convert spec to k8s yaml")
		return false
	}

	fmt.Println(yaml)

	log.WithFields(log.Fields{"JobName": spec.JobName}).Info("Creating K8s batchjob")
	jobPodNames, err := e.k8sHandler.CreateJob(yaml, spec)
	if err != nil {
		e.failureHandler.HandleError(process, err, "Failed to create k8s batchjob")
		return false
	}

	log.WithFields(log.Fields{"JobName": spec.JobName, "Pods": jobPodNames}).Info("K8s batchjob created")
	log.WithFields(log.Fields{"JobName": spec.JobName, "Pods": jobPodNames}).Info("Monitoring K8s batchjob lifecycle, and getting logs")

	err = e.FetchJobLogs(process, jobPodNames, spec.ContainersPerPod)
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
		usedCPU, usedMem, err := e.k8sHandler.GetUtilization()
		if err != nil {
			log.Error(err)
			log.Error("Retrying in 5 seconds ...")
			time.Sleep(5 * time.Second)
			continue
		}

		totalCPUStr := e.hwCPU
		totalMemStr := e.hwMem

		totalCPUInt, err := colparser.ConvertCPUToInt(totalCPUStr)
		if err != nil {
			log.Error(err)
			log.Error("Retrying in 5 seconds ...")
			time.Sleep(5 * time.Second)
			continue
		}

		totalMem, err := colparser.ConvertMemoryToBytes(totalMemStr)
		if err != nil {
			log.Error(err)
			log.Error("Retrying in 5 seconds ...")
			time.Sleep(5 * time.Second)
			continue
		}

		availavbleCPU := totalCPUInt - usedCPU
		availavbleMem := totalMem - usedMem

		availableMemInKiB := float64(availavbleMem) / math.Pow(1024, 1)
		availableCPUStr := fmt.Sprintf("%dm", availavbleCPU)
		availableMemStr := fmt.Sprintf("%dKi", int(availableMemInKiB))

		log.WithFields(log.Fields{
			"UsedCPU":        usedCPU,
			"UsedMem":        usedMem,
			"TotalCPU":       totalCPUInt,
			"TotalMem":       totalMem,
			"AvailableCPU":   availavbleCPU,
			"AvailbleMem":    availavbleMem,
			"AvailbleMemStr": availableMemStr,
		}).Info("Resource utilization")

		process, err := e.client.AssignWithContext(e.colonyName, 100, e.ctx, availableCPUStr, availableMemStr, e.executorPrvKey)
		if err != nil {
			fmt.Println(err)
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

		log.WithFields(log.Fields{
			"ProcessID":    process.ID,
			"ExecutorID":   e.executorID,
			"ExecutorName": e.executorName}).
			Info("Assigned process to executor")

		if process.FunctionSpec.FuncName == "execute" {
			if err != nil {
				return err
			}
			go func() {
				ok := e.executeK8s(process)
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
