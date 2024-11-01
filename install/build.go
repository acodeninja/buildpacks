package main

import (
	"errors"
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
		case PlanEntryAptInstall:
			packages, err := packagesFromPlanEntryMetadata(entry)
			if err != nil {
				return libcnb.BuildResult{}, fmt.Errorf("failed to resolve packages from plan metadata:\n%w", err)
			}

			result.Layers = append(result.Layers, apt.CreateLayerContributor(packages, "apt-install", b.Logger, true))
		default:
			return libcnb.BuildResult{}, fmt.Errorf("received unexpected buildpack plan entry %q", entry.Name)
		}
	}

	return result, err
}

func packagesFromPlanEntryMetadata(entry libcnb.BuildpackPlanEntry) ([]string, error) {
	rawPackages, ok := entry.Metadata["packages"]
	if !ok {
		return nil, errors.New(fmt.Sprintf("%s build plan entry is missing required metadata key \"packages\"", entry.Name))
	}

	pathArr, ok := rawPackages.([]interface{})
	if !ok {
		return nil, errors.New("expected \"packages\" to be of type []interface{}")
	}

	packages := make([]string, len(pathArr))
	for i, path := range pathArr {
		var ok bool
		packages[i], ok = path.(string)
		if !ok {
			return nil, errors.New("expected each item in \"packages\" to be of type string")
		}
	}

	return packages, nil
}
