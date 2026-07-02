package fileproject

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"
)

type SyncState struct {
	SchemaVersion        int            `json:"schemaVersion"`
	DeviceID             string         `json:"deviceId"`
	LastSentCheckpointID string         `json:"lastSentCheckpointId,omitempty"`
	LastTakeAt           *time.Time     `json:"lastTakeAt,omitempty"`
	PendingUpload        *PendingUpload `json:"pendingUpload,omitempty"`
}

type PendingUpload struct {
	CheckpointID string     `json:"checkpointId"`
	Attempts     int        `json:"attempts"`
	LastError    string     `json:"lastError,omitempty"`
	UpdatedAt    *time.Time `json:"updatedAt,omitempty"`
}

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
	if err := os.MkdirAll(filepath.Join(root, ".edda"), 0o755); err != nil {
		return fmt.Errorf("create .edda directory: %w", err)
	}
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal sync state: %w", err)
	}
	data = append(data, '\n')
	if err := os.WriteFile(syncStatePath(root), data, 0o644); err != nil {
		return fmt.Errorf("write sync state: %w", err)
	}
	return nil
}

func RecordPendingUpload(root string, checkpointID string) (SyncState, error) {
	state, err := EnsureSyncState(root)
	if err != nil {
		return SyncState{}, err
	}
	now := time.Now().UTC()
	state.PendingUpload = &PendingUpload{
		CheckpointID: checkpointID,
		Attempts:     0,
		UpdatedAt:    &now,
	}
	if err := WriteSyncState(root, state); err != nil {
		return SyncState{}, err
	}
	return state, nil
}

func CompletePendingUpload(root string) (SyncState, error) {
	state, err := EnsureSyncState(root)
	if err != nil {
		return SyncState{}, err
	}
	if state.PendingUpload == nil {
		return state, nil
	}
	state.PendingUpload.Attempts++
	state.LastSentCheckpointID = state.PendingUpload.CheckpointID
	state.PendingUpload = nil
	if err := WriteSyncState(root, state); err != nil {
		return SyncState{}, err
	}
	return state, nil
}

func RecordPendingUploadFailure(root string, message string) (SyncState, error) {
	state, err := EnsureSyncState(root)
	if err != nil {
		return SyncState{}, err
	}
	if state.PendingUpload == nil {
		return state, nil
	}
	now := time.Now().UTC()
	state.PendingUpload.Attempts++
	state.PendingUpload.LastError = message
	state.PendingUpload.UpdatedAt = &now
	if err := WriteSyncState(root, state); err != nil {
		return SyncState{}, err
	}
	return state, nil
}

func RecordTake(root string) (SyncState, error) {
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

func randomDeviceID() (string, error) {
	var data [8]byte
	if _, err := rand.Read(data[:]); err != nil {
		return "", fmt.Errorf("generate device id: %w", err)
	}
	return "device-" + hex.EncodeToString(data[:]), nil
}
