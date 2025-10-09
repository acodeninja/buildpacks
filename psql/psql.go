package main

import (
	"embed"
	"fmt"
	"os"
	"path"
	"regexp"
	"strings"
	"text/template"

	"github.com/acodeninja/buildpacks/common/apt"
	"github.com/buildpacks/libcnb"
	"github.com/paketo-buildpacks/libpak"
	"github.com/paketo-buildpacks/libpak/bard"
)

//go:embed wrapper.sh
var embeddedFiles embed.FS

type WrapperScriptInput struct {
	LibLocations       []string
	PerlLibLocation    string
	PostgresClientPath string
	PostgresCommand    string
}

type PostgresClientLayer struct {
	LayerName              string
	LayerContributor       libpak.LayerContributor
	Logger                 bard.Logger
	PostgresClientVersion  string
	PostgresClientLanguage string
	BuildpackVersion       string
}

func NewPostgresClientLayer(psqlVersion string, buildpackVersion string, logger bard.Logger) *PostgresClientLayer {
	return &PostgresClientLayer{
		LayerName: fmt.Sprintf("psql-%s", psqlVersion),
		LayerContributor: libpak.NewLayerContributor(
			fmt.Sprintf("psql-%s", psqlVersion),
			map[string]interface{}{
				"psql-version":      psqlVersion,
				"buildpack-version": buildpackVersion,
			},
			libcnb.LayerTypes{
				Build:  true,
				Launch: true,
				Cache:  true,
			},
		),
		Logger:                logger,
		PostgresClientVersion: psqlVersion,
		BuildpackVersion:      buildpackVersion,
	}
}

func (psql PostgresClientLayer) Contribute(layer libcnb.Layer) (libcnb.Layer, error) {
	psql.LayerContributor.Logger = psql.Logger

	return psql.LayerContributor.Contribute(layer, func() (libcnb.Layer, error) {
		if layer.Metadata == nil {
			layer.Metadata = map[string]interface{}{}
		}
		layer.Metadata["psql-version"] = psql.PostgresClientVersion
		layer.Metadata["buildpack-version"] = psql.BuildpackVersion

		var err error

		psql.Logger.Headerf("Setting up psql version %s", psql.PostgresClientVersion)

		err = apt.InstallAptPackages(
			layer,
			[]string{
				fmt.Sprintf("postgresql-client-%s", psql.PostgresClientVersion),
				fmt.Sprintf("postgresql-contrib-%s", psql.PostgresClientVersion),
				"libsasl2-2",
				"libldap-2.5-0",
				"libpq5",
				"libpq-dev",
			},
			[]apt.AdditionalSource{
				{
					"https://www.postgresql.org/media/keys/ACCC4CF8.asc",
					"https://apt.postgresql.org/pub/repos/apt",
					fmt.Sprintf("%s-pgdg", ResolveUbuntuVersion(psql.Logger)),
					"main",
				},
			},
			psql.Logger,
			false,
		)
		if err != nil {
			return libcnb.Layer{}, fmt.Errorf("unable to install postgresql-client\n%w", err)
		}

		psql.Logger.Header("Installing command wrappers")
		commandsToWrap := []string{
			"pg_amcheck",
			"pgbench",
			"pg_config",
			"pg_dump",
			"pg_dumpall",
			"pg_isready",
			"pg_receivewal",
			"pg_restore",
			"psql",
		}

		for _, command := range commandsToWrap {
			err = WriteWrapperToBin(layer, psql.PostgresClientVersion, psql.Logger, command)
			if err != nil {
				return libcnb.Layer{}, err
			}
		}

		psql.Logger.Header("Writing environment")
		layer.SharedEnvironment.Prepend("PATH", ":", fmt.Sprintf("%s/psql-bin", layer.Path))

		layer.LayerTypes.Build = true
		layer.LayerTypes.Launch = true
		layer.LayerTypes.Cache = true

		return layer, err
	})
}

func (psql PostgresClientLayer) Name() string {
	return "psql"
}

func WriteWrapperToBin(layer libcnb.Layer, postgresVersion string, logger bard.Logger, pgBinary string) error {
	logger.Bodyf("Writing wrapper script for %s", pgBinary)

	script, err := embeddedFiles.ReadFile("wrapper.sh")
	if err != nil {
		return fmt.Errorf("unable to read embeded %s script\n%w", pgBinary, err)
	}

	wrapperFileTemplate, err := template.New("wrapper").Parse(string(script))
	if err != nil {
		return fmt.Errorf("unable to parse %s script template\n%w", pgBinary, err)
	}

	wrapperFileLocation := fmt.Sprintf("%s/psql-bin", layer.Path)

	err = os.MkdirAll(wrapperFileLocation, os.ModePerm)
	if err != nil {
		return fmt.Errorf("unable to create %s script location\n%w", pgBinary, err)
	}

	wrapperFile, err := os.Create(fmt.Sprintf("%s/%s", wrapperFileLocation, pgBinary))
	if err != nil {
		return fmt.Errorf("unable to create %s wrapper file\n%w", pgBinary, err)
	}

	err = wrapperFileTemplate.Execute(wrapperFile, WrapperScriptInput{
		LibLocations: []string{
			fmt.Sprintf("%s/usr/lib/x86_64-linux-gnu/sasl2", layer.Path),
			fmt.Sprintf("%s/usr/lib/x86_64-linux-gnu", layer.Path),
			fmt.Sprintf("%s/lib/x86_64-linux-gnu", layer.Path),
		},
		PerlLibLocation:    fmt.Sprintf("%s/usr/share/perl5", layer.Path),
		PostgresClientPath: fmt.Sprintf("%s/usr/lib/postgresql/%s/bin", layer.Path, postgresVersion),
		PostgresCommand:    pgBinary,
	})
	if err != nil {
		return fmt.Errorf("unable to create %s wrapper file\n%w", pgBinary, err)
	}

	err = wrapperFile.Sync()
	if err != nil {
		return fmt.Errorf("unable to sync %s script\n%w", pgBinary, err)
	}

	err = wrapperFile.Close()
	if err != nil {
		return fmt.Errorf("unable to close %s script\n%w", pgBinary, err)
	}

	err = os.Chmod(fmt.Sprintf("%s/%s", wrapperFileLocation, pgBinary), 0775)
	if err != nil {
		return fmt.Errorf("unable to chmod %s script\n%w", pgBinary, err)
	}

	return nil
}

func ResolvePostgresClientVersion(context libcnb.DetectContext, logger bard.Logger) string {
	logger.Header("Resolving psql client version")

	content, err := os.ReadFile(path.Join(context.Application.Path, ".psql-version"))
	if err == nil {
		version := strings.TrimSpace(string(content))
		logger.Bodyf("found version %s in .psql-version file", version)
		return version
	}

	ubuntuVersion := ResolveUbuntuVersion(logger)

	switch ubuntuVersion {
	case "focal":
		logger.Body("found version 12 in ubuntu focal")
		return "12"
	case "jammy":
		logger.Body("found version 14 in ubuntu jammy")
		return "14"
	case "mantic":
		logger.Body("found version 15 in ubuntu mantic")
		return "15"
	case "noble":
		logger.Body("found version 16 in ubuntu noble")
		return "16"
	}

	logger.Body("no version found, defaulting to 14")
	return "14"
}

func ResolveUbuntuVersion(logger bard.Logger) string {
	ubuntuVersion := ""

	logger.Header("Resolving ubuntu version")

	ubuntuVersionMatcher := regexp.MustCompile("DISTRIB_CODENAME=(?P<version>\\S+)")

	contents, err := os.ReadFile("/etc/lsb-release")
	if err != nil {
		panic(err)
	}
	ubuntuVersionMatches := ubuntuVersionMatcher.FindStringSubmatch(string(contents))
	if len(ubuntuVersionMatches) == 2 {
		ubuntuVersion = ubuntuVersionMatches[len(ubuntuVersionMatches)-1]
	}

	logger.Bodyf("Found version %s", ubuntuVersion)

	return ubuntuVersion
}
