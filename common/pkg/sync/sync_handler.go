package sync

import (
	"errors"
	"os"
	"strconv"
	"strings"

	"github.com/colonyos/colonies/pkg/client"
	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/fs"
	"github.com/colonyos/executors/common/pkg/debug"
	"github.com/colonyos/executors/common/pkg/failure"
	log "github.com/sirupsen/logrus"
)

type SyncHandler struct {
	colonyID       string
	executorPrvKey string
	client         *client.ColoniesClient
	fsDir          string
	failureHandler *failure.FailureHandler
	debugHandler   *debug.DebugHandler
}

func CreateSyncHandler(colonyID string,
	executorPrvKey string,
	client *client.ColoniesClient,
	fsDir string,
	failureHandler *failure.FailureHandler,
	debugHandler *debug.DebugHandler) (*SyncHandler, error) {
	if client == nil {
		return nil, errors.New("colonies client is nil")
	}

	if failureHandler == nil {
		return nil, errors.New("colonies failure client is nil")
	}

	return &SyncHandler{colonyID: colonyID, executorPrvKey: executorPrvKey, client: client, fsDir: fsDir, failureHandler: failureHandler, debugHandler: debugHandler}, nil
}

func (syncHandler *SyncHandler) PreSync(process *core.Process,
	debugHandler *debug.DebugHandler,
	failureHandler *failure.FailureHandler) error {
	if process.FunctionSpec.Filesystem.Mount != "" {
		err := syncHandler.DownloadSnapshots(process)
		if err != nil {
			failureHandler.HandleError(process, err, "Failed to download snapshots")
			return err
		}
		onProcessStart := true
		err = syncHandler.Sync(process, onProcessStart)
		if err != nil {
			failureHandler.HandleError(process, err, "Failed to sync to filesystem, onProcessStart="+strconv.FormatBool(onProcessStart))
			return err
		}
	} else {
		debugHandler.LogInfo(process, "Ignoring downloading snapshots and syncing dirs as mount not definied")
	}

	return nil
}

func (syncHandler *SyncHandler) PostSync(process *core.Process,
	debugHandler *debug.DebugHandler,
	failureHandler *failure.FailureHandler,
	fsDir string,
	client *client.ColoniesClient,
	colonyID string,
	executorPrvKey string) error {
	if process.FunctionSpec.Filesystem.Mount != "" {
		for _, snapshotMount := range process.FunctionSpec.Filesystem.SnapshotMounts {
			log.WithFields(log.Fields{"ProcessID": process.ID, "SnapshotID": snapshotMount.SnapshotID}).Info("Downloading snapshots")
			if !snapshotMount.KeepSnaphot {
				if snapshotMount.SnapshotID != "" {
					err := client.DeleteSnapshotByID(colonyID, snapshotMount.SnapshotID, executorPrvKey)
					if err != nil {
						log.WithFields(log.Fields{"SnapshotID": snapshotMount.SnapshotID, "Error": err}).Error("Failed to delete snapshot")
					} else {
						log.WithFields(log.Fields{"SnapshotID": snapshotMount.SnapshotID}).Info("Snapshot deleted")
					}
				}
				if !snapshotMount.KeepFiles {
					d := fsDir + snapshotMount.Dir
					d = strings.Replace(d, "{processid}", process.ID, 1)
					debugHandler.LogInfo(process, "Deleting snapshot mount dir: "+d)
					err := os.RemoveAll(d)
					if err != nil {
						failureHandler.HandleError(nil, err, "Failed to delete snapshot files")
						return err
					}
				}
			}
		}
	} else {
		debugHandler.LogInfo(process, "Ignoring cleaning up snapshots as mount not definied")
	}

	onProcessStart := false
	log.WithFields(log.Fields{"ProcessID": process.ID}).Info("Syncing filesystem")
	err := syncHandler.Sync(process, onProcessStart)
	if err != nil {
		failureHandler.HandleError(process, err, "Failed to sync to filesystem, onProcessStart="+strconv.FormatBool(onProcessStart))
		return err
	}
	if process.FunctionSpec.Filesystem.Mount != "" {
		for _, syncDirMount := range process.FunctionSpec.Filesystem.SyncDirMounts {
			if !syncDirMount.KeepFiles {
				d := fsDir + syncDirMount.Dir
				d = strings.Replace(d, "{processid}", process.ID, 1)
				debugHandler.LogInfo(process, "Deleting syncdir mount: "+d)
				err := os.RemoveAll(d)
				if err != nil {
					failureHandler.HandleError(nil, err, "Failed to delete syncdir files")
					return err
				}
			}
		}
	} else {
		debugHandler.LogInfo(process, "Ignoring syncing dirs as mount not definied")
	}

	return nil
}

func (syncHandler *SyncHandler) DownloadSnapshots(process *core.Process) error {
	filesystem := process.FunctionSpec.Filesystem
	fsClient, err := fs.CreateFSClient(syncHandler.client, syncHandler.colonyID, syncHandler.executorPrvKey)
	if err != nil {
		syncHandler.failureHandler.HandleError(process, err, "Failed to create FSClient, trying to download snapshots")
		return err
	}

	for _, snapshotMount := range filesystem.SnapshotMounts {
		if snapshotMount.SnapshotID != "" {
			snapshot, err := syncHandler.client.GetSnapshotByID(syncHandler.colonyID, snapshotMount.SnapshotID, syncHandler.executorPrvKey)
			if err != nil {
				syncHandler.failureHandler.HandleError(process, err, "Failed to resolve snapshotID")
				return err
			}

			newDir := syncHandler.fsDir + snapshotMount.Dir
			newDir = strings.Replace(newDir, "{processid}", process.ID, 1)
			err = os.MkdirAll(newDir, 0755)
			if err != nil {
				syncHandler.failureHandler.HandleError(process, err, "Failed to create download dir")
				return err
			}

			syncHandler.debugHandler.LogInfo(process, "Creating directory: "+newDir)
			syncHandler.debugHandler.LogInfo(process, "Downloading snapshot: Label:"+snapshot.Label+" SnapshotID:"+snapshot.ID+" Dir:"+newDir)
			log.WithFields(log.Fields{"Label": snapshotMount.Label, "SnapshotId": snapshot.ID, "Dir": snapshotMount.Dir}).Info("Downloading snapshot")
			err = fsClient.DownloadSnapshot(snapshot.ID, newDir)
			if err != nil {
				syncHandler.failureHandler.HandleError(process, err, "Failed to download snapshot")
				return err
			}
		} else {
			log.Info("Ignoring downloading snapshot as snapshot Id was not set")
		}
	}

	return nil
}

func (syncHandler *SyncHandler) Sync(process *core.Process, onProcessStart bool) error {
	filesystem := process.FunctionSpec.Filesystem
	fsClient, err := fs.CreateFSClient(syncHandler.client, syncHandler.colonyID, syncHandler.executorPrvKey)
	if err != nil {
		syncHandler.failureHandler.HandleError(process, err, "Failed to create FSClient, trying to sync")
		return err
	}

	for _, syncDirMount := range filesystem.SyncDirMounts {
		d := syncHandler.fsDir + syncDirMount.Dir
		d = strings.Replace(d, "{processid}", process.ID, 1)
		l := strings.Replace(syncDirMount.Label, "{processid}", process.ID, 1)
		if l != "" && d != "" {
			err = os.MkdirAll(d, 0755)
			if err != nil {
				log.WithFields(log.Fields{"Error": err}).Error("Failed to create download dir")
			}

			keepLocal := false
			if onProcessStart {
				keepLocal = syncDirMount.ConflictResolution.OnStart.KeepLocal
			} else {
				keepLocal = syncDirMount.ConflictResolution.OnClose.KeepLocal
			}
			syncplan, err := fsClient.CalcSyncPlan(d, l, keepLocal)
			if err != nil {
				log.WithFields(log.Fields{"Error": err}).Error("Failed to sync")
				return err
			}

			strategy := "keeplocal"
			if !keepLocal {
				strategy = "keepremote"
			}

			syncHandler.debugHandler.LogInfo(process, "Syncing cfs: label:"+l+" dir:"+d+" download:"+strconv.Itoa(len(syncplan.LocalMissing))+" upload:"+strconv.Itoa(len(syncplan.RemoteMissing))+" conflicts:"+strconv.Itoa(len(syncplan.RemoteMissing))+" conflictstrategy:"+strategy)
			err = fsClient.ApplySyncPlan(syncHandler.colonyID, syncplan)
			if err != nil {
				syncHandler.failureHandler.HandleError(process, err, "Failed to apply syncplan, Label:"+l+" Dir:"+d)
				return err
			}
		} else {
			log.Warn("Cannot perform directory synchronization, Label and Dir were not set")
		}
	}

	return nil
}
