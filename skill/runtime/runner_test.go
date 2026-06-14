package runtime

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestRunnerExecutesJSONEnvelope(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell-based runner test is Unix-only")
	}
	runner := NewRunner()

	result, err := runner.Run(context.Background(), RunRequest{
		Command:        fmt.Sprintf("GORACE=atexit_sleep_ms=0 OPEN_EDDA_TEST_HELPER=echo_report %q", os.Args[0]),
		Timeout:        time.Second,
		MaxStdoutBytes: 4096,
		MaxStderrBytes: 1024,
		Input: Envelope{
			RuntimeVersion: RuntimeVersion,
			Inputs: EnvelopeInputs{
				Arguments: map[string]any{"mode": "echo"},
			},
		},
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if result.Status != StatusSucceeded {
		t.Fatalf("status = %s, want %s; stderr=%s", result.Status, StatusSucceeded, result.StderrText)
	}
	if !json.Valid([]byte(result.StdoutText)) {
		t.Fatalf("stdout is not JSON: %q", result.StdoutText)
	}
	if result.Output.Kind != OutputKindReport {
		t.Fatalf("output kind = %s, want %s", result.Output.Kind, OutputKindReport)
	}
	if !strings.Contains(result.Output.Markdown, "mode=echo") {
		t.Fatalf("output markdown = %q, want echoed envelope argument", result.Output.Markdown)
	}
}

func TestRunnerRejectsInvalidOutputKind(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("sh-based runner test is Unix-only")
	}
	runner := NewRunner()

	result, err := runner.Run(context.Background(), RunRequest{
		Command:        `printf '%s' '{"kind":"mutation","title":"x","markdown":"x"}'`,
		Timeout:        time.Second,
		MaxStdoutBytes: 4096,
		MaxStderrBytes: 1024,
		Input:          Envelope{RuntimeVersion: RuntimeVersion},
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if result.Status != StatusRejected {
		t.Fatalf("status = %s, want %s", result.Status, StatusRejected)
	}
	if !strings.Contains(result.ErrorMessage, "unsupported output kind") {
		t.Fatalf("error = %q, want unsupported output kind", result.ErrorMessage)
	}
}

func TestRunnerTimesOut(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("sh-based runner test is Unix-only")
	}
	runner := NewRunner()

	result, err := runner.Run(context.Background(), RunRequest{
		Command:        "while :; do :; done",
		Timeout:        20 * time.Millisecond,
		MaxStdoutBytes: 4096,
		MaxStderrBytes: 1024,
		Input:          Envelope{RuntimeVersion: RuntimeVersion},
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if result.Status != StatusTimedOut {
		t.Fatalf("status = %s, want %s", result.Status, StatusTimedOut)
	}
}

func TestRunnerTimesOutChildProcessGroup(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("process-group runner test is Unix-only")
	}
	runner := NewRunner()
	start := time.Now()

	result, err := runner.Run(context.Background(), RunRequest{
		Command:        "sleep 10 & wait",
		Timeout:        20 * time.Millisecond,
		MaxStdoutBytes: 4096,
		MaxStderrBytes: 1024,
		Input:          Envelope{RuntimeVersion: RuntimeVersion},
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if result.Status != StatusTimedOut {
		t.Fatalf("status = %s, want %s", result.Status, StatusTimedOut)
	}
	if elapsed := time.Since(start); elapsed > time.Second {
		t.Fatalf("Run() took %s, want child process group cleanup within 1s", elapsed)
	}
}

func TestRunnerRejectsOversizedStdoutWithoutShortWriteFailure(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("sh-based runner test is Unix-only")
	}
	runner := NewRunner()

	result, err := runner.Run(context.Background(), RunRequest{
		Command:        `printf '%s' '{"kind":"report","title":"x","markdown":"this output is intentionally too long"}'`,
		Timeout:        time.Second,
		MaxStdoutBytes: 32,
		MaxStderrBytes: 1024,
		Input:          Envelope{RuntimeVersion: RuntimeVersion},
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if result.Status != StatusRejected {
		t.Fatalf("status = %s, want %s; error=%s", result.Status, StatusRejected, result.ErrorMessage)
	}
	if !strings.Contains(result.ErrorMessage, "stdout exceeded") {
		t.Fatalf("error = %q, want stdout cap/truncation error", result.ErrorMessage)
	}
	if strings.Contains(result.ErrorMessage, "short write") {
		t.Fatalf("error = %q, want no accidental short write failure", result.ErrorMessage)
	}
}

func TestShellCommandUsesPlatformShell(t *testing.T) {
	name, args := shellCommand("echo ok")
	if runtime.GOOS == "windows" {
		if name != "cmd" {
			t.Fatalf("command name = %q, want cmd", name)
		}
		if len(args) != 2 || args[0] != "/C" || args[1] != "echo ok" {
			t.Fatalf("command args = %#v, want /C command", args)
		}
		return
	}
	if name != "sh" {
		t.Fatalf("command name = %q, want sh", name)
	}
	if len(args) != 2 || args[0] != "-c" || args[1] != "echo ok" {
		t.Fatalf("command args = %#v, want -c command", args)
	}
}

func TestMain(m *testing.M) {
	if os.Getenv("OPEN_EDDA_TEST_HELPER") == "echo_report" {
		var input Envelope
		if err := json.NewDecoder(os.Stdin).Decode(&input); err != nil {
			fmt.Fprintf(os.Stderr, "decode envelope: %v", err)
			os.Exit(2)
		}
		mode, _ := input.Inputs.Arguments["mode"].(string)
		if err := json.NewEncoder(os.Stdout).Encode(ScriptOutput{
			Kind:     OutputKindReport,
			Title:    "Envelope report",
			Markdown: "mode=" + mode,
		}); err != nil {
			fmt.Fprintf(os.Stderr, "encode output: %v", err)
			os.Exit(2)
		}
		os.Exit(0)
	}
	os.Exit(m.Run())
}
