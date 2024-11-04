package command

import (
	"bytes"
	"strings"
	"testing"
)

func TestMakingACommandWithoutArguments(t *testing.T) {
	writer := new(bytes.Buffer)
	command := Make(writer, "ls")

	err := command.Run()
	if err != nil {
		t.Errorf("%s", err)
	}
	if !strings.Contains(writer.String(), "make_test.go") {
		t.Fatalf("Output did not contain make_test.go")
	}
}

func TestMakingACommandWithArguments(t *testing.T) {
	writer := new(bytes.Buffer)
	command := Make(writer, "ls", "../")

	err := command.Run()
	if err != nil {
		t.Errorf("%s", err)
	}
	if !strings.Contains(writer.String(), "command") {
		t.Fatalf("Output did not contain command")
	}
}
