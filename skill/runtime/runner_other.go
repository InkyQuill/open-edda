//go:build !unix && !windows

package runtime

import (
	"os"
	"os/exec"
)

func configureCommand(cmd *exec.Cmd) {}

func cancelCommand(cmd *exec.Cmd) error {
	if cmd.Process == nil {
		return os.ErrProcessDone
	}
	return cmd.Process.Kill()
}
