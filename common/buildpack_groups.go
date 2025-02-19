package common

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/paketo-buildpacks/libpak/bard"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
)

type BuildpackGroups struct {
	Groups []struct {
		ID string `toml:"id"`
	} `toml:"group"`
}

func NewBuildpackGroups() (BuildpackGroups, error) {
	groups := BuildpackGroups{}
	file, err := os.ReadFile("/layers/group.toml")
	if err != nil {
		return groups, err
	}
	_, err = toml.Decode(string(file), &groups)
	return groups, err
}

type EnvironmentVariable struct {
	Key       string   `json:"key"`
	Value     []string `json:"value"`
	Delimiter string   `json:"delimiter"`
}

type EnvironmentVariables map[string]EnvironmentVariable

func (v EnvironmentVariables) GetForCommand() []string {
	var variables []string
	for key, value := range v {
		variables = append(variables, fmt.Sprintf("%s=%s", key, strings.Join(value.Value, value.Delimiter)))
	}
	return variables
}

func GetBuildpackGroupEnvironmentVariables(logger bard.Logger) (EnvironmentVariables, error) {
	variables := EnvironmentVariables{}

	matchEnvVarKey, _ := regexp.Compile("([A-Z]+)\\.[a-z]+$")

	groups, err := NewBuildpackGroups()
	if err != nil {
		return nil, err
	}

	logger.Header("Loading environment variables")

	logger.Body("loading system variables")
	for _, env := range os.Environ() {
		variable := strings.Split(env, "=")
		variables[variable[0]] = EnvironmentVariable{
			Key:   variable[0],
			Value: []string{variable[1]},
		}
	}

	for _, group := range groups.Groups {
		layerID := strings.Replace(group.ID, "/", "_", -1)
		matches, _ := filepath.Glob(fmt.Sprintf("/layers/%s/*/env", layerID))
		for _, match := range matches {
			f, _ := os.Stat(match)
			if f.IsDir() {
				logger.Body(fmt.Sprintf("loading %s", match))

				dir, _ := os.ReadDir(match)
				for _, entry := range dir {
					if !entry.IsDir() {
						envFilePath := fmt.Sprintf("%s/%s", match, entry.Name())
						envFile, _ := os.ReadFile(envFilePath)
						envFileContent := string(envFile)
						isDefault := strings.HasSuffix(entry.Name(), ".default")
						isPrepend := strings.HasSuffix(entry.Name(), ".prepend")
						isAppend := strings.HasSuffix(entry.Name(), ".append")
						isDelimiter := strings.HasSuffix(entry.Name(), ".delim")
						matchedKey := matchEnvVarKey.FindStringSubmatch(entry.Name())

						var key = ""
						if len(matchedKey) > 1 {
							key = matchedKey[1]
						}

						if len(key) > 0 {
							variable, ok := variables[key]
							if !ok {
								variable = EnvironmentVariable{
									Value: []string{},
								}
							}

							variable.Key = key

							if isDefault {
								variable.Value = []string{envFileContent}
							}

							if isPrepend {
								variable.Value = slices.Concat([]string{envFileContent}, variable.Value)
							}

							if isAppend {
								variable.Value = append(variable.Value, envFileContent)
							}

							if isDelimiter {
								variable.Delimiter = envFileContent
							}

							variables[key] = variable
						}
					}
				}
			}
		}
	}

	return variables, nil
}
