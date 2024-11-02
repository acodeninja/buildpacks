package sbom

import (
	"context"
	"crypto"
	"github.com/anchore/syft/syft"
	"github.com/anchore/syft/syft/cataloging/filecataloging"
	"github.com/anchore/syft/syft/cataloging/pkgcataloging"
	"github.com/anchore/syft/syft/file"
	"github.com/anchore/syft/syft/pkg"
	"github.com/anchore/syft/syft/sbom"
)

type Source struct {
}

func Generate(cataloger pkg.Cataloger) sbom.SBOM {
	src, err := syft.GetSource(context.Background(), "", nil)

	cfg := syft.DefaultCreateSBOMConfig().
		//WithParallelism(5).
		//WithTool("my-tool", "v1.0").
		WithFilesConfig(
			filecataloging.
				DefaultConfig().
				WithSelection(file.AllFilesSelection).
				WithHashers(
					crypto.MD5,
					crypto.SHA1,
					crypto.SHA256,
				),
		).
		WithCatalogers(pkgcataloging.NewAlwaysEnabledCatalogerReference(cataloger))

	s, err := syft.CreateSBOM(context.Background(), src, cfg)
	if err != nil {
		panic(err)
	}

	return *s
}
