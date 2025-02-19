package main

import (
	"fmt"
	"github.com/buildpacks/libcnb"
	"github.com/paketo-buildpacks/libpak/bard"
	"strings"
)

type Build struct {
	Logger bard.Logger
}

func (b Build) Build(context libcnb.BuildContext) (libcnb.BuildResult, error) {
	b.Logger.Title(context.Buildpack)

	for _, entry := range context.Plan.Entries {
		switch strings.ToLower(entry.Name) {
		case "script-run":
			result, err := RunCommandWithBuildpackEnvironmentVariables(entry, b)

			return result, err

		default:
			return libcnb.BuildResult{}, fmt.Errorf("received unexpected buildpack plan entry %q", entry.Name)
		}
	}

	return libcnb.BuildResult{}, fmt.Errorf("no plans were run")
}
