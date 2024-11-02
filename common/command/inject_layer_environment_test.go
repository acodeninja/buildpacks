package command

import (
	"bytes"
	"github.com/buildpacks/libcnb"
	"strings"
	"testing"
)

func TestInjectingALayerEnvironmentWithPrepend(t *testing.T) {
	writer := new(bytes.Buffer)
	command := Make(writer, "env")

	environment := make(libcnb.Environment)
	environment["TEST_ENV.prepend"] = "A_TEST_VARIABLE"

	InjectLayerEnvironment(command, environment)

	err := command.Run()
	if err != nil {
		t.Errorf("%s", err)
	}
	if !strings.Contains(writer.String(), "TEST_ENV") && !strings.Contains(writer.String(), "A_TEST_VARIABLE") {
		t.Fatalf("Output did not contain TEST_ENV")
	}
}

func TestInjectingALayerEnvironmentWithAppend(t *testing.T) {
	writer := new(bytes.Buffer)
	command := Make(writer, "env")

	environment := make(libcnb.Environment)
	environment["TEST_ENV.append"] = "A_TEST_VARIABLE"

	InjectLayerEnvironment(command, environment)

	err := command.Run()
	if err != nil {
		t.Errorf("%s", err)
	}
	if !strings.Contains(writer.String(), "TEST_ENV") && !strings.Contains(writer.String(), "A_TEST_VARIABLE") {
		t.Fatalf("Output did not contain TEST_ENV")
	}
}
