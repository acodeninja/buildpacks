package main

import (
	"github.com/acodeninja/buildpacks/helpers"
	"github.com/buildpacks/libcnb"
	"github.com/paketo-buildpacks/libpak/bard"
)

type Detect struct {
	Logger bard.Logger
}

func (d Detect) Detect(context libcnb.DetectContext) (libcnb.DetectResult, error) {
	d.Logger.Title(context.Buildpack)

	playwrightVersion := ResolvePlaywrightVersion(d.Logger)

	pythonDependencyRequires := "site-packages"

	if helpers.DetectInFile("/workspace/poetry.lock", "\"playwright\"", d.Logger) {
		pythonDependencyRequires = "poetry-venv"
	}

	return libcnb.DetectResult{
		Pass: true,
		Plans: []libcnb.BuildPlan{
			{
				Provides: []libcnb.BuildPlanProvide{
					{Name: "playwright-python"},
				},
				Requires: []libcnb.BuildPlanRequire{
					{Name: "cpython"},
					{Name: pythonDependencyRequires},
					{
						Name: "playwright-python",
						Metadata: map[string]interface{}{
							"playwright-python-version": playwrightVersion,
						},
					},
				},
			},
		},
	}, nil
}
