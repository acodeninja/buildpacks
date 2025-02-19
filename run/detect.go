package main

import (
	"errors"
	"github.com/buildpacks/libcnb"
	"github.com/paketo-buildpacks/libpak/bard"
	"os"
)

type Detect struct {
	Logger bard.Logger
}

func (d Detect) Detect(context libcnb.DetectContext) (libcnb.DetectResult, error) {
	d.Logger.Title(context.Buildpack)

	nonParticipatingResult := libcnb.DetectResult{
		Pass: false,
		Plans: []libcnb.BuildPlan{
			{
				Provides: []libcnb.BuildPlanProvide{
					{Name: "script-run"},
				},
			},
		},
	}

	scriptLocation, err := ResolveScriptLocation(context, d.Logger)
	if err != nil {
		return nonParticipatingResult, err
	}

	if _, err := os.Stat(scriptLocation); errors.Is(err, os.ErrNotExist) {
		return nonParticipatingResult, nil
	}

	return libcnb.DetectResult{
		Pass: true,
		Plans: []libcnb.BuildPlan{
			{
				Provides: []libcnb.BuildPlanProvide{
					{Name: "script-run"},
				},
				Requires: []libcnb.BuildPlanRequire{
					{
						Name: "script-run",
						Metadata: map[string]interface{}{
							"script-location": scriptLocation,
						},
					},
				},
			},
		},
	}, nil
}
