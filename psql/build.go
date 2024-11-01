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
	var err error

	b.Logger.Title(context.Buildpack)

	result := libcnb.NewBuildResult()

	for _, entry := range context.Plan.Entries {
		switch strings.ToLower(entry.Name) {
		case "postgres-client":
			version := "14" // entry.Metadata["psql-version"]
			if err != nil {
				return result, err
			}
			//result.Layers = append(result.Layers, apt.CreateLayerContributor([]string{"postgresql-client"}, "dependencies", b.Logger, true))
			result.Layers = append(result.Layers, NewPostgresClientLayer(version, b.Logger))
		default:
			return libcnb.BuildResult{}, fmt.Errorf("received unexpected buildpack plan entry %q", entry.Name)
		}
	}

	return result, err
}
