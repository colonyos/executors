package executor

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/colonyos/colonies/pkg/client"
	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/fs"
	"github.com/colonyos/colonies/pkg/security/crypto"
	log "github.com/sirupsen/logrus"
)

func strArr2Str(args []string) string {
	if len(args) == 0 {
		return ""
	}

	str := ""
	for _, arg := range args {
		str += arg + " "
	}

	return str[0 : len(str)-1]
}

func ifArr2StringArr(ifarr []interface{}) []string {
	strarr := make([]string, len(ifarr))
	for k, v := range ifarr {
		strarr[k] = fmt.Sprint(v)
	}

	return strarr
}

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
	slurm              *Slurm
	gres               bool
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

	e.slurm = CreateSlurm(e.fsDir, e.logDir, e.slurmPartition, e.slurmAccount, e.slurmModule, e.gres)

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

func (e *Executor) logInfo(process *core.Process, infoMsg string) {
	log.WithFields(log.Fields{"ProcessID": process.ID}).Info(infoMsg)
	if process != nil {
		err1 := e.client.AddLog(process.ID, "ColonyOS: "+infoMsg+"\n", e.executorPrvKey)
		if err1 != nil {
			log.WithFields(log.Fields{"ProcessId": process.ID, "Error": err1}).Error("Failed to add info log to process")
		}
	}
}

func (e *Executor) logError(process *core.Process, err error, errMsg string) {
	if err != nil {
		msg := "ColonyOS: "
		if errMsg != "" {
			msg += err.Error() + ":" + errMsg
		} else {
			msg += err.Error()
		}
		log.WithFields(log.Fields{"ProcessID": process.ID, "Error": err, "ErrMsg": errMsg}).Warn("Closing process as failed")
		if process != nil {
			err1 := e.client.AddLog(process.ID, msg+"\n", e.executorPrvKey)
			if err1 != nil {
				log.WithFields(log.Fields{"ProcessId": process.ID, "Error": err1}).Error("Failed to add error log to process")
			}
		}
	}
}

func (e *Executor) handleError(process *core.Process, err error, errMsg string) {
	if err != nil {
		msg := "ColonyOS: "
		if errMsg != "" {
			msg += err.Error() + ":" + errMsg
		} else {
			msg += err.Error()
		}
		if process != nil {
			log.WithFields(log.Fields{"ProcessID": process.ID, "Error": err, "ErrMsg": errMsg}).Warn("Closing process as failed")
			err1 := e.client.AddLog(process.ID, msg, e.executorPrvKey)
			if err1 != nil {
				log.WithFields(log.Fields{"ProcessId": process.ID, "Error": err1}).Error("Failed to add error log to process")
			}
			err2 := e.client.Fail(process.ID, []string{msg}, e.executorPrvKey)
			if err2 != nil {
				log.WithFields(log.Fields{"ProcessId": process.ID, "Error": err2}).Error("Failed to close process as failed")
			}
		}
	}
}

func (e *Executor) downloadSnapshots(process *core.Process) error {
	filesystem := process.FunctionSpec.Filesystem
	fsClient, err := fs.CreateFSClient(e.client, e.colonyID, e.executorPrvKey)
	if err != nil {
		e.handleError(process, err, "Failed to create FSClient, trying to download snapshots")
		return err
	}

	for _, snapshotMount := range filesystem.SnapshotMounts {
		if snapshotMount.SnapshotID != "" {
			snapshot, err := e.client.GetSnapshotByID(e.colonyID, snapshotMount.SnapshotID, e.executorPrvKey)
			if err != nil {
				e.handleError(process, err, "Failed to resolve snapshotID")
				return err
			}

			newDir := e.fsDir + snapshotMount.Dir
			newDir = strings.Replace(newDir, "{processid}", process.ID, 1)
			err = os.MkdirAll(newDir, 0755)
			if err != nil {
				e.handleError(process, err, "Failed to create download dir")
				return err
			}

			e.logInfo(process, "Creating directory: "+newDir)
			e.logInfo(process, "Downloading snapshot: Label:"+snapshot.Label+" SnapshotID:"+snapshot.ID+" Dir:"+newDir)
			log.WithFields(log.Fields{"Label": snapshotMount.Label, "SnapshotId": snapshot.ID, "Dir": snapshotMount.Dir}).Info("Downloading snapshot")
			err = fsClient.DownloadSnapshot(snapshot.ID, newDir)
			if err != nil {
				e.handleError(process, err, "Failed to download snapshot")
				return err
			}
		} else {
			log.Info("Ignoring downloading snapshot, snapshot Id was not set")
		}
	}

	return nil
}

func (e *Executor) sync(process *core.Process, onProcessStart bool) error {
	filesystem := process.FunctionSpec.Filesystem
	fsClient, err := fs.CreateFSClient(e.client, e.colonyID, e.executorPrvKey)
	if err != nil {
		e.handleError(process, err, "Failed to create FSClient, trying to sync")
		return err
	}

	for _, syncDirMount := range filesystem.SyncDirMounts {
		d := e.fsDir + syncDirMount.Dir
		d = strings.Replace(d, "{processid}", process.ID, 1)
		l := strings.Replace(syncDirMount.Label, "{processid}", process.ID, 1)
		if l != "" && d != "" {
			err = os.MkdirAll(d, 0755)
			if err != nil {
				log.WithFields(log.Fields{"Error": err}).Error("Failed to create download dir")
			}

			keepLocal := false
			if onProcessStart {
				keepLocal = syncDirMount.ConflictResolution.OnStart.KeepLocal
			} else {
				keepLocal = syncDirMount.ConflictResolution.OnClose.KeepLocal
			}
			syncplan, err := fsClient.CalcSyncPlan(d, l, keepLocal)
			if err != nil {
				log.WithFields(log.Fields{"Error": err}).Error("Failed to sync")
				return err
			}

			strategy := "keeplocal"
			if !keepLocal {
				strategy = "keepremote"
			}

			e.logInfo(process, "Starting directory synchronization: Label:"+l+" Dir:"+d+" Download:"+strconv.Itoa(len(syncplan.LocalMissing))+" Upload:"+strconv.Itoa(len(syncplan.RemoteMissing))+" Conflicts:"+strconv.Itoa(len(syncplan.RemoteMissing))+" ConflictResolutionStrategy:"+strategy)
			err = fsClient.ApplySyncPlan(e.colonyID, syncplan)
			if err != nil {
				e.handleError(process, err, "Failed to apply syncplan, Label:"+l+" Dir:"+d)
				return err
			}
		} else {
			log.Warn("Cannot perform directory synchronization, Label and Dir were not set")
		}
	}

	return nil
}

func (e *Executor) execute(process *core.Process) error {
	cmd, ok := process.FunctionSpec.KwArgs["cmd"].(string)
	if !ok {
		err := errors.New("Failed to parse cmd kwargs")
		e.handleError(process, err, "")
		return err
	}

	argsIf := process.FunctionSpec.KwArgs["args"]
	argsIfArray, ok := argsIf.([]interface{})
	var argsStr string
	if ok {
		arrStrArray := make([]string, len(argsIfArray))
		for i, v := range argsIfArray {
			arrStrArray[i] = v.(string)
		}
		argsStr = strArr2Str(ifArr2StringArr(argsIfArray))
	} else {
		err := errors.New("Failed to parse args kwargs")
		e.handleError(process, err, "")
		return err
	}

	log.WithFields(log.Fields{"Cmd": cmd, "Args": argsStr}).Info("Executing")

	execCmd := make([]string, 0)
	for _, arg := range ifArr2StringArr(argsIfArray) {
		execCmd = append(execCmd, arg)
	}

	execCmd = append([]string{cmd}, execCmd...)
	execCmdStr := strings.Join(execCmd[:], " ")

	command := exec.Command("sh", "-c", execCmdStr)
	command.Env = os.Environ()
	for _, attribute := range process.Attributes {
		command.Env = append(command.Env, attribute.Key+"="+attribute.Value)
	}

	// TODO: setting env variables does not work, it has to be set in the slurm script.
	// For an example, see how COLONIES_PROCESS_ID is set.
	command.Env = append(command.Env, "COLONIES_PROCESS_ID="+process.ID)
	command.Env = append(command.Env, "COLONIES_COLONY_ID="+e.colonyID)
	command.Env = append(command.Env, "COLONIES_PROCESS_ID="+process.ID)
	command.Env = append(command.Env, "COLONIES_SERVER_HOST="+e.coloniesServerHost)
	command.Env = append(command.Env, "COLONIES_SERVER_PORT="+strconv.Itoa(e.coloniesServerPort))
	command.Env = append(command.Env, "COLONIES_EXECUTOR_ID="+e.executorID)
	command.Env = append(command.Env, "COLONIES_EXECUTOR_PRVKEY="+e.executorPrvKey)

	// Get output
	stdout, err := command.StdoutPipe()
	if err != nil {
		e.handleError(process, err, "Failed to open stdout")
		return err
	}

	command.Stderr = command.Stdout

	if err = command.Start(); err != nil {
		e.handleError(process, err, "Failed to start cmd:"+execCmdStr)
		return err
	}

	var output string
	for {
		tmp := make([]byte, 1)
		_, err := stdout.Read(tmp)
		if err != nil {
			break
		}
		if e.logDir != "" {
			f, err := os.OpenFile(e.logDir+"/"+process.ID+".log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				log.WithFields(log.Fields{"Error": err}).Info("Failed to open logfile")
				err = e.client.Fail(process.ID, []string{"Failed to open logfile"}, e.executorPrvKey)
				return err
			}
			defer f.Close()
			_, err = f.WriteString(string(tmp))
			if err != nil {
				log.WithFields(log.Fields{"Error": err}).Info("Failed to write to logfile")
				err = e.client.Fail(process.ID, []string{"Failed to write to logile"}, e.executorPrvKey)
				return err
			}
		}
		c := string(bytes.Trim(tmp, "\x00"))
		output += c
		if c == "\n" {
			err = e.client.AddLog(process.ID, output, e.executorPrvKey)
			if err != nil {
				e.handleError(process, err, "Failed to add log to process")
				return err
			}
			output = ""
		}
	}

	if err = command.Wait(); err != nil {
		e.handleError(process, err, "Failed to execute process")
		return err
	}

	return nil
}

func (e *Executor) monitorSlurmForever() {
	go func() {
		logChan := make(chan *Log, 1000)
		jobEndedChan := make(chan *JobEnded, 1000)
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
					e.handleError(nil, err, "Failed to get process, processID="+jobEnded.ProcessID)
					continue
				}
				if process == nil {
					e.handleError(nil, err, "Failed to get process, process is nil, processID="+jobEnded.ProcessID)
					continue
				}
				if jobEnded.JobStatus == COMPLETED || jobEnded.JobStatus == COMPLETING {
					for _, snapshotMount := range process.FunctionSpec.Filesystem.SnapshotMounts {
						if !snapshotMount.KeepSnaphot {
							if snapshotMount.SnapshotID != "" {
								err = e.client.DeleteSnapshotByID(e.colonyID, snapshotMount.SnapshotID, e.executorPrvKey)
								if err != nil {
									log.WithFields(log.Fields{"SnapshotID": snapshotMount.SnapshotID, "Error": err}).Error("Failed to delete snapshot")
								} else {
									log.WithFields(log.Fields{"SnapshotID": snapshotMount.SnapshotID}).Debug("Snapshot deleted")
								}
							}
							if !snapshotMount.KeepFiles {
								d := e.fsDir + snapshotMount.Dir
								d = strings.Replace(d, "{processid}", process.ID, 1)
								e.logInfo(process, "Deleting snapshot mount dir: "+d)
								err := os.RemoveAll(d)
								if err != nil {
									e.handleError(nil, err, "Failed to delete snapshot files")
								}
							}
						}
					}
					onProcessStart := false
					err = e.sync(process, onProcessStart)
					if err != nil {
						e.handleError(process, err, "Failed to sync to filesystem, onProcessStart="+strconv.FormatBool(onProcessStart))
						continue
					}
					for _, syncDirMount := range process.FunctionSpec.Filesystem.SyncDirMounts {
						if !syncDirMount.KeepFiles {
							d := e.fsDir + syncDirMount.Dir
							d = strings.Replace(d, "{processid}", process.ID, 1)
							e.logInfo(process, "Deleting syncdir mount: "+d)
							err := os.RemoveAll(d)
							if err != nil {
								e.handleError(nil, err, "Failed to delete syncdir files")
							}
						}
					}
					err = e.client.Close(jobEnded.ProcessID, e.executorPrvKey)
					if err != nil {
						e.handleError(process, err, "Failed to close process, processID="+jobEnded.ProcessID)
						continue
					}
				} else {
					// TODO, add better error message to client
					err := errors.New("Failed to execute slurm script")
					e.handleError(process, err, "")
					continue
				}
			}
		}
	}()
}

func (e *Executor) executeSlurm(process *core.Process) error {
	err := e.downloadSnapshots(process)
	if err != nil {
		e.handleError(process, err, "Failed to download snapshots")
		return err
	}

	onProcessStart := true
	err = e.sync(process, onProcessStart)
	if err != nil {
		e.handleError(process, err, "Failed to sync to filesystem, onProcessStart="+strconv.FormatBool(onProcessStart))
		return err
	}

	imageIf := process.FunctionSpec.KwArgs["docker-image"]
	image, ok := imageIf.(string)
	if !ok {
		err := errors.New("Failed to parse docker image flag")
		e.handleError(process, err, "")
		return err
	}

	rebuildImageIf := process.FunctionSpec.KwArgs["rebuild-image"]
	rebuildImage, ok := rebuildImageIf.(bool)
	if !ok {
		e.logInfo(process, "Failed to parse rebuild image flag, setting rebuildImage to false")
		rebuildImage = false
	}

	cmd, ok := process.FunctionSpec.KwArgs["cmd"].(string)
	if !ok {
		err := errors.New("Failed to parse cmd kwarg")
		e.handleError(process, err, "")
		return err
	}

	argsIf := process.FunctionSpec.KwArgs["args"]
	argsIfArray, ok := argsIf.([]interface{})
	var argsStr string
	if ok {
		arrStrArray := make([]string, len(argsIfArray))
		for i, v := range argsIfArray {
			arrStrArray[i] = v.(string)
		}
		argsStr = strArr2Str(ifArr2StringArr(argsIfArray))
	} else {
		e.logInfo(process, "Failed to parse args, setting args to empty string")
		argsStr = ""
	}

	execCmd := make([]string, 0)
	for _, arg := range ifArr2StringArr(argsIfArray) {
		arg = strings.Replace(arg, "{processid}", process.ID, 1)
		execCmd = append(execCmd, arg)
	}

	execCmd = append([]string{cmd}, execCmd...)
	execCmdStr := strings.Join(execCmd[:], " ")

	singularity := CreateSingularity(e.imageDir)
	script, err := e.slurm.GenerateSlurmScript(process.FunctionSpec.Conditions.Nodes,
		process.FunctionSpec.Conditions.ProcessesPerNode,
		int(process.FunctionSpec.Conditions.WallTime),
		process.FunctionSpec.Conditions.Memory,
		process.FunctionSpec.Conditions.GPU.Count,
		execCmdStr,
		singularity.Sif(image),
		process.ID,
		process.FunctionSpec.Filesystem.Mount)
	if err != nil {
		e.handleError(process, err, "Failed to generate Slurm script")
		return err
	}

	e.logInfo(process, "Generated Slurm script:\n "+script)

	e.logInfo(process, "Creating singularity container: "+image+" to "+e.imageDir)
	if !singularity.SifExists(image) {
		logs, err := singularity.Pull(image)
		e.logInfo(process, logs)
		if err != nil {
			e.handleError(process, err, "Failed to pull container image: "+image)
			return err
		}
	} else {
		e.logInfo(process, "Image already exists: "+image)
	}

	jobID, err := e.slurm.Submit(script)
	if err != nil {
		e.handleError(process, err, "Failed to submit Slurm script")
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
		"Cmd":              cmd,
		"ExecCmd":          execCmdStr,
		"Args":             argsStr,
		"DockerImage":      image,
		"RebuildImage":     rebuildImage,
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
				// TODO: should we sleep here?
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
