package fileproject

import (
	"errors"
	"io/fs"
	"sync"
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
	if len(pending.PendingUploads) != 1 || pending.PendingUploads[0].CheckpointID != "checkpoint-1" {
		t.Fatalf("pending upload queue = %#v", pending.PendingUploads)
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

func TestPendingUploadQueuePreservesMultipleCheckpoints(t *testing.T) {
	root := t.TempDir()
	if _, err := RecordPendingUpload(root, "checkpoint-1"); err != nil {
		t.Fatalf("RecordPendingUpload first error = %v", err)
	}
	queued, err := RecordPendingUpload(root, "checkpoint-2")
	if err != nil {
		t.Fatalf("RecordPendingUpload second error = %v", err)
	}
	if len(queued.PendingUploads) != 2 {
		t.Fatalf("pending upload queue = %#v", queued.PendingUploads)
	}
	if queued.PendingUpload == nil || queued.PendingUpload.CheckpointID != "checkpoint-1" {
		t.Fatalf("active pending upload = %#v", queued.PendingUpload)
	}

	sent, err := CompletePendingUpload(root)
	if err != nil {
		t.Fatalf("CompletePendingUpload error = %v", err)
	}
	if sent.LastSentCheckpointID != "checkpoint-1" {
		t.Fatalf("last sent checkpoint = %q", sent.LastSentCheckpointID)
	}
	if sent.PendingUpload == nil || sent.PendingUpload.CheckpointID != "checkpoint-2" {
		t.Fatalf("next pending upload = %#v", sent.PendingUpload)
	}
}

func TestSyncStateMutatorsSerializeConcurrentUpdates(t *testing.T) {
	root := t.TempDir()
	var waitGroup sync.WaitGroup
	errs := make(chan error, 20)

	for index := 0; index < 10; index++ {
		waitGroup.Add(2)
		go func(index int) {
			defer waitGroup.Done()
			_, err := RecordPendingUpload(root, "checkpoint-"+string(rune('a'+index)))
			errs <- err
		}(index)
		go func() {
			defer waitGroup.Done()
			_, err := RecordTake(root)
			errs <- err
		}()
	}
	waitGroup.Wait()
	close(errs)

	for err := range errs {
		if err != nil {
			t.Fatalf("sync mutator error = %v", err)
		}
	}
	state, err := ReadSyncState(root)
	if err != nil {
		t.Fatalf("ReadSyncState error = %v", err)
	}
	if state.DeviceID == "" {
		t.Fatalf("device id is empty")
	}
	if state.LastTakeAt == nil {
		t.Fatalf("last take time is nil")
	}
	if state.PendingUpload == nil {
		t.Fatalf("pending upload is nil")
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
