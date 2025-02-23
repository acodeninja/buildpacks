package main

import (
	"github.com/acodeninja/buildpacks/common/apt"
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

	result.Layers = append(result.Layers, apt.CreateLayerContributor([]string{
		"libxml2-dev",
		"libxmlsec1",
		"libxmlsec1-dev",
		"libxmlsec1-openssl",
		"pkg-config",
		"xmlsec1",
	}, "apt", b.Logger, true))

	return result, err
}
