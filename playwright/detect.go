package main

import (
	"github.com/acodeninja/buildpacks/helpers"
	"github.com/buildpacks/libcnb"
	"github.com/paketo-buildpacks/libpak/bard"
	"os"
	"regexp"
)

type Detect struct {
	Logger bard.Logger
}

func (d Detect) Detect(context libcnb.DetectContext) (libcnb.DetectResult, error) {
	var err error
	d.Logger.Title(context.Buildpack)

	foundPlaywright := false

	checkers := []bool{
		helpers.DetectInFile("/workspace/Pipfile", "playwright", d.Logger),
		helpers.DetectInFile("/workspace/requirements.txt", "playwright", d.Logger),
		helpers.DetectInFile("/workspace/poetry.lock", "\"playwright\"", d.Logger),
	}

	requirementsPattern, err := regexp.Compile("^requirement.+\\.txt")
	if err != nil {
		return libcnb.DetectResult{
			Pass: false,
		}, err
	}

	files, err := os.ReadDir("/workspace")
	if err != nil {
		return libcnb.DetectResult{Pass: false}, err
	}
	for _, file := range files {
		match := requirementsPattern.MatchString(file.Name())
		if match {
			checkers = append(checkers, helpers.DetectInFile(file.Name(), "^playwright", d.Logger))
		}
	}

	for _, check := range checkers {
		if check {
			foundPlaywright = true
		}
	}

	if !foundPlaywright {
		return libcnb.DetectResult{
			Pass: false,
		}, err
	}

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
					{Name: "playwright-python"},
				},
			},
		},
	}, nil
}
