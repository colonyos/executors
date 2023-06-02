package cli

import (
	"errors"
	"os"
	"strconv"

	"github.com/colonyos/colonies/pkg/security"
	"github.com/colonyos/executors/backup/pkg/build"
	"github.com/colonyos/executors/backup/pkg/executor"
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
			"Verbose":                 Verbose,
			"ColoniesServerHost":      ColoniesServerHost,
			"ColoniesServerPort":      ColoniesServerPort,
			"ColoniesInsecure":        ColoniesInsecure,
			"ColonyId":                ColonyID,
			"ExecutorId":              ExecutorID,
			"ExecutorPrvKey":          "***********************",
			"AWSS3Secure":             AWSS3Secure,
			"AWSS3InsecureSkipVerify": AWSS3InsecureSkipVerify,
			"AWSS3Endpoint":           "AWSS3Endpoint",
			"AWSS3Region":             AWSS3Region,
			"AWSS3AccessKey":          AWSS3AccessKey,
			"AWSS3SecretAccessKey":    AWSS3SecretAccessKey,
			"AWSS3BucketName":         AWSS3BucketName,
			"DBHost":                  DBHost,
			"DBPort":                  DBPort,
			"DBUser":                  DBUser,
			"DBPassword":              "***********************"}).
			Info("Starting a Colonies PostgreSQL Backup Executor")

		executor := executor.CreateExecutor(
			executor.WithColoniesServerHost(ColoniesServerHost),
			executor.WithColoniesServerPort(ColoniesServerPort),
			executor.WithColoniesInsecure(ColoniesInsecure),
			executor.WithColonyID(ColonyID),
			executor.WithExecutorID(ExecutorID),
			executor.WithExecutorPrvKey(ExecutorPrvKey),
			executor.WithAWSS3Secure(AWSS3Secure),
			executor.WithAWSS3InsecureSkipVerify(AWSS3InsecureSkipVerify),
			executor.WithAWSS3Endpoint(AWSS3Endpoint),
			executor.WithAWSS3Region(AWSS3Region),
			executor.WithAWSS3AccessKey(AWSS3AccessKey),
			executor.WithAWSS3SecretAccessKey(AWSS3SecretAccessKey),
			executor.WithDBHost(DBHost),
			executor.WithDBPort(DBPort),
			executor.WithDBUser(DBUser),
			executor.WithDBPassword(DBPassword),
		)

		err := executor.ServeForEver()
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

	AWSS3SecureStr := os.Getenv("AWS_S3_SECURE")
	if AWSS3SecureStr != "" {
		boolValue, err := strconv.ParseBool(AWSS3SecureStr)
		CheckError(err)
		AWSS3Secure = boolValue
	}

	AWSS3InsecureSkipVerifyStr := os.Getenv("AWS_S3_INSECURE_SKIP_VERIFY")
	if AWSS3InsecureSkipVerifyStr != "" {
		boolValue, err := strconv.ParseBool(AWSS3InsecureSkipVerifyStr)
		CheckError(err)
		AWSS3InsecureSkipVerify = boolValue
	}

	AWSS3EndpointStr := os.Getenv("AWS_S3_ENDPOINT")
	if AWSS3EndpointStr != "" {
		AWSS3Endpoint = AWSS3EndpointStr
	}

	AWSS3RegionStr := os.Getenv("AWS_S3_REGION")
	if AWSS3RegionStr != "" {
		AWSS3Region = AWSS3RegionStr
	}

	AWSS3AccessKeyStr := os.Getenv("AWS_S3_ACCESS_KEY")
	if AWSS3AccessKeyStr != "" {
		AWSS3AccessKey = AWSS3AccessKeyStr
	}

	AWSS3SecretAccessKeyStr := os.Getenv("AWS_S3_SECRET_ACCESS_KEY")
	if AWSS3SecretAccessKeyStr != "" {
		AWSS3SecretAccessKey = AWSS3SecretAccessKeyStr
	}

	AWSS3BucketNameStr := os.Getenv("AWS_S3_BUCKET_NAME")
	if AWSS3BucketNameStr != "" {
		AWSS3BucketName = AWSS3BucketNameStr
	}

	DBHostStr := os.Getenv("DB_HOST")
	if DBHostStr != "" {
		DBHost = DBHostStr
	}

	DBPortStr := os.Getenv("DB_PORT")
	if DBPortStr != "" {
		DBPort, err = strconv.Atoi(DBPortStr)
		CheckError(err)
	}

	DBUserStr := os.Getenv("DB_USER")
	if DBUserStr != "" {
		DBUser = DBUserStr
	}

	DBPasswordStr := os.Getenv("DB_PASSWORD")
	if DBPasswordStr != "" {
		DBPassword = DBPasswordStr
	}
}

func CheckError(err error) {
	if err != nil {
		log.WithFields(log.Fields{"Error": err, "BuildVersion": build.BuildVersion, "BuildTime": build.BuildTime}).Error(err.Error())
		os.Exit(-1)
	}
}
