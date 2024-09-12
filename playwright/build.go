package main

import (
	"fmt"
	"github.com/acodeninja/buildpacks/helpers"
	"github.com/buildpacks/libcnb"
	"github.com/paketo-buildpacks/libpak/bard"
	"strings"
)

type Build struct {
	Logger bard.Logger
}

func (b Build) Build(context libcnb.BuildContext) (libcnb.BuildResult, error) {
	var err error

	result := libcnb.NewBuildResult()

	_, err = helpers.InitialiseBuild(context, b.Logger)
	if err != nil {
		return result, err
	}

	for _, entry := range context.Plan.Entries {
		switch strings.ToLower(entry.Name) {
		case "playwright-python":
			version := entry.Metadata["playwright-python-version"]
			result.Layers = append(result.Layers, helpers.NewAptLayer(UbuntuPackages, "dependencies", b.Logger, true))

			temporaryLayer, err := context.Layers.Layer("playwright-python")
			if err != nil {
				return result, err
			}

			result.Layers = append(result.Layers, NewPlaywrightLayer(version.(string), temporaryLayer, b.Logger))
		default:
			return libcnb.BuildResult{}, fmt.Errorf("received unexpected buildpack plan entry %q", entry.Name)
		}
	}

	return result, err
}
