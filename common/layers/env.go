package layers

import (
	"fmt"
	"github.com/buildpacks/libcnb"
	"github.com/paketo-buildpacks/libpak/bard"
	"os"
	"strings"
)

type LayerEnv struct {
	Type      string
	Value     string
	Delimiter string
	Current   string
}

func GetLayerEnvironment(layer libcnb.Layer, logger bard.Logger) []string {
	layerEnv := map[string]LayerEnv{}
	osEnv := map[string]string{}

	logger.Headerf("  Loading environment variables from %s layer", layer.Name)

	for _, envVar := range os.Environ() {
		segments := strings.Split(envVar, "=")
		osEnv[segments[0]] = segments[1]
	}

	for name, value := range layer.SharedEnvironment {
		segments := strings.Split(name, ".")
		variableName := segments[0]
		_, exists := layerEnv[variableName]
		if !exists {
			layerEnv[variableName] = LayerEnv{
				Type:      "",
				Value:     "",
				Delimiter: ":",
				Current:   "",
			}
		}
		env := layerEnv[variableName]

		settingType := segments[1]

		if settingType == "delim" {
			env.Delimiter = value
		}

		if settingType == "prepend" {
			env.Value = value
			env.Type = "prepend"
			current, ok := osEnv[variableName]
			if ok {
				env.Current = current
			}
		}

		if settingType == "append" {
			env.Value = value
			env.Type = "append"
			current, ok := osEnv[variableName]
			if ok {
				env.Current = current
			}
		}

		layerEnv[variableName] = env
	}

	var output []string

	for name, value := range layerEnv {
		outputValue := value.Current
		if value.Type == "prepend" {
			outputValue = fmt.Sprintf("%s%s%s", value.Value, value.Delimiter, value.Current)
		}
		if value.Type == "append" {
			outputValue = fmt.Sprintf("%s%s%s", value.Current, value.Delimiter, value.Value)
		}

		outputValue = strings.TrimPrefix(outputValue, value.Delimiter)
		outputValue = strings.TrimSuffix(outputValue, value.Delimiter)

		output = append(output, fmt.Sprintf("%s=%s", name, outputValue))
	}

	for _, value := range os.Environ() {
		segments := strings.Split(value, "=")
		_, inOutput := layerEnv[segments[0]]
		if !inOutput {
			output = append(output, value)
		}
	}

	return output
}
