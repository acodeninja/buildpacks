package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/BurntSushi/toml"
	"github.com/acodeninja/buildpacks/common"
	"github.com/acodeninja/buildpacks/common/apt"
	"github.com/acodeninja/buildpacks/common/command"
	"github.com/buildpacks/libcnb"
	"github.com/paketo-buildpacks/libpak"
	"github.com/paketo-buildpacks/libpak/bard"
)

type PlaywrightLayer struct {
	LayerName          string
	TemporaryLayer     libcnb.Layer
	LayerContributor   libpak.LayerContributor
	Logger             bard.Logger
	PlaywrightVersion  string
	PlaywrightLanguage string
}

func NewPlaywrightLayer(playwrightVersion string, playwrightLanguage string, tempLayer libcnb.Layer, logger bard.Logger) *PlaywrightLayer {
	return &PlaywrightLayer{
		TemporaryLayer: tempLayer,
		LayerName:      fmt.Sprintf("playwright-%s", playwrightLanguage),
		LayerContributor: libpak.NewLayerContributor(
			fmt.Sprintf("playwright-%s", playwrightLanguage),
			map[string]interface{}{
				"playwright-version":  playwrightVersion,
				"playwright-language": playwrightLanguage,
			},
			libcnb.LayerTypes{
				Build:  true,
				Launch: true,
				Cache:  true,
			},
		),
		Logger:             logger,
		PlaywrightVersion:  playwrightVersion,
		PlaywrightLanguage: playwrightLanguage,
	}
}

func (playwright PlaywrightLayer) Contribute(layer libcnb.Layer) (libcnb.Layer, error) {
	playwright.LayerContributor.Logger = playwright.Logger

	return playwright.LayerContributor.Contribute(layer, func() (libcnb.Layer, error) {
		if layer.Metadata == nil {
			layer.Metadata = map[string]interface{}{}
		}
		layer.Metadata["playwright-version"] = playwright.PlaywrightVersion
		layer.Metadata["playwright-language"] = playwright.PlaywrightLanguage

		var err error

		switch playwright.PlaywrightLanguage {
		case "python":
			err = apt.InstallAptPackages(playwright.TemporaryLayer, []string{"python3-distutils", "python3-full", "python3-pip"}, []apt.AdditionalSource{}, playwright.Logger, true)

			var pythonBinary string
			pythonBinary, err = resolvePythonBinary(playwright.TemporaryLayer.Path)
			if err != nil {
				return layer, err
			}

			playwright.Logger.Headerf("Installing playwright version %s", playwright.PlaywrightVersion)

			installPlaywright := command.Make(
				common.IndentedWriterFactory(0, playwright.Logger),
				pythonBinary,
				"-m",
				"pip",
				"install",
				fmt.Sprintf("playwright==%s", playwright.PlaywrightVersion),
			)

			command.InjectLayerEnvironment(installPlaywright, playwright.TemporaryLayer.BuildEnvironment)

			err = installPlaywright.Run()

			if err != nil {
				return layer, err
			}

			playwright.Logger.Header("Installing playwright dependencies")
			playwrightInstall := command.Make(
				common.IndentedWriterFactory(0, playwright.Logger),
				pythonBinary,
				"-m",
				"playwright",
				"install",
			)
			playwrightInstall.Env = append(
				os.Environ(),
				fmt.Sprintf("PLAYWRIGHT_BROWSERS_PATH=%s", layer.Path),
			)
			command.InjectLayerEnvironment(playwrightInstall, playwright.TemporaryLayer.BuildEnvironment)
			err = playwrightInstall.Run()

			playwright.Logger.Header("Injecting Environment")
			layer.SharedEnvironment.Prependf("PLAYWRIGHT_BROWSERS_PATH", ":", layer.Path)
		default:
			return layer, fmt.Errorf("%s is not a supported playwright language", playwright.PlaywrightLanguage)
		}

		layer.LayerTypes.Build = true
		layer.LayerTypes.Launch = true
		layer.LayerTypes.Cache = true

		return layer, err
	})
}

func (playwright PlaywrightLayer) Name() string {
	return "playwright"
}

// resolvePythonBinary probes for a python interpreter under basePath/usr/bin,
// trying a few candidate paths in priority order and returning the first that
// exists on disk. The apt-installed interpreter is not always present at
// usr/bin/python3 — depending on the base image it may only exist as
// usr/bin/python or as a version-suffixed binary such as usr/bin/python3.10.
func resolvePythonBinary(basePath string) (string, error) {
	candidates := []string{
		"usr/bin/python3",
		"usr/bin/python",
	}

	for _, candidate := range candidates {
		full := filepath.Join(basePath, candidate)
		if _, err := os.Stat(full); err == nil {
			return full, nil
		}
	}

	// Fall back to a version-suffixed interpreter, e.g. usr/bin/python3.10.
	// The `python3.*` glob also matches non-interpreters like python3.10-config
	// and python3.10m, so filter down to a real `python3.N` binary.
	versioned := regexp.MustCompile(`python3\.[0-9]+$`)
	matches, _ := filepath.Glob(filepath.Join(basePath, "usr/bin/python3.*"))
	for _, match := range matches {
		if !versioned.MatchString(match) {
			continue
		}
		if _, err := os.Stat(match); err == nil {
			return match, nil
		}
	}

	if _, err := os.Stat(systemPythonFallback); err == nil {
		return systemPythonFallback, nil
	}

	return "", fmt.Errorf("no python binary found under %s/usr/bin", basePath)
}

// systemPythonFallback is the base-image interpreter used when the layer
// contains no python of its own. It is a variable rather than a literal so
// tests can point it at a controlled path.
var systemPythonFallback = "/usr/bin/python"

func ResolvePlaywrightVersion(logger bard.Logger) (string, string) {
	playwrightVersion := "1.43.0"
	playwrightLanguage := "python"
	resolved := false

	logger.Header("Resolving playwright version")

	// Find in requirements.txt
	requirementsPattern := regexp.MustCompile("^requirement.+\\.txt")
	requirementsPatternVersion := regexp.MustCompile("playwright[^0-9\n]+([0-9.]+)")

	files, err := os.ReadDir("/workspace")
	if err == nil {
		for _, file := range files {
			match := requirementsPattern.MatchString(file.Name())
			if match && !resolved {
				logger.Bodyf("Checking %s", file.Name())
				contents, err := os.ReadFile(file.Name())
				if err == nil {
					playwrightVersionMatches := requirementsPatternVersion.FindStringSubmatch(string(contents))
					if len(playwrightVersionMatches) == 2 {
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
	}

	return playwrightVersion, playwrightLanguage
}
