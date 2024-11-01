package command

import (
	"io"
	"os/exec"
)

func Make(writer io.Writer, name string, arguments ...string) *exec.Cmd {
	cmd := exec.Command(name, arguments...)
	cmd.Stdout = writer
	cmd.Stderr = writer
	return cmd
}
