package main

import (
	"encoding/json"
	"fmt"
	"io"
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
			// If the base image already provides a python, use it and skip the
			// expensive temporary-layer apt install. Only the "base" builders,
			// which ship no system python, take the apt path below.
			pythonBinary := findPythonUnder(systemRoot)

			if pythonBinary == "" {
				playwright.Logger.Header("No system python found, installing python via APT")

				err = apt.InstallAptPackages(playwright.TemporaryLayer, []string{"python3-distutils", "python3-full", "python3-pip"}, []apt.AdditionalSource{}, playwright.Logger, true)
				if err != nil {
					return layer, err
				}

				pythonBinary, err = resolvePythonBinary(playwright.TemporaryLayer.Path)
				if err != nil {
					return layer, err
				}
			} else {
				playwright.Logger.Bodyf("Using system python at %s", pythonBinary)
			}

			// Some base images ship python without the pip module. When that is
			// the case, bootstrap pip into the temporary layer's user site with
			// get-pip.py; pythonUserBase then drives --user installs and the
			// PYTHONUSERBASE env so the bootstrapped pip and playwright resolve.
			pythonUserBase := ""
			if !pythonHasPip(pythonBinary, playwright.TemporaryLayer) {
				playwright.Logger.Header("pip module not available, bootstrapping pip with get-pip.py")

				pythonUserBase = playwright.TemporaryLayer.Path
				if err = bootstrapPip(pythonBinary, pythonUserBase, playwright.TemporaryLayer, playwright.Logger); err != nil {
					return layer, err
				}
			}

			playwright.Logger.Headerf("Installing playwright version %s", playwright.PlaywrightVersion)

			pipArgs := []string{"-m", "pip", "install", fmt.Sprintf("playwright==%s", playwright.PlaywrightVersion)}
			if pythonUserBase != "" {
				pipArgs = append(pipArgs, "--user")
			}

			installPlaywright := command.Make(
				common.IndentedWriterFactory(0, playwright.Logger),
				pythonBinary,
				pipArgs...,
			)
			if pythonUserBase != "" {
				installPlaywright.Env = append(installPlaywright.Env, fmt.Sprintf("PYTHONUSERBASE=%s", pythonUserBase))
			}

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
			if pythonUserBase != "" {
				playwrightInstall.Env = append(playwrightInstall.Env, fmt.Sprintf("PYTHONUSERBASE=%s", pythonUserBase))
			}
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

// resolvePythonBinary locates a python interpreter to drive pip and playwright.
// It prefers the layer's freshly apt-installed interpreter, then falls back to
// the base image's system python. On the paketo "full" builders python already
// ships in the base image and is never copied into the layer, so the system
// fallback is what makes those builds work — and jammy ships /usr/bin/python3
// (not an unversioned /usr/bin/python), so we must probe python3 and versioned
// names under the system root too.
func resolvePythonBinary(basePath string) (string, error) {
	for _, root := range []string{basePath, systemRoot} {
		if binary := findPythonUnder(root); binary != "" {
			return binary, nil
		}
	}

	return "", fmt.Errorf("no python binary found under %s/usr/bin or %s/usr/bin", basePath, systemRoot)
}

// findPythonUnder returns the first python interpreter under root/usr/bin,
// trying python3, then python, then a real version-suffixed python3.N, or ""
// when none is present.
func findPythonUnder(root string) string {
	for _, name := range []string{"usr/bin/python3", "usr/bin/python"} {
		full := filepath.Join(root, name)
		if info, err := os.Stat(full); err == nil && !info.IsDir() {
			return full
		}
	}

	// The `python3.*` glob also matches non-interpreters like python3.10-config
	// and python3.10m, plus the usr/lib/python3.10 directory, so keep only a
	// real python3.N file.
	versioned := regexp.MustCompile(`python3\.[0-9]+$`)
	matches, _ := filepath.Glob(filepath.Join(root, "usr/bin/python3.*"))
	for _, match := range matches {
		if !versioned.MatchString(match) {
			continue
		}
		if info, err := os.Stat(match); err == nil && !info.IsDir() {
			return match
		}
	}

	return ""
}

// systemRoot is the base image's filesystem root, probed when the layer has no
// python of its own. It is a variable so tests can point it at a temp dir.
var systemRoot = "/"

// pythonHasPip reports whether the given interpreter can import the pip module.
func pythonHasPip(pythonBinary string, tempLayer libcnb.Layer) bool {
	check := command.Make(io.Discard, pythonBinary, "-m", "pip", "--version")
	command.InjectLayerEnvironment(check, tempLayer.BuildEnvironment)
	return check.Run() == nil
}

// bootstrapPip downloads get-pip.py and installs pip into userBase's user site
// for the given interpreter, so a python that ships without pip can still drive
// pip and playwright (run with PYTHONUSERBASE=userBase).
func bootstrapPip(pythonBinary, userBase string, tempLayer libcnb.Layer, logger bard.Logger) error {
	// The temporary layer directory may not exist yet on the system-python fast
	// path (apt never ran), so DownloadFile's os.Create would otherwise fail.
	if err := os.MkdirAll(userBase, os.ModePerm); err != nil {
		return err
	}

	getPipPath := filepath.Join(userBase, "get-pip.py")
	if _, err := common.DownloadFile(getPipPath, "https://bootstrap.pypa.io/get-pip.py"); err != nil {
		return fmt.Errorf("unable to download get-pip.py\n%w", err)
	}

	getPip := command.Make(common.IndentedWriterFactory(0, logger), pythonBinary, getPipPath, "--user")
	getPip.Env = append(getPip.Env, fmt.Sprintf("PYTHONUSERBASE=%s", userBase))
	command.InjectLayerEnvironment(getPip, tempLayer.BuildEnvironment)
	return getPip.Run()
}

func ResolvePlaywrightVersion(logger bard.Logger) (string, string) {
	playwrightVersion := "1.62.0"
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
