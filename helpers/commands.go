package helpers

import (
	"io"
	"os/exec"
)

func GetCommand(writer io.Writer, name string, arguments ...string) *exec.Cmd {
	cmd := exec.Command(name, arguments...)
	cmd.Stdout = writer
	cmd.Stderr = writer
	return cmd
}

func RunCommand(writer io.Writer, name string, arguments ...string) error {
	return GetCommand(writer, name, arguments...).Run()
}
