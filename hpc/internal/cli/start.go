package cli

import (
	"errors"
	"os"
	"strconv"

	"github.com/colonyos/colonies/pkg/security"
	"github.com/colonyos/executors/hpc/pkg/build"
	"github.com/colonyos/executors/hpc/pkg/executor"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(startCmd)
	startCmd.Flags().StringVarP(&LogDir, "logdir", "", "", "Log directory")
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start executor",
	Long:  "Start executor",
	Run: func(cmd *cobra.Command, args []string) {
		parseEnv()

		fsDir := os.Getenv("EXECUTOR_FS_DIR")
		logDir := os.Getenv("EXECUTOR_LOG_DIR")
		imageDir := os.Getenv("EXECUTOR_IMAGE_DIR")

		executorType := os.Getenv("EXECUTOR_TYPE")
		if executorType == "" {
			CheckError(errors.New("Executor type not specifed, defaulting to hpc"))
		}

		swName := os.Getenv("EXECUTOR_SW_NAME")
		swType := os.Getenv("EXECUTOR_SW_TYPE")
		swVersion := os.Getenv("EXECUTOR_SW_VERSION")
		hwCPU := os.Getenv("EXECUTOR_HW_CPU")
		hwModel := os.Getenv("EXECUTOR_HW_MODEL")

		hwNodesStr := os.Getenv("EXECUTOR_HW_NODES")
		hwNodes, err := strconv.Atoi(hwNodesStr)
		CheckError(err)

		hwMem := os.Getenv("EXECUTOR_HW_MEM")
		hwStorage := os.Getenv("EXECUTOR_HW_STORAGE")
		hwGPUCountStr := os.Getenv("EXECUTOR_HW_GPU_COUNT")
		hwGPUCount, err := strconv.Atoi(hwGPUCountStr)
		CheckError(err)

		executor, err := executor.CreateExecutor(
			executor.WithVerbose(Verbose),
			executor.WithColoniesServerHost(ColoniesServerHost),
			executor.WithColoniesServerPort(ColoniesServerPort),
			executor.WithColoniesInsecure(ColoniesInsecure),
			executor.WithColonyID(ColonyID),
			executor.WithColonyPrvKey(ColonyPrvKey),
			executor.WithExecutorID(ExecutorID),
			executor.WithExecutorPrvKey(ExecutorPrvKey),
			executor.WithLogDir(logDir),
			executor.WithFsDir(fsDir),
			executor.WithImageDir(imageDir),
			executor.WithSoftwareName(swName),
			executor.WithSoftwareType(swType),
			executor.WithSoftwareVersion(swVersion),
			executor.WithHardwareCPU(hwCPU),
			executor.WithHardwareModel(hwModel),
			executor.WithHardwareNodes(hwNodes),
			executor.WithHardwareMemory(hwMem),
			executor.WithHardwareStorage(hwStorage),
			executor.WithHardwareGPUCount(hwGPUCount),
			executor.WithExecutorType(executorType),
		)
		CheckError(err)

		err = executor.ServeForEver()
		CheckError(err)
	},
}

func parseEnv() {
	var err error
	ColoniesServerHostEnv := os.Getenv("COLONIES_SERVER_HOST")
	if ColoniesServerHostEnv != "" {
		ColoniesServerHost = ColoniesServerHostEnv
	}

	ColoniesServerPortEnvStr := os.Getenv("COLONIES_SERVER_PORT")
	if ColoniesServerPortEnvStr != "" {
		ColoniesServerPort, err = strconv.Atoi(ColoniesServerPortEnvStr)
		CheckError(err)
	}

	ColoniesTLSEnv := os.Getenv("COLONIES_TLS")
	if ColoniesTLSEnv == "true" {
		ColoniesUseTLS = true
		ColoniesInsecure = false
	} else if ColoniesTLSEnv == "false" {
		ColoniesUseTLS = false
		ColoniesInsecure = true
	}

	VerboseEnv := os.Getenv("COLONIES_VERBOSE")
	if VerboseEnv == "true" {
		Verbose = true
	} else if VerboseEnv == "false" {
		Verbose = false
	}

	if ColonyID == "" {
		ColonyID = os.Getenv("COLONIES_COLONY_ID")
	}
	if ColonyID == "" {
		CheckError(errors.New("Unknown Colony Id"))
	}

	if ColonyPrvKey == "" {
		ColonyPrvKey = os.Getenv("COLONIES_COLONY_PRVKEY")
	}

	if ExecutorID == "" {
		ExecutorID = os.Getenv("COLONIES_EXECUTOR_ID")
	}
	if ExecutorID == "" {
		CheckError(errors.New("Unknown Executor Id"))
	}

	keychain, err := security.CreateKeychain(KEYCHAIN_PATH)
	CheckError(err)

	if ExecutorPrvKey == "" {
		ExecutorPrvKey = os.Getenv("COLONIES_EXECUTOR_PRVKEY")
	}
	if ExecutorPrvKey == "" {
		ExecutorPrvKey, err = keychain.GetPrvKey(ExecutorID)
		CheckError(err)
	}
}

func CheckError(err error) {
	if err != nil {
		log.WithFields(log.Fields{"Error": err, "BuildVersion": build.BuildVersion, "BuildTime": build.BuildTime}).Error(err.Error())
		os.Exit(-1)
	}
}
