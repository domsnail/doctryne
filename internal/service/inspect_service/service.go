package inspect_service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/domsnail/doctryne/cfg"
	"github.com/domsnail/doctryne/internal/entity"
	"github.com/domsnail/doctryne/internal/service"
	"github.com/domsnail/doctryne/pkg/npm"
	"github.com/domsnail/doctryne/pkg/types"
	"github.com/domsnail/doctryne/pkg/utils"
)

type InspectionService struct {
	manifests service.IManifestService

	github service.IGithubService
	npm    npm.Client
}

func (service *InspectionService) InitInspection(ctx context.Context, opts *entity.InspectionOptions) (*entity.Inspection, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	ins := entity.NewInspection(opts).WithAuthor(utils.GetClientDataFromIncomingMetadata(ctx))

	slog.InfoContext(ctx, "initializing new inspection...",
		slog.String("uuid", ins.UUID.String()),
		slog.String("scan_type", string(ins.ScanType)),
		slog.String("scan_mode", string(opts.Mode)),
		slog.String("uploaded_by", ins.UploadedBy),
	)

	switch ins.ScanType {
	case types.ScanType_URL:
		var buf bytes.Buffer
		_, err := buf.ReadFrom(ins.Target)
		if err != nil {
			return nil, fmt.Errorf("failed to read target: %w", err)
		}

		u, err := url.Parse(buf.String())
		if err != nil {
			return nil, fmt.Errorf("failed to parse target url: %w", err)
		}

		request, err := http.NewRequest(http.MethodGet, u.String(), nil)
		if err != nil {
			return nil, err
		}

		slog.DebugContext(ctx, "querying target url...",
			slog.String("scan_type", string(ins.ScanType)),
			slog.String("url", u.Redacted()),
		)

		resp, err := http.DefaultClient.Do(request.WithContext(ctx))
		if err != nil {
			return nil, fmt.Errorf("failed to perform http request: %w", err)
		}

		var manifest = entity.NewManifest()
		err = manifest.WithFilename(buf.String()).SetFileContent(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read manifest body contents: %w", err)
		}

		ins.AddManifest(manifest)
		break
	case types.ScanType_FilePath:
		var buf bytes.Buffer
		_, err := buf.ReadFrom(ins.Target)
		if err != nil {
			return nil, fmt.Errorf("failed to read target: %w", err)
		}

		file, err := os.ReadFile(buf.String())
		if err != nil {
			return nil, fmt.Errorf("failed to read from file: %w", err)
		}

		var manifest = entity.NewManifest()
		err = manifest.WithFilename(filepath.Base(buf.String())).SetFileContent(bytes.NewReader(file))
		if err != nil {
			return nil, fmt.Errorf("failed to read manifest body contents: %w", err)
		}

		ins.AddManifest(manifest)
		break
	case types.ScanType_DirPath:
		// todo: search for files
		return nil, errors.New("not implemented")
	case types.ScanType_Binary:
		var manifest = entity.NewManifest()
		err := manifest.SetFileContent(ins.Options.Target)
		if err != nil {
			return nil, fmt.Errorf("failed to read manifest body contents: %w", err)
		}

		ins.AddManifest(manifest)
		break
	default:
		return nil, errors.New("unspecified scan type")
	}

	if len(ins.Manifests) == 0 {
		return nil, errors.New("no manifests found")
	}

	if ins.Lockfile != nil {
		var lbuf bytes.Buffer
		n, err := lbuf.ReadFrom(ins.Lockfile)
		if err != nil {
			return nil, fmt.Errorf("failed to read lockfile: %w", err)
		} else if n == 0 {
			return nil, errors.New("lockfile is empty")
		}

		switch ins.ScanType {
		case types.ScanType_URL:
			u, err := url.Parse(lbuf.String())
			if err != nil {
				return nil, fmt.Errorf("failed to parse lockfile url: %w", err)
			}

			request, err := http.NewRequest(http.MethodGet, u.String(), nil)
			if err != nil {
				return nil, err
			}

			slog.DebugContext(ctx, "querying lockfile url...",
				slog.String("scan_type", string(ins.ScanType)),
				slog.String("url", u.Redacted()),
			)

			resp, err := http.DefaultClient.Do(request.WithContext(ctx))
			if err != nil {
				return nil, fmt.Errorf("failed to perform http request: %w", err)
			}

			err = ins.Manifests[0].SetLockfileContent(resp.Body)
			if err != nil {
				return nil, fmt.Errorf("failed to read lockfile body contents: %w", err)
			}
		case types.ScanType_DirPath:
			// todo: search for files
			return nil, errors.New("not implemented")
		case types.ScanType_FilePath:
			file, err := os.ReadFile(lbuf.String())
			if err != nil {
				return nil, fmt.Errorf("failed to read from lockfile: %w", err)
			}

			err = ins.Manifests[0].SetLockfileContent(bytes.NewReader(file))
			if err != nil {
				return nil, fmt.Errorf("failed to read lockfile body contents: %w", err)
			}
		}
	}

	return ins, nil
}

func (service *InspectionService) SearchManifestsInDir(ctx context.Context, target io.Reader) ([]*entity.Manifest, error) {
	maxDepth := cfg.GlobalConfig.Scan.FileSearchDepth

	var searchFilenames = make(map[string]int)
	for _, t := range types.ManifestTypes {
		searchFilenames[string(t)] = 0
	}

	var buf bytes.Buffer
	_, err := buf.ReadFrom(target)
	if err != nil {
		return nil, fmt.Errorf("failed to read target: %w", err)
	}

	if buf.String() == "" {
		return nil, fmt.Errorf("target directory path is empty")
	}

	root := filepath.Clean(buf.String())
	var manifests []*entity.Manifest

	slog.DebugContext(ctx, "searching manifests in directory...",
		slog.String("path", root),
		slog.Int("max_depth", maxDepth),
	)

	err = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		} else if ctx.Err() != nil {
			return ctx.Err()
		}

		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}

		depth := 0
		if rel != "." {
			depth = strings.Count(rel, string(filepath.Separator)) + 1
		}

		if d.IsDir() && maxDepth >= 0 && depth > maxDepth {
			return fs.SkipDir
		}

		if !d.IsDir() && maxDepth >= 0 && depth > maxDepth {
			return nil
		}

		if _, ok := searchFilenames[d.Name()]; ok {
			searchFilenames[d.Name()] = searchFilenames[d.Name()] + 1
			manifestType := types.ManifestType(d.Name())

			file, err := os.ReadFile(path)
			if err != nil {
				return fmt.Errorf("failed to read from file: %w", err)
			}

			manifest := entity.
				NewManifest().
				WithFilename(filepath.Base(path)).
				WithLanguageType(types.ManifestType_Language[manifestType], manifestType)

			err = manifest.SetFileContent(bytes.NewReader(file))
			if err != nil {
				return fmt.Errorf("failed to read manifest body contents: %w", err)
			}

			slog.DebugContext(ctx, "found manifest: "+d.Name(),
				slog.String("path", path),
				slog.Group("manifest_info",
					slog.String("uuid", manifest.UUID.String()),
					slog.String("filename", manifest.Metadata.Filename),
					slog.String("type", string(manifestType)),
					slog.String("language", string(types.ManifestType_Language[manifestType])),
				),
			)

			manifests = append(manifests, manifest)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to search manifests: %w", err)
	}

	var attrs []slog.Attr
	for k, v := range searchFilenames {
		attrs = append(attrs, slog.Attr{
			Key:   k,
			Value: slog.IntValue(v),
		})
	}

	slog.InfoContext(ctx, "successfully finished searching manifests in directory",
		slog.String("path", root),
		slog.Int("max_depth", maxDepth),
		slog.Int("matched_files", len(manifests)),
		slog.GroupAttrs("manifests_by_type",
			attrs...,
		),
	)

	return manifests, nil
}
