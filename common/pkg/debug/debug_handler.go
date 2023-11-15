package debug

import (
	"errors"

	"github.com/colonyos/colonies/pkg/client"
	"github.com/colonyos/colonies/pkg/core"
	log "github.com/sirupsen/logrus"
)

type DebugHandler struct {
	executorPrvKey string
	client         *client.ColoniesClient
	addLogs        bool
}

func CreateDebugHandler(executorPrvKey string, client *client.ColoniesClient, addLogs bool) (*DebugHandler, error) {
	if client == nil {
		return nil, errors.New("colonies client is nil")
	}

	return &DebugHandler{executorPrvKey: executorPrvKey, client: client, addLogs: addLogs}, nil
}

func (handler *DebugHandler) LogInfo(process *core.Process, infoMsg string) {
	log.WithFields(log.Fields{"ProcessID": process.ID}).Info(infoMsg)
	if process != nil {
		if handler.addLogs {
			err1 := handler.client.AddLog(process.ID, "ColonyOS: "+infoMsg+"\n", handler.executorPrvKey)
			if err1 != nil {
				log.WithFields(log.Fields{"ProcessId": process.ID, "Error": err1}).Error("Failed to add info log to process")
			}
		}
	}
}

func (handler *DebugHandler) LogError(process *core.Process, err error, errMsg string) {
	if err != nil {
		msg := "ColonyOS: "
		if errMsg != "" {
			msg += err.Error() + ":" + errMsg
		} else {
			msg += err.Error()
		}
		log.WithFields(log.Fields{"ProcessID": process.ID, "Error": err, "ErrMsg": errMsg}).Warn("Closing process as failed")
		if process != nil {
			err1 := handler.client.AddLog(process.ID, msg+"\n", handler.executorPrvKey)
			if err1 != nil {
				log.WithFields(log.Fields{"ProcessId": process.ID, "Error": err1}).Error("Failed to add error log to process")
			}
		}
	}
}
