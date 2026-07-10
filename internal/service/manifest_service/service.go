package manifest_service

import (
	"bytes"
	"context"
	"crypto/md5"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"path/filepath"
	"strings"

	"github.com/domsnail/doctryne/internal/entity"
	"github.com/domsnail/doctryne/pkg/parsers/cyclonedx_parser"
	"github.com/domsnail/doctryne/pkg/parsers/javascript_parsers"
	"github.com/domsnail/doctryne/pkg/types"
)

type ManifestServiceImpl struct {
}

func NewManifestServiceImpl() *ManifestServiceImpl {
	return &ManifestServiceImpl{}
}

func (service ManifestServiceImpl) ProcessManifest(ctx context.Context, manifest *entity.Manifest) error {
	if ctx.Err() != nil {
		return ctx.Err()
	} else if manifest == nil {
		return errors.New("manifest is empty")
	}

	contents, err := io.ReadAll(manifest.Contents)
	if err != nil {
		return fmt.Errorf("failed to read manifest contents: %w", err)
	}

	slog.DebugContext(ctx, "calculating manifest checksum...")
	manifest.WithChecksum(fmt.Sprintf("md5:%x", md5.Sum(contents)))

	// todo: add cache lookup using checksum

	slog.DebugContext(ctx, "trying to determine the type of manifest...",
		slog.String("filename", manifest.Metadata.Filename),
		slog.String("checksum", manifest.Metadata.Checksum),
	)

	var t = manifest.Type
	if manifest.Type == types.ManifestType_Unspecified || manifest.Type == "" {
		t = types.ManifestType(strings.ToLower(filepath.Base(manifest.Metadata.Filename)))
	}

	var pkg *entity.Package

	switch t {
	case types.ManifestType_CycloneDX:
		parser := cyclonedx_parser.Parser{}
		pkg, err = parser.
			WithContext(ctx).
			WithFile(bytes.NewReader(contents)).
			ParseManifest()

	case types.ManifestType_Package_Json:
		manifest.WithLanguage(types.Language_JavaScript)

		parser := javascript_parsers.Parser{}
		pkg, err = parser.
			WithContext(ctx).
			WithFile(bytes.NewReader(contents)).
			WithLockfile(manifest.Lockfile).
			ParseManifest()

	case "":
		slog.InfoContext(ctx, "empty manifest filename provided, trying to determine type from contents...")

		return errors.New("manifest file auto-detection not supported, please specify a manifest type")
	default:
		return errors.New("manifest file type not supported")
	}

	if err != nil {
		return fmt.Errorf("failed to parse manifest: %w", err)
	}

	manifest.AddPackage(pkg)
	return nil
}
