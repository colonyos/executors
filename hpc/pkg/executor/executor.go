package executor

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/colonyos/colonies/pkg/client"
	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/security/crypto"
	"github.com/colonyos/executors/common/pkg/debug"
	"github.com/colonyos/executors/common/pkg/failure"
	"github.com/colonyos/executors/common/pkg/parsers"
	"github.com/colonyos/executors/common/pkg/singularity"
	"github.com/colonyos/executors/common/pkg/slurm"
	"github.com/colonyos/executors/common/pkg/sync"
	log "github.com/sirupsen/logrus"
)

const DEFAULT_CONTAINER_MOUNT = "/cfs"

type Executor struct {
	verbose            bool
	coloniesServerHost string
	coloniesServerPort int
	coloniesInsecure   bool
	colonyID           string
	colonyPrvKey       string
	executorID         string
	executorPrvKey     string
	executorType       string
	logDir             string
	fsDir              string
	imageDir           string
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
	slurmAccount       string
	slurmPartition     string
	slurmModule        string
	ctx                context.Context
	cancel             context.CancelFunc
	client             *client.ColoniesClient
	slurm              *slurm.Slurm
	gres               bool
	syncHandler        *sync.SyncHandler
	failureHandler     *failure.FailureHandler
	debugHandler       *debug.DebugHandler
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

func WithExecutorType(executorType string) ExecutorOption {
	return func(e *Executor) {
		e.executorType = executorType
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

func WithLogDir(logDir string) ExecutorOption {
	return func(e *Executor) {
		e.logDir = logDir
	}
}

func WithFsDir(fsDir string) ExecutorOption {
	return func(e *Executor) {
		e.fsDir = fsDir
	}
}

func WithImageDir(imageDir string) ExecutorOption {
	return func(e *Executor) {
		e.imageDir = imageDir
	}
}

func WithSlurmAccount(slurmAccount string) ExecutorOption {
	return func(e *Executor) {
		e.slurmAccount = slurmAccount
	}
}

func WithSlurmPartition(slurmPartition string) ExecutorOption {
	return func(e *Executor) {
		e.slurmPartition = slurmPartition
	}
}

func WithSlurmModule(slurmModule string) ExecutorOption {
	return func(e *Executor) {
		e.slurmModule = slurmModule
	}
}

func WithGRES(gres bool) ExecutorOption {
	return func(e *Executor) {
		e.gres = gres
	}
}

func (e *Executor) createColoniesExecutorWithKey(colonyID string) (*core.Executor, string, string, error) {
	crypto := crypto.CreateCrypto()
	executorPrvKey, err := crypto.GeneratePrivateKey()
	if err != nil {
		return nil, "", "", err
	}

	executorID, err := crypto.GenerateID(executorPrvKey)
	if err != nil {
		return nil, "", "", err
	}

	executor := core.CreateExecutor(executorID, e.executorType, e.executorType+"-"+core.GenerateRandomID(), colonyID, time.Now(), time.Now())
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

	if e.verbose {
		log.SetLevel(log.DebugLevel)
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
		spec, executorID, executorPrvKey, err := e.createColoniesExecutorWithKey(e.colonyID)
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

		log.WithFields(log.Fields{"ExecutorID": e.executorID}).Info("Self-registered")
	}
	function := &core.Function{ExecutorID: e.executorID, ColonyID: e.colonyID, FuncName: "execute"}
	e.client.AddFunction(function, e.executorPrvKey)

	e.slurm = slurm.CreateSlurm(e.fsDir, e.logDir, e.slurmPartition, e.slurmAccount, e.slurmModule, e.gres)

	var err error
	e.failureHandler, err = failure.CreateFailureHandler(e.executorPrvKey, e.client)
	if err != nil {
		return nil, err
	}

	e.debugHandler, err = debug.CreateDebugHandler(e.executorPrvKey, e.client)
	if err != nil {
		return nil, err
	}

	e.syncHandler, err = sync.CreateSyncHandler(e.colonyID, e.executorPrvKey, e.client, e.fsDir, e.failureHandler, e.debugHandler)
	if err != nil {
		return nil, err
	}

	log.WithFields(log.Fields{
		"Verbose":               e.verbose,
		"ColoniesServerHost":    e.coloniesServerHost,
		"ColoniesServerPort":    e.coloniesServerPort,
		"ColoniesInsecure":      e.coloniesInsecure,
		"LogDir":                e.logDir,
		"FsDir":                 e.fsDir,
		"ImageDir":              e.imageDir,
		"ColonyId":              e.colonyID,
		"ColonyPrvKey":          "***********************",
		"ExecutorId":            e.executorID,
		"ExecutorPrvKey":        "***********************",
		"Longitude":             e.long,
		"Latitude":              e.lat,
		"LocationDesc":          e.locDesc,
		"SlurmAccount":          e.slurmAccount,
		"SlurmPartition":        e.slurmPartition,
		"SlurmModule":           e.slurmModule,
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
		"GRES":                  e.gres,
		"SoftwareVersion":       e.swVersion,
		"SoftwareType":          e.swType}).
		Info("HPC Executor started")

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

func (e *Executor) monitorSlurmForever() {
	go func() {
		logChan := make(chan *slurm.Log, 1000)
		jobEndedChan := make(chan *slurm.JobEnded, 1000)
		log.WithFields(log.Fields{"LogDir": e.logDir}).Info("Starting Slurm monitor")
		e.slurm.Monitor(e.logDir, logChan, jobEndedChan)
		for {
			select {
			case log := <-logChan:
				e.client.AddLog(log.ProcessID, log.Log, e.executorPrvKey)
			case jobEnded := <-jobEndedChan:
				log.WithFields(log.Fields{"ProcessId": jobEnded.ProcessID, "SlurmJobId": jobEnded.JobID, "JobStatus": jobEnded.JobStatus}).Info("Slurm job completed")
				process, err := e.client.GetProcess(jobEnded.ProcessID, e.executorPrvKey)
				if err != nil {
					e.failureHandler.HandleError(nil, err, "Failed to get process, processID="+jobEnded.ProcessID)
					continue
				}
				if process == nil {
					e.failureHandler.HandleError(nil, err, "Failed to get process, process is nil, processID="+jobEnded.ProcessID)
					continue
				}
				if jobEnded.JobStatus == slurm.COMPLETED || jobEnded.JobStatus == slurm.COMPLETING {
					err = e.syncHandler.PostSync(process, e.debugHandler, e.failureHandler, e.fsDir, e.client, e.colonyID, e.executorPrvKey)
					if err != nil {
						continue
					}
					err = e.client.Close(jobEnded.ProcessID, e.executorPrvKey)
					if err != nil {
						e.failureHandler.HandleError(process, err, "Failed to close process, processID="+jobEnded.ProcessID)
						continue
					}
				} else {
					// TODO, add better error message to client
					err := errors.New("Failed to execute slurm script")
					e.failureHandler.HandleError(process, err, "")
					continue
				}
			}
		}
	}()
}

func (e *Executor) executeSlurm(process *core.Process) error {
	kwArgs, err := parsers.ParseKwArgs(process, e.failureHandler, e.debugHandler)
	if err != nil {
		return err
	}

	err = e.syncHandler.PreSync(process, e.debugHandler, e.failureHandler)
	if err != nil {
		return err
	}

	containerMount := process.FunctionSpec.Filesystem.Mount
	if containerMount == "" {
		containerMount = DEFAULT_CONTAINER_MOUNT
	}

	singularity := singularity.CreateSingularity(e.imageDir)
	script, err := e.slurm.GenerateSlurmScript(process.FunctionSpec.Conditions.Nodes,
		process.FunctionSpec.Conditions.ProcessesPerNode,
		int(process.FunctionSpec.Conditions.WallTime),
		process.FunctionSpec.Conditions.Memory,
		process.FunctionSpec.Conditions.GPU.Count,
		kwArgs.ExecCmd,
		singularity.Sif(kwArgs.Image),
		process.ID,
		process,
		containerMount,
		fmt.Sprintf("%t", !e.coloniesInsecure),
		e.coloniesServerHost,
		strconv.Itoa(e.coloniesServerPort),
		e.colonyID,
		e.executorID,
		e.executorPrvKey)
	if err != nil {
		e.failureHandler.HandleError(process, err, "Failed to generate Slurm script")
		return err
	}

	fmt.Println(script)

	if kwArgs.RebuildImage {
		err := singularity.RemoveSif(kwArgs.Image)
		e.failureHandler.HandleError(process, err, "Failed to remove Singularity image: "+kwArgs.Image)
	}

	e.debugHandler.LogInfo(process, "Creating singularity container: "+kwArgs.Image+" to "+e.imageDir)
	if !singularity.SifExists(kwArgs.Image) {
		logs, err := singularity.Pull(kwArgs.Image)
		e.debugHandler.LogInfo(process, logs)
		if err != nil {
			e.failureHandler.HandleError(process, err, "Failed to pull container image: "+kwArgs.Image)
			return err
		}
	} else {
		e.debugHandler.LogInfo(process, "Image already exists: "+kwArgs.Image)
	}

	jobID, err := e.slurm.Submit(script)
	if err != nil {
		e.failureHandler.HandleError(process, err, "Failed to submit Slurm script")
		return err
	}

	log.WithFields(log.Fields{
		"ProcessID":        process.ID,
		"SlurmJobID":       jobID,
		"Nodes":            process.FunctionSpec.Conditions.Nodes,
		"Memory":           process.FunctionSpec.Conditions.Memory,
		"CPU":              process.FunctionSpec.Conditions.CPU,
		"Processes":        process.FunctionSpec.Conditions.Processes,
		"ProcessesPerNode": process.FunctionSpec.Conditions.ProcessesPerNode,
		"Walltime":         process.FunctionSpec.Conditions.WallTime,
		"GPUName":          process.FunctionSpec.Conditions.GPU.Name,
		"GPUMemory":        process.FunctionSpec.Conditions.GPU.Memory,
		"GPUCount":         process.FunctionSpec.Conditions.GPU.Count,
		"Cmd":              kwArgs.Cmd,
		"ExecCmd":          kwArgs.ExecCmd,
		"Args":             kwArgs.Args,
		"DockerImage":      kwArgs.Image,
		"RebuildImage":     kwArgs.RebuildImage,
		"SlurmBatchScript": script,
		"ExecutorType":     process.FunctionSpec.Conditions.ExecutorType}).
		Info("Executing process")
	return nil
}

func (e *Executor) ServeForEver() error {
	e.monitorSlurmForever()

	for {
		process, err := e.client.AssignWithContext(e.colonyID, 100, e.ctx, e.executorPrvKey)
		if err != nil {
			var coloniesError *core.ColoniesError
			if errors.As(err, &coloniesError) {
				if coloniesError.Status == 404 { // No processes can be selected for executor
					log.Warn(err)
					continue
				}
			}

			log.Error(err)
			log.Error("Retrying in 5 seconds ...")
			time.Sleep(5 * time.Second)
			continue
		}

		log.WithFields(log.Fields{"ProcessID": process.ID, "ExecutorID": e.executorID}).Info("Assigned process to executor")

		if process.FunctionSpec.FuncName == "execute" {
			err := e.executeSlurm(process)
			if err != nil {
				log.WithFields(log.Fields{"ProcessID": process.ID, "ExecutorID": e.executorID}).Error("Failed to execute process")
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
