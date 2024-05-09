package main

import (
	"encoding/json"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/acodeninja/buildpacks/helpers"
	"github.com/buildpacks/libcnb"
	"github.com/paketo-buildpacks/libpak"
	"github.com/paketo-buildpacks/libpak/bard"
	"os"
	"regexp"
)

type PlaywrightLayer struct {
	LayerName        string
	AptLayer         libcnb.Layer
	LayerContributor libpak.LayerContributor
	Logger           bard.Logger
}

func NewPlaywrightLayer(aptLayer libcnb.Layer) *PlaywrightLayer {
	return &PlaywrightLayer{
		AptLayer:  aptLayer,
		LayerName: "playwright",
		LayerContributor: libpak.NewLayerContributor(
			"playwright",
			map[string]interface{}{},
			libcnb.LayerTypes{
				Build:  true,
				Launch: true,
				Cache:  false,
			},
		),
		Logger: bard.Logger{},
	}
}

func (playwright PlaywrightLayer) Contribute(layer libcnb.Layer) (libcnb.Layer, error) {
	return playwright.LayerContributor.Contribute(layer, func() (libcnb.Layer, error) {
		var err error

		err = helpers.InstallAptPackages(playwright.AptLayer, []string{"python3", "python3-pip"}, playwright.Logger)

		playwrightVersion := ResolvePlaywrightVersion(playwright.Logger)

		err = helpers.RunCommand(
			helpers.IndentedWriterFactory(4, playwright.Logger),
			"pip3",
			"install",
			fmt.Sprintf("playwright==%s", playwrightVersion),
		)
		if err != nil {
			return layer, err
		}

		playwrightInstall := helpers.GetCommand(
			helpers.IndentedWriterFactory(4, playwright.Logger),
			"playwright",
			"install",
		)
		playwrightInstall.Env = append(
			os.Environ(),
			fmt.Sprintf("PLAYWRIGHT_BROWSERS_PATH=%s", layer.Path),
		)
		err = playwrightInstall.Run()

		playwright.Logger.Header("Injecting Environment")
		layer.SharedEnvironment.Prependf("PLAYWRIGHT_BROWSERS_PATH", ":", layer.Path)
		layer.LayerTypes.Build = true
		layer.LayerTypes.Launch = true
		layer.LayerTypes.Cache = true

		return layer, err
	})
}

func (playwright PlaywrightLayer) Name() string {
	return "playwright"
}

func ResolvePlaywrightVersion(logger bard.Logger) string {
	playwrightVersion := "1.43.0"
	resolved := false

	logger.Header("Resolving playwright version")

	// Find in requirements.txt
	requirementsPattern := regexp.MustCompile("^requirement.+\\.txt")
	requirementsPatternVersion := regexp.MustCompile("playwright[^0-9]+([0-9.]+)")

	files, err := os.ReadDir("/workspace")
	if err == nil {
		for _, file := range files {
			match := requirementsPattern.MatchString(file.Name())
			if match {
				logger.Bodyf("Checking %s", file.Name())
				contents, err := os.ReadFile(file.Name())
				if err == nil {
					playwrightVersionMatches := requirementsPatternVersion.FindStringSubmatch(string(contents))
					playwrightVersion = playwrightVersionMatches[len(playwrightVersionMatches)-1]
					resolved = true
					logger.Bodyf("Found playwright version %s in %s", playwrightVersion, file.Name())
				}
			}
		}
	}

	// Find in Pipfile
	if !resolved {
		var pipFile map[string]interface{}
		_, err = os.Stat("/workspace/Pipfile.lock")
		if err == nil {
			pipFileContents, err := os.ReadFile("/workspace/Pipfile.lock")

			logger.Body("Checking /workspace/Pipfile.lock")

			if err == nil {
				err = json.Unmarshal(pipFileContents, &pipFile)
				if err == nil {
					foundPipfileVersion := pipFile["default"].(map[string]interface{})["playwright"].(map[string]interface{})["version"].(string)
					versionPattern := regexp.MustCompile("([0-9.]+)")

					matches := versionPattern.FindStringSubmatch(foundPipfileVersion)
					playwrightVersion = matches[1]
					resolved = true
					logger.Bodyf("Found playwright version %s in /workspace/Pipfile.lock", playwrightVersion)
				}
			}
		}
	}

	// Find in Poetry.lock
	if !resolved {
		var poetryFile map[string]interface{}
		_, err = os.Stat("/workspace/poetry.lock")
		if err == nil {
			poetryFileContents, err := os.ReadFile("/workspace/poetry.lock")

			logger.Body("Checking /workspace/poetry.lock")

			if err == nil {
				err = toml.Unmarshal(poetryFileContents, &poetryFile)
				if err == nil {
					for _, p := range poetryFile["package"].([]map[string]interface{}) {
						if p["name"] == "playwright" {
							versionPattern := regexp.MustCompile("([0-9.]+)")
							matches := versionPattern.FindStringSubmatch(p["version"].(string))
							playwrightVersion = matches[1]
							resolved = true
							logger.Bodyf("Found playwright version %s in /workspace/poetry.lock", playwrightVersion)
						}
					}
				}
			}
		}
	}

	return playwrightVersion
}
