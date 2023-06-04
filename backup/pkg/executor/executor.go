package executor

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/colonyos/colonies/pkg/client"
	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/security/crypto"
	"github.com/colonyos/executors/backup/pkg/backup"
	"github.com/colonyos/executors/backup/pkg/s3"
	log "github.com/sirupsen/logrus"
)

type Executor struct {
	coloniesServerHost      string
	coloniesServerPort      int
	coloniesInsecure        bool
	colonyID                string
	colonyPrvKey            string
	executorID              string
	executorPrvKey          string
	awsS3Secure             bool
	awsS3InsecureSkipVerify bool
	awsS3Endpoint           string
	awsS3Region             string
	awsS3AccessKey          string
	awsS3SecretAccessKey    string
	awsS3BucketName         string
	dbHost                  string
	dbPort                  int
	dbDatabase              string
	dbUser                  string
	dbPassword              string
	fullbackups             int
	backupPath              string
	ctx                     context.Context
	cancel                  context.CancelFunc
	client                  *client.ColoniesClient
	s3                      *s3.S3
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

func WithAWSS3Secure(secure bool) ExecutorOption {
	return func(e *Executor) {
		e.awsS3Secure = secure
	}
}

func WithAWSS3InsecureSkipVerify(skipVerify bool) ExecutorOption {
	return func(e *Executor) {
		e.awsS3InsecureSkipVerify = skipVerify
	}
}

func WithAWSS3Endpoint(endpoint string) ExecutorOption {
	return func(e *Executor) {
		e.awsS3Endpoint = endpoint
	}
}

func WithAWSS3Region(region string) ExecutorOption {
	return func(e *Executor) {
		e.awsS3Region = region
	}
}

func WithAWSS3AccessKey(accessKey string) ExecutorOption {
	return func(e *Executor) {
		e.awsS3AccessKey = accessKey
	}
}

func WithAWSS3SecretAccessKey(secretAccessKey string) ExecutorOption {
	return func(e *Executor) {
		e.awsS3SecretAccessKey = secretAccessKey
	}
}

func WithAWSS3BucketName(bucketName string) ExecutorOption {
	return func(e *Executor) {
		e.awsS3BucketName = bucketName
	}
}

func WithDBHost(host string) ExecutorOption {
	return func(e *Executor) {
		e.dbHost = host
	}
}

func WithDBPort(port int) ExecutorOption {
	return func(e *Executor) {
		e.dbPort = port
	}
}

func WithDBDatabase(name string) ExecutorOption {
	return func(e *Executor) {
		e.dbDatabase = name
	}
}

func WithDBUser(user string) ExecutorOption {
	return func(e *Executor) {
		e.dbUser = user
	}
}

func WithDBPassword(password string) ExecutorOption {
	return func(e *Executor) {
		e.dbPassword = password
	}
}

func WithFullbackups(count int) ExecutorOption {
	return func(e *Executor) {
		e.fullbackups = count
	}
}

func WithBackupPath(path string) ExecutorOption {
	return func(e *Executor) {
		e.backupPath = path
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

	return core.CreateExecutor(executorID, "backup", core.GenerateRandomID(), colonyID, time.Now(), time.Now()), executorID, executorPrvKey, nil
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

	err := genCredFile(e.dbHost, e.dbPort, e.dbDatabase, e.dbUser, e.dbPassword)
	if err != nil {
		return nil, err
	}

	s3, err := s3.CreateS3Client(e.awsS3Endpoint, e.awsS3Region, e.awsS3AccessKey, e.awsS3SecretAccessKey, e.awsS3Secure, e.awsS3InsecureSkipVerify, e.awsS3BucketName, "backups")
	if err != nil {
		return nil, err
	}
	e.s3 = s3

	e.client = client.CreateColoniesClient(e.coloniesServerHost, e.coloniesServerPort, e.coloniesInsecure, e.awsS3InsecureSkipVerify)

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

func (e *Executor) removeOldBackups() error {
	backupFiles, err := e.s3.ListDirectory("/backups/")
	if err != nil {
		return err
	}

	filesToDelete := cleanup(e.fullbackups, backupFiles)
	for _, file := range filesToDelete {
		log.WithFields(log.Fields{"Filename": file}).Info("Removing old backup file")
		if strings.HasPrefix(file, "backups/backup_") {
			err = e.s3.RemoveFile(file)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (e *Executor) backup() (string, error) {
	log.Info("Backing up database ...")
	size, execTime, path, filename, err := backup.ExecBackupDB(e.backupPath)
	if err != nil {
		return "", err
	}
	log.WithFields(log.Fields{"Filename": filename, "ExecTime": strconv.FormatInt(execTime, 10) + " seconds", "Size": strconv.FormatInt(size, 10) + " bytes"}).Info("Successfully backup database")

	log.Info("Uploading backup file to S3 ...")
	execTimeS3, sizeS3, err := e.s3.Upload(path, filename, size)

	log.WithFields(log.Fields{"ExecTime": strconv.FormatInt(execTimeS3, 10) + " seconds", "Size": strconv.FormatInt(sizeS3, 10) + " bytes"}).Info("Successfully uploaded backup file")

	err = os.Remove(path + "/" + filename)
	log.WithFields(log.Fields{"Filepath": path + "/" + filename}).Info("Removed local backup file")

	if size != sizeS3 {
		return "", errors.New("File sizes missmatches")
	}

	log.Info("Removing old backup files ...")
	err = e.removeOldBackups()
	if err != nil {
		return "", err
	}

	log.Info("Backup procedure completed, closing process ...")
	result := Result{
		Filename:         filename,
		Bucket:           e.awsS3BucketName,
		FilePathS3:       "/backups/" + filename,
		ExecTimeBackup:   execTime,
		ExecTimeUploadS3: execTimeS3,
		Size:             size,
		SizeS3:           sizeS3,
	}

	json, err := result.ToJSON()
	if err != nil {
		return "", err
	}

	return json, nil
}

func (e *Executor) ServeForEver() error {
	for {
		process, err := e.client.AssignWithContext(e.colonyID, 100, e.ctx, e.executorPrvKey)
		if err != nil {
			log.Warn(err)
			log.Warn("Retrying in 5 seconds ...")
			time.Sleep(5 * time.Second)
			continue
		}
		log.WithFields(log.Fields{"ProcessID": process.ID, "ExecutorID": e.executorID}).Info("Assigned process to executor")

		funcName := process.FunctionSpec.FuncName
		if funcName == "backup" {
			result, err := e.backup()
			if err != nil {
				log.WithFields(log.Fields{"ProcessID": process.ID, "ExecutorID": e.executorID, "Err": err}).Info("Failed to call backup")
				err = e.client.Fail(process.ID, []string{err.Error()}, e.executorPrvKey)
			}
			output := make([]interface{}, 1)
			output[0] = result
			err = e.client.CloseWithOutput(process.ID, output, e.executorPrvKey)
		} else {
			log.WithFields(log.Fields{"ProcessID": process.ID, "ExecutorID": e.executorID, "FuncName": funcName}).Info("Unsupported function")
			err = e.client.Fail(process.ID, []string{"Unsupported function: " + funcName}, e.executorPrvKey)
		}
	}
}
