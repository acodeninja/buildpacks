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

	playwrightVersion, playwrightLanguage := ResolvePlaywrightVersion(d.Logger)

	return libcnb.DetectResult{
		Pass: true,
		Plans: []libcnb.BuildPlan{
			{
				Provides: []libcnb.BuildPlanProvide{
					{Name: "playwright"},
				},
				Requires: []libcnb.BuildPlanRequire{
					{
						Name: "playwright",
						Metadata: map[string]interface{}{
							"playwright-version":  playwrightVersion,
							"playwright-language": playwrightLanguage,
						},
					},
				},
			},
		},
	}, nil
}
