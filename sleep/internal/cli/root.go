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

func init() {
	rootCmd.PersistentFlags().BoolVarP(&Verbose, "verbose", "v", false, "verbose output")
}

var rootCmd = &cobra.Command{
	Use:   "sleep_executor",
	Short: "Colonies Sleep Executor",
	Long:  "Colonies Sleep Executor",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
