package helpers

import (
	"fmt"
	"github.com/buildpacks/libcnb"
	"io"
	"os"
	"os/exec"
	"slices"
	"strings"
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

func InjectLayerEnvironment(command *exec.Cmd, environment libcnb.Environment) {
	command.Env = append(command.Env, os.Environ()...)

	for name, value := range environment {
		if strings.HasSuffix(name, ".prepend") {
			variable := strings.TrimSuffix(name, ".prepend")

			idx := slices.IndexFunc(command.Env, func(c string) bool { return strings.HasPrefix(c, variable) })

			if idx != -1 {
				command.Env[idx] = fmt.Sprintf(
					"%s=%s:%s",
					variable,
					value,
					strings.TrimPrefix(command.Env[idx], fmt.Sprintf("%s=", variable)),
				)
			} else {
				command.Env = append(command.Env, fmt.Sprintf("%s=%s", variable, value))
			}
		}

		if strings.HasSuffix(name, ".append") {
			variable := strings.TrimSuffix(name, ".append")

			idx := slices.IndexFunc(command.Env, func(c string) bool { return strings.HasPrefix(c, variable) })

			if idx != -1 {
				command.Env[idx] = fmt.Sprintf(
					"%s=%s:%s",
					variable,
					strings.TrimPrefix(command.Env[idx], fmt.Sprintf("%s=", variable)),
					value,
				)
			} else {
				command.Env = append(command.Env, fmt.Sprintf("%s=%s", variable, value))
			}
		}
	}
}
