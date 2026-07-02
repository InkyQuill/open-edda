package fileproject

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
	"syscall"
	"time"
)

type SyncState struct {
	SchemaVersion        int             `json:"schemaVersion"`
	DeviceID             string          `json:"deviceId"`
	LastSentCheckpointID string          `json:"lastSentCheckpointId,omitempty"`
	LastTakeAt           *time.Time      `json:"lastTakeAt,omitempty"`
	PendingUpload        *PendingUpload  `json:"pendingUpload,omitempty"`
	PendingUploads       []PendingUpload `json:"pendingUploads,omitempty"`
}

type PendingUpload struct {
	CheckpointID string     `json:"checkpointId"`
	Attempts     int        `json:"attempts"`
	LastError    string     `json:"lastError,omitempty"`
	UpdatedAt    *time.Time `json:"updatedAt,omitempty"`
}

var syncStateMu sync.Mutex

func ReadSyncState(root string) (SyncState, error) {
	data, err := os.ReadFile(syncStatePath(root))
	if err != nil {
		if os.IsNotExist(err) {
			return SyncState{}, fs.ErrNotExist
		}
		return SyncState{}, fmt.Errorf("read sync state: %w", err)
	}
	var state SyncState
	if err := json.Unmarshal(data, &state); err != nil {
		return SyncState{}, fmt.Errorf("parse sync state: %w", err)
	}
	if state.DeviceID == "" {
		return SyncState{}, fmt.Errorf("sync state device id is required")
	}
	normalizePendingUploads(&state)
	return state, nil
}

func EnsureSyncState(root string) (SyncState, error) {
	state, err := ReadSyncState(root)
	if err == nil {
		if state.SchemaVersion == 0 {
			state.SchemaVersion = CurrentSchemaVersion
		}
		return state, nil
	}
	if err != fs.ErrNotExist {
		return SyncState{}, err
	}
	deviceID, err := randomDeviceID()
	if err != nil {
		return SyncState{}, err
	}
	state = SyncState{SchemaVersion: CurrentSchemaVersion, DeviceID: deviceID}
	if err := WriteSyncState(root, state); err != nil {
		return SyncState{}, err
	}
	return state, nil
}

func WriteSyncState(root string, state SyncState) error {
	if state.SchemaVersion == 0 {
		state.SchemaVersion = CurrentSchemaVersion
	}
	if state.DeviceID == "" {
		deviceID, err := randomDeviceID()
		if err != nil {
			return err
		}
		state.DeviceID = deviceID
	}
	normalizePendingUploads(&state)
	if err := os.MkdirAll(filepath.Join(root, ".edda"), 0o755); err != nil {
		return fmt.Errorf("create .edda directory: %w", err)
	}
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal sync state: %w", err)
	}
	data = append(data, '\n')
	target := syncStatePath(root)
	tmp, err := os.CreateTemp(filepath.Dir(target), "state-*.json")
	if err != nil {
		return fmt.Errorf("create temporary sync state: %w", err)
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("write temporary sync state: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("sync temporary sync state: %w", err)
	}
	if err := tmp.Chmod(0o644); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("chmod temporary sync state: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close temporary sync state: %w", err)
	}
	if err := os.Rename(tmpPath, target); err != nil {
		return fmt.Errorf("write sync state: %w", err)
	}
	return nil
}

func RecordPendingUpload(root string, checkpointID string) (SyncState, error) {
	syncStateMu.Lock()
	defer syncStateMu.Unlock()
	unlock, err := lockSyncStateFile(root)
	if err != nil {
		return SyncState{}, err
	}
	defer unlock()
	state, err := EnsureSyncState(root)
	if err != nil {
		return SyncState{}, err
	}
	now := time.Now().UTC()
	state.PendingUploads = append(state.PendingUploads, PendingUpload{
		CheckpointID: checkpointID,
		Attempts:     0,
		UpdatedAt:    &now,
	})
	normalizePendingUploads(&state)
	if err := WriteSyncState(root, state); err != nil {
		return SyncState{}, err
	}
	return state, nil
}

func CompletePendingUpload(root string) (SyncState, error) {
	syncStateMu.Lock()
	defer syncStateMu.Unlock()
	unlock, err := lockSyncStateFile(root)
	if err != nil {
		return SyncState{}, err
	}
	defer unlock()
	state, err := EnsureSyncState(root)
	if err != nil {
		return SyncState{}, err
	}
	if len(state.PendingUploads) == 0 {
		return state, nil
	}
	state.PendingUploads[0].Attempts++
	state.LastSentCheckpointID = state.PendingUploads[0].CheckpointID
	state.PendingUploads = state.PendingUploads[1:]
	state.PendingUpload = nil
	normalizePendingUploads(&state)
	if err := WriteSyncState(root, state); err != nil {
		return SyncState{}, err
	}
	return state, nil
}

func RecordPendingUploadFailure(root string, message string) (SyncState, error) {
	syncStateMu.Lock()
	defer syncStateMu.Unlock()
	unlock, err := lockSyncStateFile(root)
	if err != nil {
		return SyncState{}, err
	}
	defer unlock()
	state, err := EnsureSyncState(root)
	if err != nil {
		return SyncState{}, err
	}
	if len(state.PendingUploads) == 0 {
		return state, nil
	}
	now := time.Now().UTC()
	state.PendingUploads[0].Attempts++
	state.PendingUploads[0].LastError = message
	state.PendingUploads[0].UpdatedAt = &now
	normalizePendingUploads(&state)
	if err := WriteSyncState(root, state); err != nil {
		return SyncState{}, err
	}
	return state, nil
}

func RecordTake(root string) (SyncState, error) {
	syncStateMu.Lock()
	defer syncStateMu.Unlock()
	unlock, err := lockSyncStateFile(root)
	if err != nil {
		return SyncState{}, err
	}
	defer unlock()
	state, err := EnsureSyncState(root)
	if err != nil {
		return SyncState{}, err
	}
	now := time.Now().UTC()
	state.LastTakeAt = &now
	if err := WriteSyncState(root, state); err != nil {
		return SyncState{}, err
	}
	return state, nil
}

func syncStatePath(root string) string {
	return filepath.Join(root, ".edda", "state.local.json")
}

func lockSyncStateFile(root string) (func(), error) {
	eddaDir := filepath.Join(root, ".edda")
	if err := os.MkdirAll(eddaDir, 0o755); err != nil {
		return nil, fmt.Errorf("create .edda directory: %w", err)
	}
	file, err := os.OpenFile(filepath.Join(eddaDir, "state.local.lock"), os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		return nil, fmt.Errorf("open sync state lock: %w", err)
	}
	if err := syscall.Flock(int(file.Fd()), syscall.LOCK_EX); err != nil {
		_ = file.Close()
		return nil, fmt.Errorf("lock sync state: %w", err)
	}
	return func() {
		_ = syscall.Flock(int(file.Fd()), syscall.LOCK_UN)
		_ = file.Close()
	}, nil
}

func normalizePendingUploads(state *SyncState) {
	if len(state.PendingUploads) == 0 && state.PendingUpload != nil {
		state.PendingUploads = []PendingUpload{*state.PendingUpload}
	}
	if len(state.PendingUploads) == 0 {
		state.PendingUpload = nil
		return
	}
	state.PendingUpload = &state.PendingUploads[0]
}

func randomDeviceID() (string, error) {
	var data [8]byte
	if _, err := rand.Read(data[:]); err != nil {
		return "", fmt.Errorf("generate device id: %w", err)
	}
	return "device-" + hex.EncodeToString(data[:]), nil
}
