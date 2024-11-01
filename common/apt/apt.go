package apt

import (
	"errors"
	"fmt"
	"github.com/acodeninja/buildpacks/common"
	"github.com/acodeninja/buildpacks/common/command"
	"github.com/buildpacks/libcnb"
	"github.com/paketo-buildpacks/libpak/bard"
	"io"
	"os"
	"path/filepath"
	"slices"
)

func InstallAptPackages(layer libcnb.Layer, packageList []string, logger bard.Logger, buildOnly bool) error {
	var err error

	logger.Headerf("Installing APT packages in %s layer", layer.Name)

	aptFolder := layer.Path
	aptCacheDirectory := filepath.Join(aptFolder, "cache")
	aptStateDirectory := filepath.Join(aptFolder, "state")
	aptSourcesDirectory := filepath.Join(aptFolder, "sources")
	aptArchiveDirectory := filepath.Join(aptFolder, "archives")
	aptListsDirectory := filepath.Join(aptFolder, "lists")

	logger.Headerf("  Creating APT directories")

	_, err = os.Stat(aptFolder)

	if errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir(aptFolder, os.ModePerm)
		if err != nil {
			return err
		}
		logger.Body("  Created", aptFolder)
	}

	aptDirectories := []string{
		aptCacheDirectory,
		aptStateDirectory,
		aptSourcesDirectory,
		aptArchiveDirectory,
		aptListsDirectory,
	}

	for _, directory := range aptDirectories {
		err := os.Mkdir(directory, os.ModePerm)
		if err != nil {
			return err
		}
		logger.Body("  Created", directory)
	}

	err = common.CopyFile("/etc/apt/sources.list", fmt.Sprintf("%s/sources.list", aptSourcesDirectory))
	if err != nil {
		return err
	}

	logger.Header("  Updating APT sources")
	err = aptUpdate(
		common.IndentedWriterFactory(2, logger),
		aptCacheDirectory,
		aptStateDirectory,
		aptSourcesDirectory,
	)
	if err != nil {
		return err
	}

	logger.Header("  Downloading APT packages")
	err = aptDownload(
		common.IndentedWriterFactory(2, logger),
		aptCacheDirectory,
		aptStateDirectory,
		aptSourcesDirectory,
		packageList,
	)
	if err != nil {
		return err
	}

	logger.Header("  Installing APT packages")
	err = dpkgInstall(common.IndentedWriterFactory(2, logger), aptCacheDirectory, aptFolder)

	if buildOnly {
		layer.BuildEnvironment.Prependf("PATH", ":", "%s/usr/bin:%s/bin", aptFolder, aptFolder)

		libPath := fmt.Sprintf("%s/lib/x86_64-linux-gnu:%s/lib/i386-linux-gnu:%s/lib:%s/usr/lib/x86_64-linux-gnu:%s/usr/lib/i386-linux-gnu:%s/usr/lib", aptFolder, aptFolder, aptFolder, aptFolder, aptFolder, aptFolder)
		layer.BuildEnvironment.Prepend("LD_LIBRARY_PATH", ":", libPath)
		layer.BuildEnvironment.Prepend("LIBRARY_PATH", ":", libPath)

		includePath := fmt.Sprintf("%s/usr/include:%s/usr/include/x86_64-linux-gnu", aptFolder, aptFolder)
		layer.BuildEnvironment.Prepend("INCLUDE_PATH", ":", includePath)
		layer.BuildEnvironment.Prepend("CPATH", ":", includePath)
		layer.BuildEnvironment.Prepend("CPPPATH", ":", includePath)
	} else {
		layer.SharedEnvironment.Prependf("PATH", ":", "%s/usr/bin:%s/bin", aptFolder, aptFolder)

		libPath := fmt.Sprintf("%s/lib/x86_64-linux-gnu:%s/lib/i386-linux-gnu:%s/lib:%s/usr/lib/x86_64-linux-gnu:%s/usr/lib/i386-linux-gnu:%s/usr/lib", aptFolder, aptFolder, aptFolder, aptFolder, aptFolder, aptFolder)
		layer.SharedEnvironment.Prepend("LD_LIBRARY_PATH", ":", libPath)
		layer.SharedEnvironment.Prepend("LIBRARY_PATH", ":", libPath)

		includePath := fmt.Sprintf("%s/usr/include:%s/usr/include/x86_64-linux-gnu", aptFolder, aptFolder)
		layer.SharedEnvironment.Prepend("INCLUDE_PATH", ":", includePath)
		layer.SharedEnvironment.Prepend("CPATH", ":", includePath)
		layer.SharedEnvironment.Prepend("CPPPATH", ":", includePath)
	}

	return nil
}

func aptUpdate(writer io.Writer, aptCacheDirectory, aptStateDirectory, aptSourcesDirectory string) error {
	return command.Run(
		writer,
		"apt-get",
		"-o", "debug::nolocking=true",
		"-o", "dir::etc::sourceparts=/dev/null",
		"-o", fmt.Sprintf("dir::cache=%s", aptCacheDirectory),
		"-o", fmt.Sprintf("dir::state=%s", aptStateDirectory),
		"-o", fmt.Sprintf("dir::etc::sourcelist=%s/sources.list", aptSourcesDirectory),
		"update",
	)
}

func aptDownload(writer io.Writer, aptCacheDirectory, aptStateDirectory, aptSourcesDirectory string, packages []string) error {
	args := []string{
		"-o", "debug::nolocking=true",
		"-o", "dir::etc::sourceparts=/dev/null",
		"-o", fmt.Sprintf("dir::cache=%s", aptCacheDirectory),
		"-o", fmt.Sprintf("dir::state=%s", aptStateDirectory),
		"-o", fmt.Sprintf("dir::etc::sourcelist=%s/sources.list", aptSourcesDirectory),
		"-y",
		"--allow-downgrades",
		"--allow-remove-essential",
		"--allow-change-held-packages",
		"-d",
		"install",
		"--reinstall",
	}

	return command.Run(writer, "apt-get", slices.Concat(args, packages)...)
}
