package runtime

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	goruntime "runtime"
	"strings"
	"time"
)

var ErrInvalidRequest = errors.New("invalid script runtime request")

type Runner struct{}

func NewRunner() *Runner {
	return &Runner{}
}

func (r *Runner) Run(ctx context.Context, request RunRequest) (RunResult, error) {
	if strings.TrimSpace(request.Command) == "" {
		return RunResult{}, ErrInvalidRequest
	}
	if request.Timeout <= 0 {
		request.Timeout = 5 * time.Second
	}
	if request.MaxStdoutBytes <= 0 {
		request.MaxStdoutBytes = 64 * 1024
	}
	if request.MaxStderrBytes <= 0 {
		request.MaxStderrBytes = 16 * 1024
	}
	if request.Input.RuntimeVersion == "" {
		request.Input.RuntimeVersion = RuntimeVersion
	}

	inputBytes, err := json.Marshal(request.Input)
	if err != nil {
		return RunResult{}, fmt.Errorf("marshal script input: %w", err)
	}
	workdir, err := os.MkdirTemp("", "open-edda-skill-script-*")
	if err != nil {
		return RunResult{}, fmt.Errorf("create runtime temp dir: %w", err)
	}
	defer os.RemoveAll(workdir)

	runCtx, cancel := context.WithTimeout(ctx, request.Timeout)
	defer cancel()

	commandName, commandArgs := shellCommand(request.Command)
	cmd := exec.CommandContext(runCtx, commandName, commandArgs...)
	cmd.Dir = workdir
	cmd.Env = []string{
		"OPEN_EDDA_SKILL_RUNTIME=1",
		"NO_COLOR=1",
	}
	cmd.Stdin = bytes.NewReader(inputBytes)

	stdout := limitedBuffer{limit: request.MaxStdoutBytes}
	stderr := limitedBuffer{limit: request.MaxStderrBytes}
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	start := time.Now()
	err = cmd.Run()
	result := RunResult{
		Status:     StatusSucceeded,
		StdoutText: stdout.String(),
		StderrText: stderr.String(),
		Duration:   time.Since(start),
	}
	if cmd.ProcessState != nil {
		result.ExitCode = cmd.ProcessState.ExitCode()
	}

	if runCtx.Err() == context.DeadlineExceeded {
		result.Status = StatusTimedOut
		result.ErrorMessage = "script timed out"
		return result, nil
	}
	if err != nil {
		result.Status = StatusFailed
		result.ErrorMessage = err.Error()
		return result, nil
	}
	if stdout.truncated {
		result.Status = StatusRejected
		result.ErrorMessage = "script stdout exceeded byte limit"
		return result, nil
	}

	var output ScriptOutput
	if err := json.Unmarshal([]byte(result.StdoutText), &output); err != nil {
		result.Status = StatusRejected
		result.ErrorMessage = "script stdout must be valid JSON"
		return result, nil
	}
	if err := validateOutput(output); err != nil {
		result.Status = StatusRejected
		result.ErrorMessage = err.Error()
		return result, nil
	}
	result.Output = output
	return result, nil
}

func validateOutput(output ScriptOutput) error {
	switch output.Kind {
	case OutputKindReport, OutputKindProposal, OutputKindDraft, OutputKindGeneratedData:
	default:
		return fmt.Errorf("unsupported output kind %q", output.Kind)
	}
	if strings.TrimSpace(output.Title) == "" {
		return errors.New("script output title is required")
	}
	if strings.TrimSpace(output.Markdown) == "" && len(output.Proposals) == 0 && len(output.GeneratedData) == 0 {
		return errors.New("script output must include markdown, proposals, or generatedData")
	}
	for _, proposal := range output.Proposals {
		if strings.TrimSpace(proposal.TargetType) == "" || strings.TrimSpace(proposal.Title) == "" {
			return errors.New("proposal targetType and title are required")
		}
	}
	return nil
}

func shellCommand(command string) (string, []string) {
	if goruntime.GOOS == "windows" {
		return "cmd", []string{"/C", command}
	}
	return "sh", []string{"-c", command}
}

type limitedBuffer struct {
	buf       bytes.Buffer
	limit     int64
	truncated bool
}

func (b *limitedBuffer) Write(p []byte) (int, error) {
	if b.limit <= 0 {
		return len(p), nil
	}
	remaining := b.limit - int64(b.buf.Len())
	if remaining <= 0 {
		b.truncated = true
		return len(p), nil
	}
	if int64(len(p)) > remaining {
		b.truncated = true
		p = p[:remaining]
	}
	_, _ = b.buf.Write(p)
	return len(p), nil
}

func (b *limitedBuffer) String() string {
	return b.buf.String()
}
