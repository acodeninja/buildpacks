package apt

import (
	"github.com/buildpacks/libcnb"
	"github.com/paketo-buildpacks/libpak"
	"github.com/paketo-buildpacks/libpak/bard"
)

type Layer struct {
	Packages         []string
	LayerName        string
	LayerContributor libpak.LayerContributor
	Logger           bard.Logger
}

func CreateLayerContributor(packages []string, name string, logger bard.Logger, cache bool) *Layer {
	return &Layer{
		Packages:  packages,
		LayerName: name,
		LayerContributor: libpak.NewLayerContributor(
			name,
			map[string]interface{}{},
			libcnb.LayerTypes{
				Build:  true,
				Launch: true,
				Cache:  cache,
			},
		),
		Logger: logger,
	}
}

func (apt Layer) Contribute(layer libcnb.Layer) (libcnb.Layer, error) {
	apt.LayerContributor.Logger = apt.Logger
	apt.LayerContributor.ExpectedMetadata = map[string]interface{}{
		"packages": apt.Packages,
	}

	return apt.LayerContributor.Contribute(layer, func() (libcnb.Layer, error) {
		if layer.Metadata == nil {
			layer.Metadata = map[string]interface{}{}
		}
		layer.Metadata["packages"] = apt.Packages

		err := InstallAptPackages(layer, apt.Packages, apt.Logger, false)

		return layer, err
	})
}

func (apt Layer) Name() string {
	return apt.LayerName
}
