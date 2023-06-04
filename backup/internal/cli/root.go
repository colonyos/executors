package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

const TimeLayout = "2006-01-02 15:04:05"
const KEYCHAIN_PATH = ".colonies"

var Verbose bool
var ColoniesServerHost string
var ColoniesServerPort int
var ColoniesInsecure bool
var ColoniesSkipTLSVerify bool
var ColoniesUseTLS bool
var ColonyID string
var ColonyPrvKey string
var ExecutorName string
var ExecutorID string
var ExecutorType string
var ExecutorPrvKey string
var AWSS3Secure bool
var AWSS3InsecureSkipVerify bool
var AWSS3Endpoint string
var AWSS3Region string
var AWSS3AccessKey string
var AWSS3SecretAccessKey string
var AWSS3BucketName string
var DBHost string
var DBPort int
var DBUser string
var DBDatabase string
var DBPassword string
var FullBackups int
var BackupPath string

func init() {
	rootCmd.PersistentFlags().BoolVarP(&Verbose, "verbose", "v", false, "verbose output")
}

var rootCmd = &cobra.Command{
	Use:   "backup",
	Short: "Colonies PostgreSQL Backup Executor",
	Long:  "Colonies PostgreSQL Backup Executor",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
