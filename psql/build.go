package main

import (
	"fmt"
	"strings"

	"github.com/buildpacks/libcnb"
	"github.com/paketo-buildpacks/libpak/bard"
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
			if version, ok := entry.Metadata["psql-version"].(string); ok {
				result.Layers = append(result.Layers, NewPostgresClientLayer(version, context.Buildpack.Info.Version, b.Logger))
			}
		default:
			return libcnb.BuildResult{}, fmt.Errorf("received unexpected buildpack plan entry %q", entry.Name)
		}
	}

	return result, err
}
