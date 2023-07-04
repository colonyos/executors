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

	return core.CreateExecutor(executorID, "unix", core.GenerateRandomID(), colonyID, time.Now(), time.Now()), executorID, executorPrvKey, nil
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

		argsStr := strArr2Str(ifArr2StringArr(process.FunctionSpec.Args))
		log.WithFields(log.Fields{"FuncName": process.FunctionSpec.FuncName, "Args": argsStr}).Info("Register function and lauching process")

		function := &core.Function{ExecutorID: e.executorID, ColonyID: e.colonyID, FuncName: process.FunctionSpec.FuncName}
		e.client.AddFunction(function, e.executorPrvKey)

		execCmd := make([]string, 0)
		for _, arg := range ifArr2StringArr(process.FunctionSpec.Args) {
			execCmd = append(execCmd, arg)
		}

		execCmd = append([]string{process.FunctionSpec.FuncName}, execCmd...)
		execCmdStr := strings.Join(execCmd[:], " ")

		cmd := exec.Command("sh", "-c", execCmdStr)
		cmd.Env = os.Environ()
		for _, attribute := range process.Attributes {
			cmd.Env = append(cmd.Env, attribute.Key+"="+attribute.Value)
		}

		cmd.Env = append(cmd.Env, "COLONIES_COLONY_ID="+e.colonyID)
		cmd.Env = append(cmd.Env, "COLONIES_PROCESS_ID="+process.ID)
		cmd.Env = append(cmd.Env, "COLONIES_SERVER_HOST="+e.coloniesServerHost)
		cmd.Env = append(cmd.Env, "COLONIES_SERVER_PORT="+strconv.Itoa(e.coloniesServerPort))
		cmd.Env = append(cmd.Env, "COLONIES_EXECUTOR_ID="+e.executorID)
		cmd.Env = append(cmd.Env, "COLONIES_EXECUTOR_PRVKEY="+e.executorPrvKey)

		// Get output
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			log.WithFields(log.Fields{"Error": err}).Info("Failed to open stdout")
			err = e.client.Fail(process.ID, []string{"Failed to open stdout"}, e.executorPrvKey)
			continue
		}

		cmd.Stderr = cmd.Stdout

		if err = cmd.Start(); err != nil {
			log.WithFields(log.Fields{"Error": err}).Info("Failed to start cmd")
			err = e.client.Fail(process.ID, []string{"Failed to start cmd"}, e.executorPrvKey)
			continue
		}

		output := ""
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
					continue
				}
				defer f.Close()
				_, err = f.WriteString(string(tmp))
				if err != nil {
					log.WithFields(log.Fields{"Error": err}).Info("Failed to write to logfile")
					err = e.client.Fail(process.ID, []string{"Failed to write to logile"}, e.executorPrvKey)
					continue
				}
			}
			output += string(bytes.Trim(tmp, "\x00"))
		}

		if err = cmd.Wait(); err != nil {
			log.WithFields(log.Fields{"ProcessID": process.ID}).Info("Closing process as failed")
			e.client.Fail(process.ID, []string{output}, e.executorPrvKey)
			continue
		}

		log.WithFields(log.Fields{"ProcessID": process.ID}).Info("Closing process as successful")
		ifArr := make([]interface{}, 1)
		ifArr[0] = output
		e.client.CloseWithOutput(process.ID, ifArr, e.executorPrvKey)
	}
}
