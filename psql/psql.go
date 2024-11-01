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

//go:embed psql.sh
var embeddedFiles embed.FS

type PSQLScriptData struct {
	LibLocations       []string
	PerlLibLocation    string
	PostgresClientPath string
}

type PostgresClientLayer struct {
	LayerName              string
	LayerContributor       libpak.LayerContributor
	Logger                 bard.Logger
	PostgresClientVersion  string
	PostgresClientLanguage string
}

func NewPostgresClientLayer(psqlVersion string, logger bard.Logger) *PostgresClientLayer {
	return &PostgresClientLayer{
		LayerName: fmt.Sprintf("psql-%s", psqlVersion),
		LayerContributor: libpak.NewLayerContributor(
			fmt.Sprintf("psql-%s", psqlVersion),
			map[string]interface{}{
				"psql-version": psqlVersion,
			},
			libcnb.LayerTypes{
				Build:  true,
				Launch: true,
				Cache:  true,
			},
		),
		Logger:                logger,
		PostgresClientVersion: psqlVersion,
	}
}

func (psql PostgresClientLayer) Contribute(layer libcnb.Layer) (libcnb.Layer, error) {
	psql.LayerContributor.Logger = psql.Logger

	return psql.LayerContributor.Contribute(layer, func() (libcnb.Layer, error) {
		if layer.Metadata == nil {
			layer.Metadata = map[string]interface{}{}
		}
		layer.Metadata["psql-version"] = psql.PostgresClientVersion

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
			},
			psql.Logger,
			true,
		)
		if err != nil {
			return libcnb.Layer{}, fmt.Errorf("unable to install postgresql-client\n%w", err)
		}

		psql.Logger.Header("Installing bash wrapper")

		psqlScript, err := embeddedFiles.ReadFile("psql.sh")
		if err != nil {
			return libcnb.Layer{}, fmt.Errorf("unable to read embeded psql script\n%w", err)
		}

		psqlFileTemplate, err := template.New("psql").Parse(string(psqlScript))
		if err != nil {
			return libcnb.Layer{}, fmt.Errorf("unable to parse psql script template\n%w", err)
		}

		psqlFileLocation := fmt.Sprintf("%s/psql-bin", layer.Path)

		err = os.MkdirAll(psqlFileLocation, os.ModePerm)
		if err != nil {
			return libcnb.Layer{}, fmt.Errorf("unable to create psql script location\n%w", err)
		}

		psqlFile, err := os.Create(fmt.Sprintf("%s/psql", psqlFileLocation))
		if err != nil {
			return libcnb.Layer{}, fmt.Errorf("unable to create psql wrapper file\n%w", err)
		}

		err = psqlFileTemplate.Execute(psqlFile, PSQLScriptData{
			LibLocations: []string{
				fmt.Sprintf("%s/usr/lib/x86_64-linux-gnu/sasl2", layer.Path),
				fmt.Sprintf("%s/usr/lib/x86_64-linux-gnu", layer.Path),
				fmt.Sprintf("%s/lib/x86_64-linux-gnu", layer.Path),
			},
			PerlLibLocation:    fmt.Sprintf("%s/usr/share/perl5", layer.Path),
			PostgresClientPath: fmt.Sprintf("%s/usr/lib/postgresql/14/bin", layer.Path),
		})
		if err != nil {
			return libcnb.Layer{}, fmt.Errorf("unable to create psql wrapper file\n%w", err)
		}

		err = psqlFile.Sync()
		if err != nil {
			return libcnb.Layer{}, fmt.Errorf("unable to sync psql script\n%w", err)
		}

		err = psqlFile.Close()
		if err != nil {
			return libcnb.Layer{}, fmt.Errorf("unable to close psql script\n%w", err)
		}

		err = os.Chmod(fmt.Sprintf("%s/psql", psqlFileLocation), 0775)
		if err != nil {
			return libcnb.Layer{}, fmt.Errorf("unable to chmod psql script\n%w", err)
		}

		layer.SharedEnvironment.Prepend("PATH", ":", psqlFileLocation)

		layer.LayerTypes.Build = true
		layer.LayerTypes.Launch = true
		layer.LayerTypes.Cache = true

		return layer, err
	})
}

func (psql PostgresClientLayer) Name() string {
	return "psql"
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
