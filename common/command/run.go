package command

import "io"

func Run(writer io.Writer, name string, arguments ...string) error {
	return Make(writer, name, arguments...).Run()
}
