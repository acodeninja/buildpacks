package main

import (
	"embed"
	"fmt"
	"github.com/acodeninja/buildpacks/common/apt"
	"github.com/buildpacks/libcnb"
	"github.com/paketo-buildpacks/libpak"
	"github.com/paketo-buildpacks/libpak/bard"
	"os"
	"regexp"
	"text/template"
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
			err = WriteWrapperToBin(layer, psql.Logger, command)
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

func WriteWrapperToBin(layer libcnb.Layer, logger bard.Logger, pgBinary string) error {
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
		PostgresClientPath: fmt.Sprintf("%s/usr/lib/postgresql/14/bin", layer.Path),
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

func ResolvePostgresClientVersion(logger bard.Logger) string {
	psqlVersion := "14"
	ubuntuVersion := ResolveUbuntuVersion(logger)

	logger.Header("Resolving psql version")

	switch ubuntuVersion {
	case "focal":
		psqlVersion = "12"
	case "jammy":
		psqlVersion = "14"
	case "mantic":
		psqlVersion = "15"
	case "nobel":
		psqlVersion = "16"
	default:
		psqlVersion = ""
	}

	logger.Bodyf("Found version %s", psqlVersion)

	return psqlVersion
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
