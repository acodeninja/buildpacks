package main

import (
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

	result.Layers = append(result.Layers, NewGraphvizLayer(context, b.Logger))

	return result, err
}
