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

	log.WithFields(log.Fields{
		"Verbose":               e.verbose,
		"ColoniesServerHost":    e.coloniesServerHost,
		"ColoniesServerPort":    e.coloniesServerPort,
		"ColoniesInsecure":      e.coloniesInsecure,
		"LogDir":                e.logDir,
		"fsDir":                 e.fsDir,
		"imageDir":              e.imageDir,
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
		"SoftwareVersion":       e.swVersion,
		"SoftwareType":          e.swType}).
		Info("Starting a Colonies Unix executor")

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

func (e *Executor) failProcess(process *core.Process, errMsg string) {
	log.WithFields(log.Fields{"ProcessID": process.ID}).Error(errMsg)
	err := e.client.Fail(process.ID, []string{errMsg}, e.executorPrvKey)
	if err != nil {
		log.WithFields(log.Fields{"ProcessId": process.ID, "Error": err}).Error("Failed to close process as failed")
	}
}

func (e *Executor) downloadSnapshots(filesystem []*core.SyncDir) error {
	fsClient, err := fs.CreateFSClient(e.client, e.colonyID, e.executorPrvKey)
	if err != nil {
		log.WithFields(log.Fields{"Error": err}).Error("Failed to create FSClient")
		log.Error(err)
		return err
	}

	for _, syncDir := range filesystem {
		if syncDir.SnapshotID != "" {
			snapshot, err := e.client.GetSnapshotByID(e.colonyID, syncDir.SnapshotID, e.executorPrvKey)
			if err != nil {
				log.WithFields(log.Fields{"Error": err, "SnapshotId": syncDir.SnapshotID}).Error("Failed to resolve snapshot Id")
				return err
			}
			log.WithFields(log.Fields{"Label": syncDir.Label, "SnapshotId": snapshot.ID, "Dir": syncDir.Dir}).Info("Downloading snapshot")

			err = os.MkdirAll(e.fsDir+"/"+syncDir.Dir, 0755)
			if err != nil {
				log.WithFields(log.Fields{"Error": err}).Error("Failed to create download dir")
			}

			err = fsClient.DownloadSnapshot(snapshot.ID, e.fsDir+"/"+syncDir.Dir)
			if err != nil {
				log.WithFields(log.Fields{"Error": err}).Error("Failed to download snapshot")
				return err
			}
		} else {
			log.Info("Ignoring downloading snapshot, snapshot Id was not set")
		}
	}

	return nil
}

func (e *Executor) sync(filesystem []*core.SyncDir) error {
	fsClient, err := fs.CreateFSClient(e.client, e.colonyID, e.executorPrvKey)
	if err != nil {
		log.WithFields(log.Fields{"Error": err}).Error("Failed to create FSClient")
		log.Error(err)
		return err
	}

	for _, syncDir := range filesystem {
		if syncDir.SyncOnCompletion && syncDir.Label != "" && syncDir.Dir != "" {
			err = os.MkdirAll(syncDir.Dir, 0755)
			if err != nil {
				log.WithFields(log.Fields{"Error": err}).Error("Failed to create download dir")
			}

			syncplan, err := fsClient.CalcSyncPlan(syncDir.Dir, syncDir.Label, true)
			if err != nil {
				log.WithFields(log.Fields{"Error": err}).Error("Failed to sync")
				return err
			}

			err = fsClient.ApplySyncPlan(e.colonyID, syncplan)
			if err != nil {
				log.WithFields(log.Fields{"Error": err, "Label": syncDir.Label, "Dir": syncDir.Dir}).Error("Failed to apply sync plan")
				return err
			}
		} else {
			log.Info("Ignoring sync, SyncOnCompletion, Label and Dir were not set")
		}
	}

	return nil
}

func (e *Executor) execute(process *core.Process) error {
	cmd, ok := process.FunctionSpec.KwArgs["cmd"].(string)
	if !ok {
		errMsg := "Failed to parse cmd kwarg"
		e.failProcess(process, errMsg)
		return nil
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

	command.Env = append(command.Env, "COLONIES_COLONY_ID="+e.colonyID)
	command.Env = append(command.Env, "COLONIES_PROCESS_ID="+process.ID)
	command.Env = append(command.Env, "COLONIES_SERVER_HOST="+e.coloniesServerHost)
	command.Env = append(command.Env, "COLONIES_SERVER_PORT="+strconv.Itoa(e.coloniesServerPort))
	command.Env = append(command.Env, "COLONIES_EXECUTOR_ID="+e.executorID)
	command.Env = append(command.Env, "COLONIES_EXECUTOR_PRVKEY="+e.executorPrvKey)

	// Get output
	stdout, err := command.StdoutPipe()
	if err != nil {
		log.WithFields(log.Fields{"Error": err}).Info("Failed to open stdout")
		err = e.client.Fail(process.ID, []string{"Failed to open stdout"}, e.executorPrvKey)
		return err
	}

	command.Stderr = command.Stdout

	if err = command.Start(); err != nil {
		log.WithFields(log.Fields{"Error": err}).Info("Failed to start cmd")
		err = e.client.Fail(process.ID, []string{"Failed to start cmd"}, e.executorPrvKey)
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
				e.failProcess(process, err.Error())
				return err
			}
			output = ""
		}
	}

	if err = command.Wait(); err != nil {
		log.WithFields(log.Fields{"ProcessID": process.ID}).Info("Closing process as failed")
		e.client.Fail(process.ID, []string{"failed to execute process"}, e.executorPrvKey)
		return err
	}

	return nil
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

		if process.FunctionSpec.FuncName == "execute" {
			fmt.Println(process.FunctionSpec.Conditions)
			err = e.downloadSnapshots(process.FunctionSpec.Filesystem)
			if err != nil {
				fmt.Println("err", err)
				log.WithFields(log.Fields{"ProcessID": process.ID, "Error": err}).Info("Failed to download snapshot")
				e.failProcess(process, err.Error())
				continue
			}
			e.client.Close(process.ID, e.executorPrvKey)
			singularity := CreateSingularity(e.imageDir)
			fmt.Println(singularity)

			imageIf := process.FunctionSpec.KwArgs["docker-image"]
			image, ok := imageIf.(string)
			if !ok {
				log.WithFields(log.Fields{"ProcessID": process.ID}).Info("Failed to parse docker image flag")
				continue
			}
			if !singularity.SifExists(image) {
				singularity.Build(image)
			} else {
				log.WithFields(log.Fields{"Image": image}).Info("Image already exists")
			}

			//slurm := CreateSlurm(e.fsDir, e.imageDir, e.logDir, e.slurmPartition, e.slurmAccount, e.slurmModule)
			//slurm.Submit()

			// err = e.execute(process)
			// if err == nil {
			// 	keepSnapshotsIf := process.FunctionSpec.KwArgs["keep_snapshots"]
			// 	keepSnapshots, ok := keepSnapshotsIf.(bool)
			// 	if !ok {
			// 		log.WithFields(log.Fields{"ProcessID": process.ID}).Info("Failed to parse keep snapshot flag")
			// 	}
			// 	if !keepSnapshots {
			// 		for _, syncDir := range process.FunctionSpec.Filesystem {
			// 			if syncDir.SnapshotID != "" {
			// 				err = e.client.DeleteSnapshotByID(e.colonyID, syncDir.SnapshotID, e.executorPrvKey)
			// 				if err != nil {
			// 					log.WithFields(log.Fields{"SnapshotID": syncDir.SnapshotID, "Error": err}).Error("Failed to delete snapshot")
			// 				} else {
			// 					log.WithFields(log.Fields{"SnapshotID": syncDir.SnapshotID}).Debug("Snapshot deleted")
			// 				}
			// 			}
			// 		}
			// 	}

			// 	err = e.sync(process.FunctionSpec.Filesystem)
			// 	if err != nil {
			// 		log.WithFields(log.Fields{"ProcessID": process.ID, "Error": err}).Info("Failed to sync")
			// 		e.failProcess(process, err.Error())
			// 		continue
			// 	}

			// 	log.WithFields(log.Fields{"ProcessID": process.ID}).Info("Done executing, closing process as successful")
			// 	err = e.client.Close(process.ID, e.executorPrvKey)
			// 	if err != nil {
			// 		log.WithFields(log.Fields{"ProcessId": process.ID, "Error": err}).Error("Failed to close process as successful")
			// 	}
			// }
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
