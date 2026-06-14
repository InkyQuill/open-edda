//go:build unix

package runtime

import (
	"os"
	"os/exec"
	"syscall"
)

func configureCommand(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}

func cancelCommand(cmd *exec.Cmd) error {
	if cmd.Process == nil {
		return os.ErrProcessDone
	}
	if err := syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL); err != nil {
		if errorsIsProcessDone(err) {
			return os.ErrProcessDone
		}
		return err
	}
	return nil
}

func errorsIsProcessDone(err error) bool {
	return err == syscall.ESRCH
}
