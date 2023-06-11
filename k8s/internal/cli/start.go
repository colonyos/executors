package cli

import (
	"errors"
	"os"
	"strconv"

	"github.com/colonyos/colonies/pkg/security"
	"github.com/colonyos/executors/k8s/pkg/build"
	"github.com/colonyos/executors/k8s/pkg/executor"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(startCmd)
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start executor",
	Long:  "Start executor",
	Run: func(cmd *cobra.Command, args []string) {
		parseEnv()

		if Verbose {
			log.SetLevel(log.DebugLevel)
		}

		log.WithFields(log.Fields{
			"Verbose":            Verbose,
			"ColoniesServerHost": ColoniesServerHost,
			"ColoniesServerPort": ColoniesServerPort,
			"ColoniesInsecure":   ColoniesInsecure,
			"ColonyId":           ColonyID,
			"ColonyPrvKey":       "***********************",
			"ExecutorNamespace":  ExecutorNamespace,
			"ExecutorId":         ExecutorID,
			"ExecutorPrvKey":     "***********************"}).
			Info("Starting a Colonies K8s executor")

		executor, err := executor.CreateExecutor(
			executor.WithColoniesServerHost(ColoniesServerHost),
			executor.WithColoniesServerPort(ColoniesServerPort),
			executor.WithColoniesInsecure(ColoniesInsecure),
			executor.WithColonyID(ColonyID),
			executor.WithColonyPrvKey(ColonyPrvKey),
			executor.WithExecutorID(ExecutorID),
			executor.WithExecutorPrvKey(ExecutorPrvKey),
			executor.WithExecutorNamespace(ExecutorNamespace),
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

	if ExecutorNamespace == "" {
		ExecutorNamespace = os.Getenv("EXECUTOR_NAMESPACE")
	}
	if ExecutorNamespace == "" {
		CheckError(errors.New("Unknown Executor namespace"))
	}
}

func CheckError(err error) {
	if err != nil {
		log.WithFields(log.Fields{"Error": err, "BuildVersion": build.BuildVersion, "BuildTime": build.BuildTime}).Error(err.Error())
		os.Exit(-1)
	}
}
