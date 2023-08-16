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
	coloniesServerHost string
	coloniesServerPort int
	coloniesInsecure   bool
	colonyID           string
	colonyPrvKey       string
	executorID         string
	executorPrvKey     string
	logdir             string
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

func WithLogdir(logdir string) ExecutorOption {
	return func(e *Executor) {
		e.logdir = logdir
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

	executorType := os.Getenv("EXECUTOR_TYPE")
	if executorType == "" {
		log.Debug("Executor type not specifed, defaulting to ml")
		executorType = "ml"
	}

	executor := core.CreateExecutor(executorID, executorType, executorType+"-"+core.GenerateRandomID(), colonyID, time.Now(), time.Now())
	executor.Capabilities.Software.Name = os.Getenv("EXECUTOR_SW_NAME")
	executor.Capabilities.Software.Type = os.Getenv("EXECUTOR_SW_TYPE")
	executor.Capabilities.Software.Version = os.Getenv("EXECUTOR_SW_VERSION")
	executor.Capabilities.Hardware.CPU = os.Getenv("EXECUTOR_HW_CPU")
	executor.Capabilities.Hardware.Model = os.Getenv("EXECUTOR_HW_MODEL")
	executor.Capabilities.Hardware.Memory = os.Getenv("EXECUTOR_HW_MEM")
	executor.Capabilities.Hardware.Storage = os.Getenv("EXECUTOR_HW_STORAGE")
	gpuCountStr := os.Getenv("EXECUTOR_HW_GPU_COUNT")
	gpuCount, err := strconv.Atoi(gpuCountStr)
	if err != nil {
		log.Error("Failed to set gpu count")
	}
	executor.Capabilities.Hardware.GPU.Count = gpuCount
	executor.Capabilities.Hardware.GPU.Name = os.Getenv("EXECUTOR_HW_GPU_NAME")
	executor.Location.Description = os.Getenv("EXECUTOR_LOCATION_DESC")
	longStr := os.Getenv("EXECUTOR_LOCATION_LONG")
	long, err := strconv.ParseFloat(longStr, 64)
	if err != nil {
		log.Error("Failed to set location long")
	}
	executor.Location.Long = long
	latStr := os.Getenv("EXECUTOR_LOCATION_LONG")
	lat, err := strconv.ParseFloat(latStr, 64)
	if err != nil {
		log.Error("Failed to set location long")
	}
	executor.Location.Lat = lat

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

		log.WithFields(log.Fields{"ExecutorID": e.executorID}).Info("Self-registered")
	}
	function := &core.Function{ExecutorID: e.executorID, ColonyID: e.colonyID, FuncName: "execute"}
	e.client.AddFunction(function, e.executorPrvKey)

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
		snapshot, err := e.client.GetSnapshotByID(e.colonyID, syncDir.SnapshotID, e.executorPrvKey)
		if err != nil {
			log.WithFields(log.Fields{"Error": err}).Error("Failed to resolve snapshot Id")
			return err
		}
		log.WithFields(log.Fields{"Label": syncDir.Label, "SnapshotID": snapshot.ID, "Dir": syncDir.Dir}).Info("Downloading snapshot")

		err = os.MkdirAll(syncDir.Dir, 0755)
		if err != nil {
			log.WithFields(log.Fields{"Error": err}).Error("Failed to create download dir")
		}

		err = fsClient.DownloadSnapshot(snapshot.ID, syncDir.Dir)
		if err != nil {
			log.WithFields(log.Fields{"Error": err}).Error("Failed to download snapshot")
			return err
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
		if e.logdir != "" {
			f, err := os.OpenFile(e.logdir+"/"+process.ID+".log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
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
			err = e.downloadSnapshots(process.FunctionSpec.Filesystem)
			if err != nil {
				errMsg := "Failed to parse snapshots kwarg"
				e.failProcess(process, errMsg)
				continue
			}

			err = e.execute(process)
			if err == nil {
				log.WithFields(log.Fields{"ProcessID": process.ID}).Info("Done executing, closing process as successful")
				err = e.client.Close(process.ID, e.executorPrvKey)
				if err != nil {
					log.WithFields(log.Fields{"ProcessId": process.ID, "Error": err}).Error("Failed to close process as successful")
				}
				keepSnapshotsIf := process.FunctionSpec.KwArgs["keep_snapshots"]
				keepSnapshots, ok := keepSnapshotsIf.(bool)
				if !ok {
					log.WithFields(log.Fields{"ProcessID": process.ID}).Info("Failed to parse keep snapshot flag")
				}
				if !keepSnapshots {
					for _, syncDir := range process.FunctionSpec.Filesystem {
						// fsClient, err := fs.CreateFSClient(e.client, e.colonyID, e.executorPrvKey)
						// if err != nil {
						// 	log.withfields(log.fields{"error": err, "snapshotid": syncdir.snapshotid}).error("failed to create fsclient, cannot delete snapshot")
						// }
						err = e.client.DeleteSnapshotByID(e.colonyID, syncDir.SnapshotID, e.executorPrvKey)
						if err != nil {
							log.WithFields(log.Fields{"SnapshotID": syncDir.SnapshotID, "Error": err}).Error("Failed to delete snapshot")
						} else {
							log.WithFields(log.Fields{"SnapshotID": syncDir.SnapshotID}).Debug("Snapshot deleted")
						}
					}
				}
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
