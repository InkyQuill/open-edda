package fileproject

import (
	"errors"
	"io/fs"
	"testing"
)

func TestEnsureSyncStateCreatesLocalState(t *testing.T) {
	root := t.TempDir()
	state, err := EnsureSyncState(root)
	if err != nil {
		t.Fatalf("EnsureSyncState error = %v", err)
	}
	if state.DeviceID == "" {
		t.Fatalf("device id is empty")
	}
	read, err := ReadSyncState(root)
	if err != nil {
		t.Fatalf("ReadSyncState error = %v", err)
	}
	if read.DeviceID != state.DeviceID {
		t.Fatalf("read device id = %q, want %q", read.DeviceID, state.DeviceID)
	}
}

func TestReadSyncStateMissing(t *testing.T) {
	_, err := ReadSyncState(t.TempDir())
	if !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("ReadSyncState missing error = %v, want fs.ErrNotExist", err)
	}
}

func TestPendingUploadLifecycle(t *testing.T) {
	root := t.TempDir()
	pending, err := RecordPendingUpload(root, "checkpoint-1")
	if err != nil {
		t.Fatalf("RecordPendingUpload error = %v", err)
	}
	if pending.PendingUpload == nil || pending.PendingUpload.CheckpointID != "checkpoint-1" {
		t.Fatalf("pending upload = %#v", pending.PendingUpload)
	}

	failed, err := RecordPendingUploadFailure(root, "offline")
	if err != nil {
		t.Fatalf("RecordPendingUploadFailure error = %v", err)
	}
	if failed.PendingUpload == nil || failed.PendingUpload.Attempts != 1 || failed.PendingUpload.LastError != "offline" {
		t.Fatalf("failed pending upload = %#v", failed.PendingUpload)
	}

	sent, err := CompletePendingUpload(root)
	if err != nil {
		t.Fatalf("CompletePendingUpload error = %v", err)
	}
	if sent.PendingUpload != nil {
		t.Fatalf("pending upload not cleared = %#v", sent.PendingUpload)
	}
	if sent.LastSentCheckpointID != "checkpoint-1" {
		t.Fatalf("last sent checkpoint = %q", sent.LastSentCheckpointID)
	}
}

func TestRecordTakeStoresCursorTime(t *testing.T) {
	root := t.TempDir()
	state, err := RecordTake(root)
	if err != nil {
		t.Fatalf("RecordTake error = %v", err)
	}
	if state.LastTakeAt == nil {
		t.Fatalf("last take time is nil")
	}
}
