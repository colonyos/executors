package executor

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
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

	return core.CreateExecutor(executorID, "ml", core.GenerateRandomID(), colonyID, time.Now(), time.Now()), executorID, executorPrvKey, nil
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

		codeSrc, ok := process.FunctionSpec.KwArgs["code_src"].(string)
		if !ok {
			errMsg := "Failed to parse code_src kwarg"
			log.WithFields(log.Fields{"ProcessID": process.ID}).Error(errMsg)
			e.client.Fail(process.ID, []string{errMsg}, e.executorPrvKey)
			continue
		}
		codeDst, ok := process.FunctionSpec.KwArgs["code_dst"].(string)
		if !ok {
			errMsg := "Failed to parse code_dst kwarg"
			log.WithFields(log.Fields{"ProcessID": process.ID}).Error(errMsg)
			e.client.Fail(process.ID, []string{errMsg}, e.executorPrvKey)
			continue
		}
		dataSrc, ok := process.FunctionSpec.KwArgs["data_src"].(string)
		if !ok {
			errMsg := "Failed to parse data_src kwarg"
			log.WithFields(log.Fields{"ProcessID": process.ID}).Error(errMsg)
			e.client.Fail(process.ID, []string{errMsg}, e.executorPrvKey)
			continue
		}
		dataDst, ok := process.FunctionSpec.KwArgs["data_dst"].(string)
		if !ok {
			errMsg := "Failed to parse data_dst kwarg"
			log.WithFields(log.Fields{"ProcessID": process.ID}).Error(errMsg)
			e.client.Fail(process.ID, []string{errMsg}, e.executorPrvKey)
			continue
		}
		out, ok := process.FunctionSpec.KwArgs["out"].(string)
		if !ok {
			errMsg := "Failed to parse out kwarg"
			log.WithFields(log.Fields{"ProcessID": process.ID}).Error(errMsg)
			e.client.Fail(process.ID, []string{errMsg}, e.executorPrvKey)
			continue
		}
		s3URL, ok := process.FunctionSpec.KwArgs["s3_url"].(string)
		if !ok {
			errMsg := "Failed to parse s3_url kwarg"
			log.WithFields(log.Fields{"ProcessID": process.ID}).Error(errMsg)
			e.client.Fail(process.ID, []string{errMsg}, e.executorPrvKey)
			continue
		}
		s3Bucket, ok := process.FunctionSpec.KwArgs["s3_bucket"].(string)
		if !ok {
			errMsg := "Failed to parse s3_bucket kwarg"
			log.WithFields(log.Fields{"ProcessID": process.ID}).Error(errMsg)
			e.client.Fail(process.ID, []string{errMsg}, e.executorPrvKey)
			continue
		}
		s3AccessKey, ok := process.FunctionSpec.KwArgs["s3_accesskey"].(string)
		if !ok {
			errMsg := "Failed to parse s3_accesskey kwarg"
			log.WithFields(log.Fields{"ProcessID": process.ID}).Error(errMsg)
			e.client.Fail(process.ID, []string{errMsg}, e.executorPrvKey)
			continue
		}
		s3SecretKey, ok := process.FunctionSpec.KwArgs["s3_secretkey"].(string)
		if !ok {
			errMsg := "Failed to parse s3_secretkey kwarg"
			log.WithFields(log.Fields{"ProcessID": process.ID}).Error(errMsg)
			e.client.Fail(process.ID, []string{errMsg}, e.executorPrvKey)
			continue
		}
		cmd, ok := process.FunctionSpec.KwArgs["cmd"].(string)
		if !ok {
			errMsg := "Failed to parse cmd kwarg"
			log.WithFields(log.Fields{"ProcessID": process.ID}).Error(errMsg)
			e.client.Fail(process.ID, []string{errMsg}, e.executorPrvKey)
			continue
		}
		argsIf := process.FunctionSpec.KwArgs["args"]

		fmt.Println(codeSrc)
		fmt.Println(codeDst)
		fmt.Println(dataSrc)
		fmt.Println(dataDst)
		fmt.Println(out)
		fmt.Println(s3URL)
		fmt.Println(s3Bucket)
		fmt.Println(s3AccessKey)
		fmt.Println(s3SecretKey)
		fmt.Println(cmd)
		fmt.Println(argsIf)

		useTLS := true
		region := ""
		s3, err := CreateS3(s3URL, region, s3AccessKey, s3SecretKey, useTLS, false, s3Bucket, "/tmp")
		if err != nil {
			errMsg := "Failed to create s3 client"
			log.WithFields(log.Fields{"ProcessID": process.ID, "Error": err}).Error(errMsg)
			e.client.Fail(process.ID, []string{errMsg}, e.executorPrvKey)
			continue
		}

		err = os.MkdirAll(codeDst, os.ModePerm)
		if err != nil {
			errMsg := "Failed to create code dst dir"
			log.WithFields(log.Fields{"ProcessID": process.ID, "Error": err}).Error(errMsg)
			e.client.Fail(process.ID, []string{errMsg}, e.executorPrvKey)
			continue
		}

		err = os.MkdirAll(dataDst, os.ModePerm)
		if err != nil {
			errMsg := "Failed to create data dst dir"
			log.WithFields(log.Fields{"ProcessID": process.ID, "Error": err}).Error(errMsg)
			e.client.Fail(process.ID, []string{errMsg}, e.executorPrvKey)
			continue
		}

		codeSrc = strings.TrimSuffix(codeSrc, "/")
		codeSrc = strings.TrimPrefix(codeSrc, "/")
		sourceFiles, err := s3.List(codeSrc + "/")
		if err != nil {
			errMsg := "Failed to list source files"
			log.WithFields(log.Fields{"ProcessID": process.ID, "Error": err}).Error(errMsg)
			e.client.Fail(process.ID, []string{errMsg}, e.executorPrvKey)
			continue
		}

		for _, sourceFile := range sourceFiles {
			f, err := s3.Download(sourceFile)
			if err != nil {
				errMsg := "Failed to download source code file"
				log.WithFields(log.Fields{"ProcessID": process.ID, "Error": err}).Error(errMsg)
				e.client.Fail(process.ID, []string{errMsg}, e.executorPrvKey)
				continue
			}

			filename := filepath.Base(sourceFile)
			fmt.Println(filename)
			err = os.WriteFile(codeDst+"/"+filename, *f, 0644)
			if err != nil {
				errMsg := "Failed to save source file " + filename + " to " + codeDst
				log.WithFields(log.Fields{"ProcessID": process.ID, "Error": err}).Error(errMsg)
				e.client.Fail(process.ID, []string{errMsg}, e.executorPrvKey)
				continue
			}
		}

		dataSrc = strings.TrimSuffix(dataSrc, "/")
		dataSrc = strings.TrimPrefix(dataSrc, "/")
		sourceFiles, err = s3.List(dataSrc + "/")
		if err != nil {
			errMsg := "Failed to list source files"
			log.WithFields(log.Fields{"ProcessID": process.ID, "Error": err}).Error(errMsg)
			e.client.Fail(process.ID, []string{errMsg}, e.executorPrvKey)
			continue
		}

		for _, sourceFile := range sourceFiles {
			f, err := s3.Download(sourceFile)
			if err != nil {
				errMsg := "Failed to download data file"
				log.WithFields(log.Fields{"ProcessID": process.ID, "Error": err}).Error(errMsg)
				e.client.Fail(process.ID, []string{errMsg}, e.executorPrvKey)
				continue
			}

			filename := filepath.Base(sourceFile)
			err = os.WriteFile(dataDst+"/"+filename, *f, 0644)
			if err != nil {
				errMsg := "Failed to save data file " + filename + " to " + codeDst
				log.WithFields(log.Fields{"ProcessID": process.ID, "Error": err}).Error(errMsg)
				e.client.Fail(process.ID, []string{errMsg}, e.executorPrvKey)
				continue
			}
		}

		argsIfArray := argsIf.([]interface{})
		arrStrArray := make([]string, len(argsIfArray))

		for i, v := range argsIfArray {
			arrStrArray[i] = v.(string)
		}
		argsStr := strArr2Str(ifArr2StringArr(argsIfArray))
		log.WithFields(log.Fields{"FuncName": process.FunctionSpec.FuncName, "Args": argsStr}).Info("Register function and lauching process")

		//e.client.CloseWithOutput(process.ID, nil, e.executorPrvKey)

		function := &core.Function{ExecutorID: e.executorID, ColonyID: e.colonyID, FuncName: process.FunctionSpec.FuncName}
		e.client.AddFunction(function, e.executorPrvKey)

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
			continue
		}

		command.Stderr = command.Stdout

		if err = command.Start(); err != nil {
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

		if err = command.Wait(); err != nil {
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
