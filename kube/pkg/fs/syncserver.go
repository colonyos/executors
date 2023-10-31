package fs

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/colonyos/colonies/pkg/client"
	"github.com/colonyos/executors/common/pkg/debug"
	"github.com/colonyos/executors/common/pkg/failure"
	"github.com/colonyos/executors/common/pkg/sync"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

type SyncServer struct {
	serverPort         int
	verbose            bool
	coloniesServerHost string
	coloniesServerPort int
	coloniesInsecure   bool
	colonyID           string
	executorID         string
	executorPrvKey     string
	ginHandler         *gin.Engine
	fsDir              string
	httpServer         *http.Server
	client             *client.ColoniesClient
	syncHandler        *sync.SyncHandler
	failureHandler     *failure.FailureHandler
	debugHandler       *debug.DebugHandler
	ctx                context.Context
	cancel             context.CancelFunc
}

type ServerOption func(*SyncServer)

func WithVerbose(verbose bool) ServerOption {
	return func(s *SyncServer) {
		s.verbose = verbose
	}
}

func WithColoniesServerHost(host string) ServerOption {
	return func(s *SyncServer) {
		s.coloniesServerHost = host
	}
}

func WithServerPort(port int) ServerOption {
	return func(s *SyncServer) {
		s.serverPort = port
	}
}

func WithColoniesServerPort(port int) ServerOption {
	return func(s *SyncServer) {
		s.coloniesServerPort = port
	}
}

func WithColoniesInsecure(insecure bool) ServerOption {
	return func(s *SyncServer) {
		s.coloniesInsecure = insecure
	}
}

func WithColonyID(id string) ServerOption {
	return func(s *SyncServer) {
		s.colonyID = id
	}
}

func WithExecutorPrvKey(key string) ServerOption {
	return func(s *SyncServer) {
		s.executorPrvKey = key
	}
}

func WithFsDir(fsDir string) ServerOption {
	return func(s *SyncServer) {
		s.fsDir = fsDir
	}
}

func CreateSyncServer(opts ...ServerOption) (*SyncServer, error) {
	server := &SyncServer{}
	for _, opt := range opts {
		opt(server)
	}

	ctx, cancel := context.WithCancel(context.Background())
	server.ctx = ctx
	server.cancel = cancel

	sigc := make(chan os.Signal)
	signal.Notify(sigc, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT, syscall.SIGSEGV)
	go func() {
		<-sigc
		server.Shutdown()
		os.Exit(1)
	}()

	server.client = client.CreateColoniesClient(server.coloniesServerHost, server.coloniesServerPort, server.coloniesInsecure, false)

	server.ginHandler = gin.Default()
	httpServer := &http.Server{
		Addr:    ":" + strconv.Itoa(server.serverPort),
		Handler: server.ginHandler,
	}

	server.httpServer = httpServer

	server.setupRoutes()

	server.client = client.CreateColoniesClient(server.coloniesServerHost, server.coloniesServerPort, server.coloniesInsecure, false)

	var err error
	server.failureHandler, err = failure.CreateFailureHandler(server.executorPrvKey, server.client)
	if err != nil {
		return nil, err
	}

	server.debugHandler, err = debug.CreateDebugHandler(server.executorPrvKey, server.client)
	if err != nil {
		return nil, err
	}

	syncHandler, err := sync.CreateSyncHandler(server.colonyID,
		server.executorPrvKey,
		server.client,
		server.fsDir,
		server.failureHandler,
		server.debugHandler)

	server.syncHandler = syncHandler

	return server, nil
}

func (server *SyncServer) setupRoutes() {
	server.ginHandler.GET("/sync", server.handleSyncRequest)
}

func (server *SyncServer) handleSyncRequest(c *gin.Context) {
	initValue := c.DefaultQuery("init", "false")
	if initValue == "true" {
		c.JSON(200, gin.H{
			"message": "Initialization set to true.",
		})
	} else {
		c.JSON(200, gin.H{
			"message": "Initialization set to false.",
		})
	}
}

func (server *SyncServer) ServeForever() error {
	if err := server.httpServer.ListenAndServe(); err != nil && errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return nil
}

func (server *SyncServer) Shutdown() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.httpServer.Shutdown(ctx); err != nil {
		log.WithFields(log.Fields{"Error": err}).Warning("SyncServer forced to shutdown")
	}
}
