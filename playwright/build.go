package main

import (
	"github.com/acodeninja/buildpacks/helpers"
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

	result.Layers = append(result.Layers, helpers.NewAptLayer(UbuntuPackages, "apt", b.Logger, true))

	temporaryLayer, err := context.Layers.Layer("apt-temporary")
	result.Layers = append(result.Layers, NewPlaywrightLayer(temporaryLayer))

	return result, err
}
