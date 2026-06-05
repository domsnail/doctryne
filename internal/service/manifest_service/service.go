package manifest_service

import (
	"context"
	"crypto/md5"
	"errors"
	"fmt"
	"log/slog"

	"github.com/domsnail/doctryne/internal/entity"
	"github.com/domsnail/doctryne/pkg/parsers/javascript_parsers"
	"github.com/domsnail/doctryne/pkg/types"
	"github.com/domsnail/doctryne/pkg/utils"
)

type ManifestServiceImpl struct {
}

func (service ManifestServiceImpl) ProcessManifest(ctx context.Context, filename string, contents []byte) (*entity.Manifest, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	} else if len(contents) == 0 {
		return nil, errors.New("manifest content is empty")
	}

	slog.DebugContext(ctx, "calculating manifest checksum...")
	checksum := fmt.Sprintf("md5:%x", md5.Sum(contents))

	var manifest = entity.NewManifest().
		WithBinaryContents(contents).
		WithFile(filename, checksum).
		WithAuthor(utils.GetClientDataFromIncomingMetadata(ctx))

	slog.DebugContext(ctx, "trying to determine the type of manifest...",
		slog.String("filename", filename),
	)

	switch filename {
	case "package.json":
		manifest.WithLanguageType(types.Language_JavaScript, types.ManifestType_Package_Json)

		parser := javascript_parsers.Parser{}
		err := parser.ParseManifest(ctx, manifest)
		if err != nil {
			return nil, fmt.Errorf("failed to parse manifest: %w", err)
		}
	case "":
		slog.InfoContext(ctx, "empty manifest filename provided, trying to determine type from contents...")

		return nil, errors.New("manifest file auto-detection not supported, please specify a manifest type")
	default:
		return nil, errors.New("manifest file type not supported")
	}

	return manifest, nil
}
