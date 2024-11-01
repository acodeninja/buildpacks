package common

import (
	"fmt"
	"github.com/buildpacks/libcnb"
	"github.com/paketo-buildpacks/libpak/bard"
	"log"
	"regexp"
	"sort"
	"strings"
)

type IndentedWriter struct {
	Indent int
	Logger bard.Logger
}

func (iw IndentedWriter) Write(p []byte) (int, error) {
	indent := strings.Repeat(" ", iw.Indent)
	input := fmt.Sprintf("%s", p)

	for _, line := range strings.Split(input, "\n") {
		if line != "" {
			regex := regexp.MustCompile(`^`)
			output := regex.ReplaceAllString(line, indent)

			iw.Logger.Body(strings.ReplaceAll(output, "\n", ""))
		}
	}

	return len(p), nil
}

func IndentedWriterFactory(indent int, logger bard.Logger) IndentedWriter {
	return IndentedWriter{Indent: indent, Logger: logger}
}

func SummariseContributions(layer libcnb.Layer) {
	log.Println("")
	log.Println("  Build summary")
	log.Printf("    Contributed layer %s", layer.Name)
	log.Println("")

	var buildEnvironment []string
	for name, value := range layer.BuildEnvironment {
		if strings.HasSuffix(name, ".delim") {
			continue
		}

		buildEnvironment = append(buildEnvironment, fmt.Sprintf("      %s -> \"%s\"", strings.Split(name, ".")[0], value))
	}

	var launchEnvironment []string
	for name, value := range layer.LaunchEnvironment {
		if strings.HasSuffix(name, ".delim") {
			continue
		}

		launchEnvironment = append(launchEnvironment, fmt.Sprintf("      %s -> \"%s\"", strings.Split(name, ".")[0], value))
	}

	for name, value := range layer.SharedEnvironment {
		if strings.HasSuffix(name, ".delim") {
			continue
		}

		launchEnvironment = append(launchEnvironment, fmt.Sprintf("      %s -> \"%s\"", strings.Split(name, ".")[0], value))
		buildEnvironment = append(buildEnvironment, fmt.Sprintf("      %s -> \"%s\"", strings.Split(name, ".")[0], value))
	}
	sort.Strings(launchEnvironment)
	sort.Strings(buildEnvironment)

	if len(buildEnvironment) > 0 {
		log.Println("")
		log.Println("    Configured build environment")
		for _, line := range buildEnvironment {
			log.Println(line)
		}
	}

	if len(buildEnvironment) > 0 {
		log.Println("")
		log.Println("    Configured launch environment")
		for _, line := range launchEnvironment {
			log.Println(line)
		}
	}

	log.Println("")
}
