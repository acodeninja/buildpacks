package apt

import (
	"fmt"
	"github.com/acodeninja/buildpacks/common/command"
	"github.com/paketo-buildpacks/libpak/bard"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

type PackageInfo struct {
	Name     string
	Version  string
	Location string
}

func GetPackages(location string, logger bard.Logger) ([]string, error) {
	cmd := exec.Command(
		"ls",
		fmt.Sprintf("%s/", location),
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return []string{}, err
	}

	debFileMatcher := regexp.MustCompile("\\S+\\.deb")
	matches := debFileMatcher.FindAllString(string(output), -1)

	return matches, nil
}

func GetPackageInfo(packageName string, aptCacheDirectory string, logger bard.Logger) (PackageInfo, error) {
	packageInfo := PackageInfo{}

	cmd := exec.Command(
		"dpkg",
		"--info",
		fmt.Sprintf("%s/archives/%s", aptCacheDirectory, packageName),
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return packageInfo, err
	}

	logger.Body(string(output))

	packageNameMatcher := regexp.MustCompile("Package: (\\S+)")
	foundPackageName := packageNameMatcher.FindStringSubmatch(string(output))
	if len(foundPackageName) > 1 {
		packageInfo.Name = foundPackageName[1]
	}

	packageVersionMatcher := regexp.MustCompile("Version: (\\S+)")
	foundPackageVersion := packageVersionMatcher.FindStringSubmatch(string(output))
	if len(foundPackageVersion) > 1 {
		packageInfo.Version = foundPackageVersion[1]
	}

	return packageInfo, nil
}

func dpkgInstall(writer io.Writer, aptCacheDirectory, aptFolder string) error {
	aptArchives := fmt.Sprintf("%s/archives/", aptCacheDirectory)
	files, err := os.ReadDir(aptArchives)
	if err != nil {
		return err
	}

	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".deb") {
			_, err := writer.Write([]byte(fmt.Sprintf("Installing %s", file.Name())))
			if err != nil {
				return err
			}
			err = command.Run(
				writer,
				"dpkg",
				"-x",
				fmt.Sprintf("%s/%s", aptArchives, file.Name()),
				aptFolder,
			)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
