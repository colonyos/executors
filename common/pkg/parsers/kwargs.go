package parsers

import (
	"errors"
	"fmt"
	"strings"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/executors/common/pkg/debug"
	"github.com/colonyos/executors/common/pkg/failure"
)

type KwArgs struct {
	Image        string
	RebuildImage bool
	Cmd          string
	Args         string
	ExecCmd      string
}

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

func ParseKwArgs(process *core.Process, failureHandler *failure.FailureHandler, debugHandler *debug.DebugHandler) (*KwArgs, error) {
	imageIf := process.FunctionSpec.KwArgs["docker-image"]
	image, ok := imageIf.(string)
	if !ok {
		err := errors.New("Failed to parse docker image flag")
		failureHandler.HandleError(process, err, "")
		return nil, err
	}

	rebuildImageIf := process.FunctionSpec.KwArgs["rebuild-image"]
	rebuildImage, ok := rebuildImageIf.(bool)
	if !ok {
		debugHandler.LogInfo(process, "Failed to parse rebuild image flag, setting rebuildImage to false")
		rebuildImage = false
	}

	cmd, ok := process.FunctionSpec.KwArgs["cmd"].(string)
	if !ok {
		err := errors.New("Failed to parse cmd kwarg")
		failureHandler.HandleError(process, err, "")
		return nil, err
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
		debugHandler.LogInfo(process, "Failed to parse args, setting args to empty string")
		argsStr = ""
	}

	execCmd := make([]string, 0)
	for _, arg := range ifArr2StringArr(argsIfArray) {
		arg = strings.Replace(arg, "{processid}", process.ID, 1)
		execCmd = append(execCmd, arg)
	}

	execCmd = append([]string{cmd}, execCmd...)
	execCmdStr := strings.Join(execCmd[:], " ")

	return &KwArgs{Image: image, RebuildImage: rebuildImage, Cmd: cmd, Args: argsStr, ExecCmd: execCmdStr}, nil
}
