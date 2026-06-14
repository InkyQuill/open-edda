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
		Command:        fmt.Sprintf("OPEN_EDDA_TEST_HELPER=echo_report %q", os.Args[0]),
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
