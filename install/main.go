package main

import (
	"github.com/paketo-buildpacks/libpak"
	"github.com/paketo-buildpacks/libpak/bard"
	"os"
)

func main() {
	libpak.Main(
		Detect{
			Logger: bard.NewLogger(os.Stdout),
		},
		Build{
			Logger: bard.NewLogger(os.Stdout),
		})
}
