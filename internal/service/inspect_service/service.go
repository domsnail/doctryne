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
	"sync"

	"github.com/domsnail/doctryne/cfg"
	"github.com/domsnail/doctryne/internal/entity"
	"github.com/domsnail/doctryne/internal/service"
	"github.com/domsnail/doctryne/pkg/types"
	"github.com/domsnail/doctryne/pkg/utils"
	"golang.org/x/sync/errgroup"
)

type InspectionService struct {
	manifests service.IManifestService

	github   service.IGithubService
	registry service.IRegistryService
}

func NewInspectionService(manifests service.IManifestService, github service.IGithubService, registry service.IRegistryService) *InspectionService {
	return &InspectionService{manifests: manifests, github: github, registry: registry}
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
		err = manifest.WithFilename(buf.String()).WithType(ins.Options.ManifestType).SetFileContent(resp.Body)
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

		var manifest = entity.NewManifest().WithType(ins.Options.ManifestType)
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
		var manifest = entity.NewManifest().WithType(ins.Options.ManifestType)
		err := manifest.SetFileContent(ins.Options.Manifest)
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

	if ins.TargetLockfile != nil {
		var lbuf bytes.Buffer
		n, err := lbuf.ReadFrom(ins.TargetLockfile)
		if err != nil {
			return nil, fmt.Errorf("failed to read lockfile: %w", err)
		} else if n != 0 {

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
	}

	return ins, nil
}

func (service *InspectionService) searchManifestsInDir(ctx context.Context, target io.Reader) ([]*entity.Manifest, error) {
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
				WithType(manifestType).
				WithLanguage(types.ManifestType_Language[manifestType])

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

func (service *InspectionService) InspectPackages(ctx context.Context, inspection *entity.Inspection) error {
	slog.DebugContext(ctx, "starting inspection packages processing...",
		slog.Int("total_manifests", len(inspection.Manifests)),
	)

	packages, err := service.extractPackages(ctx, inspection)
	if err != nil {
		return fmt.Errorf("failed to extract packages: %w", err)
	}

	var pool = NewPackageInspectionPool(ctx, service.registry)
	for _, pkg := range packages {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		pool.Inspect(pkg)
	}

	err = pool.Wait()
	if err != nil {
		slog.WarnContext(ctx, "manifest packages inspection failed",
			slog.String("error", err.Error()),
		)
	}

	inspection.Packages = packages
	return nil
}

func (service *InspectionService) inspectPackage(ctx context.Context, pkg *entity.Package) error {
	return nil
}

func (service *InspectionService) extractPackages(ctx context.Context, inspection *entity.Inspection) ([]*entity.Package, error) {
	slog.DebugContext(ctx, "extracting packages...",
		slog.Int("total_manifests", len(inspection.Manifests)),
	)

	depsMap := sync.Map{}
	group, groupCtx := errgroup.WithContext(ctx)
	for _, m := range inspection.Manifests {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		if len(m.DiscoveredPackages) == 0 {
			slog.DebugContext(groupCtx, "no packages discovered in manifest, skipping...",
				slog.String("manifest", m.Metadata.Filename),
			)

			continue
		}

		for _, pkg := range m.DiscoveredPackages {
			slog.DebugContext(groupCtx, "inspecting top level package",
				slog.String("package_name", pkg.Name),
				slog.String("package_version", pkg.Version),
			)

			group.Go(func() error {
				dependencies, err := service.extractPackage(groupCtx, pkg)
				if err != nil {
					return fmt.Errorf("failed to inspect package '%s': %w", pkg.Name, err)
				}

				for _, d := range dependencies {
					depsMap.Store(fmt.Sprintf("%s@%s", d.Name, d.Version), d)
				}

				return nil
			})
		}
	}

	err := group.Wait()
	if err != nil {
		slog.WarnContext(ctx, "packages processing completed with errors",
			slog.String("error", err.Error()),
		)
	}

	var depsSlice []*entity.Package
	depsMap.Range(func(k, v interface{}) bool {
		depsSlice = append(depsSlice, v.(*entity.Package))
		return true
	})

	slog.DebugContext(groupCtx, "packages extracted successfully",
		slog.Int("total_unique_packages", len(depsSlice)),
		slog.Int("total_manifests", len(inspection.Manifests)),
	)

	return depsSlice, nil
}

func (service *InspectionService) extractPackage(ctx context.Context, pkg *entity.Package) ([]*entity.Package, error) {
	var pkgs []*entity.Package

	var children = pkg.Dependencies[:]
	pkg.Dependencies = nil

	pkgs = append(pkgs, pkg)

	for _, child := range children {
		childPkgs, err := service.extractPackage(ctx, child)
		if err != nil {
			return nil, err
		}

		pkgs = append(pkgs, childPkgs...)
	}

	return pkgs, nil
}

func (service *InspectionService) InspectDevelopers(ctx context.Context, inspection *entity.Inspection) error {
	//TODO implement me
	panic("implement me")
}

func (service *InspectionService) CollectViolations(ctx context.Context, inspection *entity.Inspection) ([]*entity.Violation, error) {
	//TODO implement me
	panic("implement me")
}
