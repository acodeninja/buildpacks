package main

import (
	"fmt"
	"github.com/acodeninja/buildpacks/common"
	"github.com/acodeninja/buildpacks/common/command"
	"github.com/buildpacks/libcnb"
	"github.com/paketo-buildpacks/libpak"
	"github.com/paketo-buildpacks/libpak/bard"
	"strings"
)

func ResolveScriptLocation(context libcnb.DetectContext, logger bard.Logger) (string, error) {
	var scriptLocation string

	cr, err := libpak.NewConfigurationResolver(context.Buildpack, &logger)
	if err != nil {
		return "", fmt.Errorf("unable to create configuration resolver\n%w", err)
	}

	scriptLocation, isSet := cr.Resolve("BP_RUN_SCRIPT_LOCATION")

	if !isSet {
		scriptLocation = "buildpack.run.sh"
	}

	scriptLocation = "/workspace/" + scriptLocation
	return scriptLocation, nil
}

func RunCommandWithBuildpackEnvironmentVariables(entry libcnb.BuildpackPlanEntry, b Build) (libcnb.BuildResult, error) {
	scriptLocation, ok := entry.Metadata["script-location"]
	if !ok {
		return libcnb.BuildResult{}, fmt.Errorf("failed to resolve script-location from plan metadata")
	}

	result := libcnb.NewBuildResult()

	variables, err := common.GetBuildpackGroupEnvironmentVariables(b.Logger)
	if err != nil {
		return result, err
	}

	runCommand := command.Make(
		common.IndentedWriterFactory(0, b.Logger),
		"bash",
		fmt.Sprintf("%s", scriptLocation),
	)

	runCommand.Env = variables.GetForCommand()

	b.Logger.Header("Running command")
	b.Logger.Body(fmt.Sprintf("bash %s", scriptLocation))
	b.Logger.Header("With environment")
	for _, v := range runCommand.Env {
		if !strings.HasPrefix(v, "CNB_") {
			b.Logger.Body(fmt.Sprintf("%s", v))
		}
	}
	b.Logger.Header("Output")
	err = runCommand.Run()
	if err != nil {
		return libcnb.BuildResult{}, err
	}
	return result, err
}
