package executor

type Executor struct {
	coloniesServerHost      string
	coloniesServerPort      int
	coloniesInsecure        bool
	colonyId                string
	executorId              string
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
	dbUser                  string
	dbPassword              string
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
		e.colonyId = id
	}
}

func WithExecutorID(id string) ExecutorOption {
	return func(e *Executor) {
		e.executorId = id
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

func CreateExecutor(opts ...ExecutorOption) *Executor {
	executor := &Executor{}
	for _, opt := range opts {
		opt(executor)
	}

	return executor
}

func (e *Executor) ServeForEver() error {
	return nil
}
