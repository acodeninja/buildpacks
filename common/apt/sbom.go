package apt

import (
	"context"
	"fmt"
	"github.com/anchore/syft/syft/artifact"
	"github.com/anchore/syft/syft/file"
	"github.com/anchore/syft/syft/pkg"
	"github.com/buildpacks/libcnb"
	"io"
	"path"
	"path/filepath"
)

var _ pkg.Cataloger = (*aptConfigurationCataloger)(nil)

type aptConfigurationCataloger struct {
	Layer libcnb.Layer
}

func NewAptConfigurationCataloger(layer libcnb.Layer) pkg.Cataloger {
	return aptConfigurationCataloger{
		Layer: layer,
	}
}

func (m aptConfigurationCataloger) Name() string {
	return "apt-packages-cataloger"
}

func (m aptConfigurationCataloger) Catalog(_ context.Context, resolver file.Resolver) ([]pkg.Package, []artifact.Relationship, error) {
	//packages := GetPackages(filepath.Join(m.Layer.Path, "cache"))
	//

	fmt.Println(filepath.Join(m.Layer.Path, "cache", "archive"))
	//version, versionLocations, err := getVersion(resolver)
	//if err != nil {
	//	return nil, nil, fmt.Errorf("unable to get apt version: %w", err)
	//}
	//if len(versionLocations) == 0 {
	//	// this doesn't mean we should stop cataloging, just that we don't have a version to use, thus no package to raise up
	//	return nil, nil, nil
	//}
	//
	//metadata, metadataLocations, err := newAptConfiguration(resolver)
	//if err != nil {
	//	return nil, nil, err
	//}
	//
	//var locations []file.Location
	//locations = append(locations, versionLocations...)
	//locations = append(locations, metadataLocations...)
	//
	//p := newPackage(name, version, *metadata, locations...)

	return []pkg.Package{}, nil, nil
}

func newPackage(name string, version string, metadata AptConfiguration, locations ...file.Location) pkg.Package {
	return pkg.Package{
		Name:      name,
		Version:   version,
		Locations: file.NewLocationSet(locations...),
		Type:      pkg.Type("apt"),
		Metadata:  metadata,
	}
}

func newAptConfiguration(resolver file.Resolver) (*AptConfiguration, []file.Location, error) {
	var locations []file.Location

	//keys, keyLocations, err := getAPKKeys(resolver)
	//if err != nil {
	//	return nil, nil, err
	//}
	//
	//locations = append(locations, keyLocations...)

	return &AptConfiguration{}, locations, nil
}

func getVersion(resolver file.Resolver) (string, []file.Location, error) {
	locations, err := resolver.FilesByPath("/etc/apt-release")
	if err != nil {
		return "", nil, fmt.Errorf("unable to get apt version: %w", err)
	}
	if len(locations) == 0 {
		return "", nil, nil
	}

	reader, err := resolver.FileContentsByLocation(locations[0])
	if err != nil {
		return "", nil, fmt.Errorf("unable to read apt version: %w", err)
	}

	version, err := io.ReadAll(reader)
	if err != nil {
		return "", nil, fmt.Errorf("unable to read apt version: %w", err)
	}

	return string(version), locations, nil
}

func getAPKKeys(resolver file.Resolver) (map[string]string, []file.Location, error) {
	// name-to-content values
	keyContent := make(map[string]string)

	locations, err := resolver.FilesByGlob("/etc/apk/keys/*.rsa.pub")
	if err != nil {
		return nil, nil, fmt.Errorf("unable to get apk keys: %w", err)
	}
	for _, location := range locations {
		basename := path.Base(location.RealPath)
		reader, err := resolver.FileContentsByLocation(location)
		content, err := io.ReadAll(reader)
		if err != nil {
			return nil, nil, fmt.Errorf("unable to read apk key content at %s: %w", location.RealPath, err)
		}
		keyContent[basename] = string(content)
	}
	return keyContent, locations, nil
}

type AptConfiguration struct {
	// Add more data you want to capture as part of the package metadata here...
}
