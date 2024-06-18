package helpers

import (
	"github.com/buildpacks/libcnb"
	"github.com/paketo-buildpacks/libpak"
	"github.com/paketo-buildpacks/libpak/bard"
)

func InitialiseBuild(context libcnb.BuildContext, logger bard.Logger) (*libpak.ConfigurationResolver, error) {
	logger.Title(context.Buildpack)

	config, err := libpak.NewConfigurationResolver(context.Buildpack, &logger)
	if err != nil {
		return nil, err
	}

	return &config, nil
}
