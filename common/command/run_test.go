package command

import (
	"bytes"
	"strings"
	"testing"
)

func TestRunningACommandWithoutArguments(t *testing.T) {
	writer := new(bytes.Buffer)
	err := Run(writer, "ls")
	if err != nil {
		t.Errorf("%s", err)
	}
	if !strings.Contains(writer.String(), "make_test.go") {
		t.Fatalf("Output did not contain make_test.go")
	}
}

func TestRunningACommandWithArguments(t *testing.T) {
	writer := new(bytes.Buffer)
	err := Run(writer, "ls", "../")
	if err != nil {
		t.Errorf("%s", err)
	}
	if !strings.Contains(writer.String(), "command") {
		t.Fatalf("Output did not contain command")
	}
}
