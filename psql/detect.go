package main

import (
	"github.com/buildpacks/libcnb"
	"github.com/paketo-buildpacks/libpak/bard"
)

type Detect struct {
	Logger bard.Logger
}

func (d Detect) Detect(context libcnb.DetectContext) (libcnb.DetectResult, error) {
	d.Logger.Title(context.Buildpack)

	psqlVersion := ResolvePostgresClientVersion(d.Logger)

	return libcnb.DetectResult{
		Pass: true,
		Plans: []libcnb.BuildPlan{
			{
				Provides: []libcnb.BuildPlanProvide{
					{Name: "postgres-client"},
				},
				Requires: []libcnb.BuildPlanRequire{
					{
						Name: "postgres-client",
						Metadata: map[string]interface{}{
							"psql-version":      psqlVersion,
							"buildpack-version": context.Buildpack.Info.Version,
						},
					},
				},
			},
		},
	}, nil
}
