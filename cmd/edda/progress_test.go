package main

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/InkyQuill/open-edda/fileproject"
)

type progressSignal struct {
	bytes.Buffer
	lines chan struct{}
}

func (w *progressSignal) Write(p []byte) (int, error) {
	n, err := w.Buffer.Write(p)
	select {
	case w.lines <- struct{}{}:
	default:
	}
	return n, err
}

func TestSyncProgressWhileWaitingAndQuiet(t *testing.T) {
	output := &progressSignal{lines: make(chan struct{}, 8)}
	ctx, _, finish := observeSync(context.Background(), output, false)
	syncPhase(ctx, "Checking remote manifest")
	<-output.lines // Immediate stage, before a slow operation completes.
	select {
	case <-output.lines: // Heartbeat while no new progress arrives.
	case <-time.After(3 * time.Second):
		finish()
		t.Fatal("slow operation stayed silent")
	}
	fileproject.ReportProgress(ctx, fileproject.Progress{Phase: "Uploading changed files", Done: 2, Total: 2, Bytes: 1024, Path: "chapter\n.md"})
	finish()
	if strings.Count(output.String(), "Checking remote manifest") < 2 || !strings.Contains(output.String(), "2/2 (100%)") || !strings.Contains(output.String(), `chapter\n.md`) {
		t.Fatalf("missing or unsafe progress: %s", output.String())
	}
	var quiet bytes.Buffer
	ctx, _, finish = observeSync(context.Background(), &quiet, true)
	syncPhase(ctx, "Hidden")
	finish()
	if quiet.Len() != 0 {
		t.Fatalf("quiet emitted progress: %s", &quiet)
	}
}
