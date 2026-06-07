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

	"github.com/domsnail/doctryne/internal/entity"
	"github.com/domsnail/doctryne/pkg/parsers/javascript_parsers"
	"github.com/domsnail/doctryne/pkg/types"
)

type ManifestServiceImpl struct {
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

	switch filepath.Base(manifest.Metadata.Filename) {
	case "package.json":
		manifest.WithLanguageType(types.Language_JavaScript, types.ManifestType_Package_Json)

		parser := javascript_parsers.Parser{}
		pkg, err := parser.
			WithContext(ctx).
			WithFile(bytes.NewReader(contents)).
			WithLockfile(manifest.Lockfile).
			ParseManifest()

		if err != nil {
			return fmt.Errorf("failed to parse manifest: %w", err)
		}

		manifest.AddPackage(pkg)
	case "":
		slog.InfoContext(ctx, "empty manifest filename provided, trying to determine type from contents...")

		return errors.New("manifest file auto-detection not supported, please specify a manifest type")
	default:
		return errors.New("manifest file type not supported")
	}

	return nil
}
