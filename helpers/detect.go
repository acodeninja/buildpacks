package helpers

import (
	"github.com/paketo-buildpacks/libpak/bard"
	"os"
	"regexp"
)

func DetectInFile(file string, pattern string, logger bard.Logger) bool {
	logger.Headerf("Checking for '%s' in file %s", pattern, file)

	fileContents, err := os.ReadFile(file)
	if err != nil {
		logger.Bodyf("File %s not found in workspace", file)
		return false
	}

	matcher, err := regexp.Compile(pattern)

	found := matcher.FindStringSubmatch(string(fileContents))
	logger.Body(found)
	if len(found) > 0 {
		logger.Bodyf("Found %s in file %s", pattern, file)
		return true
	} else {
		logger.Bodyf("%s not in file %s", pattern, file)
		return false
	}
}
