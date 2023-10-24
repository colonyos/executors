package failure

import (
	"errors"

	"github.com/colonyos/colonies/pkg/client"
	"github.com/colonyos/colonies/pkg/core"
	log "github.com/sirupsen/logrus"
)

type FailureHandler struct {
	executorPrvKey string
	client         *client.ColoniesClient
}

func CreateFailureHandler(executorPrvKey string, client *client.ColoniesClient) (*FailureHandler, error) {
	if client == nil {
		return nil, errors.New("colonies client is nil")
	}

	return &FailureHandler{executorPrvKey: executorPrvKey, client: client}, nil
}

func (handler *FailureHandler) HandleError(process *core.Process, err error, errMsg string) {
	if err != nil {
		msg := "ColonyOS: "
		if errMsg != "" {
			msg += err.Error() + ":" + errMsg
		} else {
			msg += err.Error()
		}
		if process != nil {
			log.WithFields(log.Fields{"ProcessID": process.ID, "Error": err, "ErrMsg": errMsg}).Warn("Closing process as failed")
			err1 := handler.client.AddLog(process.ID, msg, handler.executorPrvKey)
			if err1 != nil {
				log.WithFields(log.Fields{"ProcessId": process.ID, "Error": err1}).Error("Failed to add error log to process")
			}
			err2 := handler.client.Fail(process.ID, []string{msg}, handler.executorPrvKey)
			if err2 != nil {
				log.WithFields(log.Fields{"ProcessId": process.ID, "Error": err2}).Error("Failed to close process as failed")
			}
		}
	}
}
