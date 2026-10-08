package main

import (
	"context"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/InkyQuill/open-edda/fileproject"
)

type syncOutput struct {
	mu     sync.Mutex
	writer io.Writer
}

func (w *syncOutput) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.writer.Write(p)
}

// The command owns the heartbeat and joins it before returning. Slow requests
// remain visible even when no bytes have arrived; no guessed overall percentage.
func observeSync(ctx context.Context, output io.Writer, quiet bool) (context.Context, io.Writer, func()) {
	if quiet {
		return ctx, output, func() {}
	}
	writer := &syncOutput{writer: output}
	var mu sync.Mutex
	var latest fileproject.Progress
	var started, last time.Time
	emit := func() {
		if latest.Phase == "" {
			return
		}
		var line strings.Builder
		elapsed := time.Since(started).Round(time.Second)
		if latest.Total > 0 {
			fmt.Fprintf(&line, "%s: %d/%d (%d%%), %.1f MiB, %s", latest.Phase, latest.Done, latest.Total, 100*latest.Done/latest.Total, float64(latest.Bytes)/(1<<20), elapsed)
		} else if latest.Done > 0 || latest.Bytes > 0 {
			fmt.Fprintf(&line, "%s: %d entries, %.1f MiB, %s", latest.Phase, latest.Done, float64(latest.Bytes)/(1<<20), elapsed)
		} else {
			fmt.Fprintf(&line, "%s: %s", latest.Phase, elapsed)
		}
		if latest.Path != "" {
			fmt.Fprintf(&line, " — %q", latest.Path)
		}
		fmt.Fprintln(writer, line.String())
		last = time.Now()
	}
	observe := func(p fileproject.Progress) {
		mu.Lock()
		defer mu.Unlock()
		changed := p.Phase != latest.Phase
		if changed {
			if latest.Phase != "" && (latest.Total == 0 || latest.Done != latest.Total) {
				emit()
			}
			started = time.Now()
		}
		latest = p
		if changed || time.Since(last) >= time.Second || p.Total > 0 && p.Done == p.Total {
			emit()
		}
	}
	done, joined := make(chan struct{}), make(chan struct{})
	go func() {
		defer close(joined)
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				mu.Lock()
				if time.Since(last) >= time.Second {
					emit()
				}
				mu.Unlock()
			}
		}
	}()
	return fileproject.WithProgress(ctx, observe), writer, func() { close(done); <-joined }
}
func syncPhase(ctx context.Context, phase string) {
	fileproject.ReportProgress(ctx, fileproject.Progress{Phase: phase})
}

// Report bytes as the transport consumes them, including while one large file
// is in flight. The observer throttles display independently of reads.
type progressReader struct {
	io.Reader
	advance func(int)
}

func (r progressReader) Read(p []byte) (int, error) {
	n, err := r.Reader.Read(p)
	if n > 0 {
		r.advance(n)
	}
	return n, err
}
