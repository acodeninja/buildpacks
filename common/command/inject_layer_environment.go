package command

import (
	"fmt"
	"github.com/buildpacks/libcnb"
	"os"
	"os/exec"
	"slices"
	"strings"
)

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
