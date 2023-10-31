package cli

import (
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(syncCmd)
}

func parseSyncEnv() {
	//var err error
	// ColoniesServerHostEnv := os.Getenv("COLONIES_SERVER_HOST")
	// if ColoniesServerHostEnv != "" {
	// 	ColoniesServerHost = ColoniesServerHostEnv
	// }

	// ColoniesServerPortEnvStr := os.Getenv("COLONIES_SERVER_PORT")
	// if ColoniesServerPortEnvStr != "" {
	// 	ColoniesServerPort, err = strconv.Atoi(ColoniesServerPortEnvStr)
	// 	CheckError(err)
	// }

	// ColoniesTLSEnv := os.Getenv("COLONIES_TLS")
	// if ColoniesTLSEnv == "true" {
	// 	ColoniesUseTLS = true
	// 	ColoniesInsecure = false
	// } else if ColoniesTLSEnv == "false" {
	// 	ColoniesUseTLS = false
	// 	ColoniesInsecure = true
	// }

	// if ColonyID == "" {
	// 	ColonyID = os.Getenv("COLONIES_COLONY_ID")
	// }

	// if ExecutorID == "" {
	// 	ExecutorID = os.Getenv("COLONIES_EXECUTOR_ID")
	// }
	// if ExecutorID == "" {
	// 	CheckError(errors.New("Unknown Executor Id"))
	// }

	ProcessID = os.Getenv("PROCESS_ID")
}

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Start a sync server",
	Long:  "Start a sync server",
	Run: func(cmd *cobra.Command, args []string) {
		if Verbose {
			log.SetLevel(log.DebugLevel)
		}

		parseSyncEnv()

		// syncServer := fs.CreateSyncTool(ProcessID)

		// err := syncServer.Start()
	},
}
