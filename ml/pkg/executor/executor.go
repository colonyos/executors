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

	executor := core.CreateExecutor(executorID, "ml", "ml-"+core.GenerateRandomID(), colonyID, time.Now(), time.Now())
	executor.Capabilities.Software.Name = "Colonies ML Kubernetes Executor"
	executor.Capabilities.Software.Type = "ml"
	executor.Capabilities.Software.Version = "1.0.0-beta1"
	executor.Capabilities.Hardware.CPU = "32000m"
	executor.Capabilities.Hardware.GPU.Count = 1
	executor.Capabilities.Hardware.GPU.Name = "NVIDIA GeForce RTX 3080 Ti"
	executor.Capabilities.Hardware.Model = "AMD Ryzen 9 5950X (32) @ 3.400GHz"
	executor.Capabilities.Hardware.Model = "80337Mi"
	executor.Capabilities.Hardware.Storage = "10Gi"
	executor.Location.Description = "Rutvik"

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

type snapshotArg struct {
	name string
	dir  string
}

func (e *Executor) parseSnapshotArg(process *core.Process) ([]snapshotArg, error) {
	var s []snapshotArg

	snapshotArgsIfArr, ok := process.FunctionSpec.KwArgs["snapshots"] //.([]map[string]string)
	if !ok {
		errMsg := "Failed to parse snapshots kwarg"
		e.failProcess(process, errMsg)
		return s, errors.New(errMsg)
	}

	var snapshotArgsMap []map[string]string
	for _, item := range snapshotArgsIfArr.([]interface{}) {
		if m, ok := item.(map[string]interface{}); ok {
			newMap := make(map[string]string)
			for k, v := range m {
				if strVal, ok := v.(string); ok {
					newMap[k] = strVal
				} else {
					errMsg := "Failed to parse snapshots kwarg"
					e.failProcess(process, errMsg)
					return s, errors.New(errMsg)
				}
			}
			snapshotArgsMap = append(snapshotArgsMap, newMap)
		}
	}

	for _, m := range snapshotArgsMap {
		dir, ok := m["dir"]
		if !ok {
			errMsg := "Failed to parse snapshot dir key"
			e.failProcess(process, errMsg)
			return s, errors.New(errMsg)
		}
		snapshotName, ok := m["name"]
		if !ok {
			errMsg := "Failed to parse snapshot name key"
			e.failProcess(process, errMsg)
			return s, errors.New(errMsg)
		}

		s = append(s, snapshotArg{name: snapshotName, dir: dir})
	}

	return s, nil
}

func (e *Executor) downloadSnapshots(snapshotsArg []snapshotArg) error {
	fsClient, err := fs.CreateFSClient(e.client, e.colonyID, e.executorPrvKey)
	if err != nil {
		log.WithFields(log.Fields{"Error": err}).Error("Failed to create FSClient")
		log.Error(err)
		return err
	}

	for _, s := range snapshotsArg {
		snapshot, err := e.client.GetSnapshotByName(e.colonyID, s.name, e.executorPrvKey)
		if err != nil {
			log.WithFields(log.Fields{"Error": err}).Error("Failed to resolve snapshot Id")
			return err
		}
		log.WithFields(log.Fields{"SnapshotName": s.name, "SnapshotID": snapshot.ID, "Dir": s.dir}).Info("Downloading snapshot")

		err = os.Mkdir(s.dir, 0755)
		if err == nil {
			log.WithFields(log.Fields{"Error": err}).Error("Failed to create download dir")
			return err
		}

		err = fsClient.DownloadSnapshot(snapshot.ID, s.dir)
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
	argsIfArray := argsIf.([]interface{})
	arrStrArray := make([]string, len(argsIfArray))

	for i, v := range argsIfArray {
		arrStrArray[i] = v.(string)
	}
	argsStr := strArr2Str(ifArr2StringArr(argsIfArray))

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
			snapshotsArg, err := e.parseSnapshotArg(process)
			if err != nil {
				continue
			}

			err = e.downloadSnapshots(snapshotsArg)
			if err != nil {
				continue
			}

			err = e.execute(process)
			if err == nil {
				log.WithFields(log.Fields{"ProcessID": process.ID}).Info("Done executing, closing process as successful")
				err = e.client.Close(process.ID, e.executorPrvKey)
				if err != nil {
					log.WithFields(log.Fields{"ProcessId": process.ID, "Error": err}).Error("Failed to close process as successful")
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
