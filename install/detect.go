package main

import (
	"github.com/buildpacks/libcnb"
	"github.com/paketo-buildpacks/libpak/bard"
	"os"
	"path"
	"strings"
)

type Detect struct {
	Logger bard.Logger
}

const (
	PlanEntryAptInstall = "apt-install"
)

func (d Detect) Detect(context libcnb.DetectContext) (libcnb.DetectResult, error) {
	d.Logger.Title(context.Buildpack)

	provides := []libcnb.BuildPlanProvide{{Name: PlanEntryAptInstall}}
	var requires []libcnb.BuildPlanRequire

	content, err := os.ReadFile(path.Join(context.Application.Path, ".InstallPackages"))
	if err != nil {
		return libcnb.DetectResult{
			Pass: false,
		}, err
	}

	packages := strings.Split(string(content), "\n")

	if len(packages) > 0 {
		requires = append(requires, libcnb.BuildPlanRequire{
			Name: PlanEntryAptInstall,
			Metadata: map[string]interface{}{
				"packages": packages,
			},
		})
	}

	return libcnb.DetectResult{
		Pass: true,
		Plans: []libcnb.BuildPlan{
			{
				Provides: provides,
				Requires: requires,
			},
		},
	}, nil
}
