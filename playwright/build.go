package main

import (
	"fmt"
	"github.com/acodeninja/buildpacks/common"
	"github.com/acodeninja/buildpacks/common/apt"
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

	_, err = common.InitialiseBuild(context, b.Logger)
	if err != nil {
		return result, err
	}

	for _, entry := range context.Plan.Entries {
		switch strings.ToLower(entry.Name) {
		case "playwright":
			version := entry.Metadata["playwright-version"]
			language := entry.Metadata["playwright-language"]
			temporaryLayer, err := context.Layers.Layer("temporary")
			if err != nil {
				return result, err
			}
			result.Layers = append(result.Layers, apt.CreateLayerContributor(UbuntuPackages, "dependencies", b.Logger, true))
			result.Layers = append(result.Layers, NewPlaywrightLayer(version.(string), language.(string), temporaryLayer, b.Logger))
		default:
			return libcnb.BuildResult{}, fmt.Errorf("received unexpected buildpack plan entry %q", entry.Name)
		}
	}

	return result, err
}
