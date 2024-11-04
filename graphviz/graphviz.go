package main

import (
	"embed"
	"fmt"
	"github.com/acodeninja/buildpacks/common"
	"github.com/acodeninja/buildpacks/common/apt"
	"github.com/acodeninja/buildpacks/common/command"
	"github.com/acodeninja/buildpacks/common/fontconfig"
	"github.com/acodeninja/buildpacks/common/layers"
	"github.com/buildpacks/libcnb"
	"github.com/paketo-buildpacks/libpak"
	"github.com/paketo-buildpacks/libpak/bard"
	"os"
	"os/exec"
	"text/template"
)

//go:embed wrapper.sh
var embeddedFiles embed.FS

type WrapperScriptInput struct {
	Command            string
	FontConfigLocation string
	LibLocations       []string
	Path               string
	GraphvizBinDir     string
}

type GraphvizLayer struct {
	LayerName        string
	LayerContributor libpak.LayerContributor
	Logger           bard.Logger
	BuildContext     libcnb.BuildContext
}

func NewGraphvizLayer(context libcnb.BuildContext, logger bard.Logger) *GraphvizLayer {
	return &GraphvizLayer{
		LayerName:    "graphviz",
		Logger:       logger,
		BuildContext: context,
		LayerContributor: libpak.NewLayerContributor(
			"graphviz",
			map[string]interface{}{},
			libcnb.LayerTypes{
				Build:  true,
				Launch: true,
				Cache:  false,
			},
		),
	}
}

func (graphviz GraphvizLayer) Contribute(layer libcnb.Layer) (libcnb.Layer, error) {
	graphviz.LayerContributor.Logger = graphviz.Logger

	return graphviz.LayerContributor.Contribute(layer, func() (libcnb.Layer, error) {
		var err error
		var cmd *exec.Cmd

		err = apt.InstallAptPackages(
			layer,
			[]string{
				"graphviz",
				"libbrotli1",
				"glib2.0",
			},
			graphviz.Logger,
			true,
		)
		if err != nil {
			return libcnb.Layer{}, err
		}

		graphviz.Logger.Header("Configuring graphviz")
		err = fontconfig.ConfigPathRepoint(layer)
		if err != nil {
			return libcnb.Layer{}, err
		}

		err = WriteWrapperToBin(layer, graphviz.Logger, "dot")
		if err != nil {
			return libcnb.Layer{}, err
		}

		cmd = command.Make(common.IndentedWriterFactory(2, graphviz.Logger), fmt.Sprintf("%s/graphviz-bin/dot", layer.Path), "-c")
		cmd.Env = layers.GetLayerEnvironment(layer, graphviz.Logger)
		err = cmd.Run()
		if err != nil {
			return libcnb.Layer{}, err
		}

		graphviz.Logger.Header("Writing environment")
		layer.SharedEnvironment.Prepend("PATH", ":", fmt.Sprintf("%s/graphviz-bin", layer.Path))

		layer.LayerTypes.Build = true
		layer.LayerTypes.Launch = true
		layer.LayerTypes.Cache = true

		return layer, err
	})
}

func (graphviz GraphvizLayer) Name() string {
	return "graphviz"
}

func WriteWrapperToBin(layer libcnb.Layer, logger bard.Logger, binary string) error {
	logger.Bodyf("Writing wrapper script for %s", binary)

	script, err := embeddedFiles.ReadFile("wrapper.sh")
	if err != nil {
		return fmt.Errorf("unable to read embeded %s script\n%w", binary, err)
	}

	wrapperFileTemplate, err := template.New("wrapper").Parse(string(script))
	if err != nil {
		return fmt.Errorf("unable to parse %s script template\n%w", binary, err)
	}

	wrapperFileLocation := fmt.Sprintf("%s/graphviz-bin", layer.Path)

	err = os.MkdirAll(wrapperFileLocation, os.ModePerm)
	if err != nil {
		return fmt.Errorf("unable to create %s script location\n%w", binary, err)
	}

	wrapperFile, err := os.Create(fmt.Sprintf("%s/%s", wrapperFileLocation, binary))
	if err != nil {
		return fmt.Errorf("unable to create %s wrapper file\n%w", binary, err)
	}

	err = wrapperFileTemplate.Execute(wrapperFile, WrapperScriptInput{
		Command: binary,
		LibLocations: []string{
			fmt.Sprintf("%s/usr/lib/x86_64-linux-gnu", layer.Path),
			fmt.Sprintf("%s/lib/x86_64-linux-gnu", layer.Path),
		},
		Path:               fmt.Sprintf("%s/usr/bin/", layer.Path),
		FontConfigLocation: fmt.Sprintf("%s/etc/fonts", layer.Path),
		GraphvizBinDir:     fmt.Sprintf("%s/usr/lib/x86_64-linux-gnu/graphviz", layer.Path),
	})
	if err != nil {
		return fmt.Errorf("unable to create %s wrapper file\n%w", binary, err)
	}

	err = wrapperFile.Sync()
	if err != nil {
		return fmt.Errorf("unable to sync %s script\n%w", binary, err)
	}

	err = wrapperFile.Close()
	if err != nil {
		return fmt.Errorf("unable to close %s script\n%w", binary, err)
	}

	err = os.Chmod(fmt.Sprintf("%s/%s", wrapperFileLocation, binary), 0775)
	if err != nil {
		return fmt.Errorf("unable to chmod %s script\n%w", binary, err)
	}

	return nil
}
