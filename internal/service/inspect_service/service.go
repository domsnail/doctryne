package inspect_service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"maps"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/domsnail/doctryne/cfg"
	"github.com/domsnail/doctryne/internal/entity"
	"github.com/domsnail/doctryne/internal/service"
	"github.com/domsnail/doctryne/pkg/stack_exchange"
	"github.com/domsnail/doctryne/pkg/types"
	"github.com/domsnail/doctryne/pkg/utils"
	"golang.org/x/sync/errgroup"
)

const (
	defaultDelay = 500 * time.Millisecond
)

type InspectionService struct {
	manifests service.IManifestService
	registry  service.IRegistryService

	github service.IGithubService

	stackExchange *stack_exchange.Client
}

func NewInspectionService(manifests service.IManifestService, github service.IGithubService, stackExchange *stack_exchange.Client, registry service.IRegistryService) *InspectionService {
	return &InspectionService{manifests: manifests, github: github, stackExchange: stackExchange, registry: registry}
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
	slog.DebugContext(ctx, "starting packages inspection...",
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

		time.Sleep(defaultDelay)
		pool.Inspect(pkg)
	}

	err = pool.Wait()
	if err != nil {
		slog.WarnContext(ctx, "manifest packages inspection failed",
			slog.String("error", err.Error()),
		)
	}

	inspection.Packages = packages

	// === analyse github ===

	slog.DebugContext(ctx, "starting packages github pages processing...",
		slog.Int("total_packages", len(inspection.Packages)),
	)

	var githubPool = NewGitHubInspectionPool(ctx, service.github, GitHubInspectionOptions{
		Mode:                       inspection.Options.Mode,
		DeepRepositoryInspection:   inspection.Options.DeepRepositoryInspection,
		ExtractFullContributorInfo: inspection.Options.ExtractFullContributorInfo,
		InspectIssues:              inspection.Options.InspectIssues,
	})

	for _, pkg := range inspection.Packages {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		gitUrl := pkg.GetGitURL()
		switch {
		case gitUrl == nil:
			slog.DebugContext(ctx, "missing git url, skipping package...",
				slog.String("package_name", pkg.Name),
			)

			continue
		case gitUrl.Host != "github.com":
			slog.WarnContext(ctx, "unsupported vcs url, skipping package...",
				slog.String("git_url", gitUrl.Redacted()),
				slog.String("package_name", pkg.Name),
			)

			continue
		}

		// dedupe same links is handled via github client (http) cache
		time.Sleep(defaultDelay)
		githubPool.Inspect(pkg)
	}

	err = githubPool.Wait()
	if err != nil {
		slog.WarnContext(ctx, "packages github pages inspection failed",
			slog.String("error", err.Error()),
		)
	}

	slog.InfoContext(ctx, "successfully inspected manifest packages",
		slog.Int("total_manifests", len(inspection.Manifests)),
		slog.Int("total_packages", len(inspection.Packages)),
		slog.Int("total_repositories", len(inspection.Repositories)),
	)

	// continue with git catalog analysis
	if inspection.Options.DeepRepositoryInspection {
		slog.DebugContext(ctx, "starting deep git repository inspection...")
		// todo: inspect git history
	}

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

func (service *InspectionService) InspectRepositories(ctx context.Context, inspection *entity.Inspection) error {
	dedupeRepositories(ctx, inspection)

	if inspection.Options.DeepRepositoryInspection {
		slog.DebugContext(ctx, "starting deep git repository inspection...")
	} else {
		slog.DebugContext(ctx, "starting shallow git repository inspection...")
	}

	return nil
}

func dedupeRepositories(ctx context.Context, inspection *entity.Inspection) {
	slog.DebugContext(ctx, "deduping repositories...")

	var (
		urlMap = make(map[string]*entity.Repository)
		total  = 0
	)

	for _, pkg := range inspection.Packages {
		if pkg.Git != nil && pkg.Git.Repository != nil {
			var key string
			total++

			if pkg.Git.Repository.GithubMetadata != nil && pkg.Git.Repository.GithubMetadata.GitURL != nil {
				key = pkg.Git.Repository.GithubMetadata.GitURL.Host + pkg.Git.Repository.GithubMetadata.GitURL.Path

			} else if pkg.Git.Repository.GitURL != nil {
				key = pkg.Git.Repository.GitURL.Host + pkg.Git.Repository.GitURL.Path

			} else {
				inspection.Repositories = append(inspection.Repositories, pkg.Git.Repository)
				continue
			}

			prev, ok := urlMap[key]
			if !ok {
				urlMap[key] = pkg.Git.Repository
				continue
			}

			// fill previous data
			if prev.GithubMetadata == nil {
				prev.GithubMetadata = pkg.Git.Repository.GithubMetadata
			}

			// reassign new ptr to previous value
			pkg.Git.Repository = prev
		}
	}

	inspection.Repositories = append(inspection.Repositories, slices.Collect(maps.Values(urlMap))...)
	slog.DebugContext(ctx, fmt.Sprintf("removed %d repositories after dedupe", total-len(inspection.Repositories)))

	return
}

func (service *InspectionService) InspectDevelopersAndOrganizations(ctx context.Context, inspection *entity.Inspection) error {
	if ctx.Err() != nil {
		return ctx.Err()
	} else if !(inspection.Options.ExtractFullContributorInfo || inspection.Options.ExtractFullOrganizationInfo) {
		slog.InfoContext(ctx, "skipped developers/organizations inspection")
		return nil
	}

	slog.DebugContext(ctx, "starting developers/organizations inspection...")

	developers, orgs := extractAndDedupeAllDevelopers(ctx, inspection)
	wg := sync.WaitGroup{}

	if inspection.Options.ExtractFullContributorInfo {
		wg.Go(func() {
			availableSources := service.availableSources()
			if len(availableSources) == 0 {
				slog.WarnContext(ctx, "skipped developers/organizations inspection: no available sources")
				return
			}

			if len(developers) == 0 {
				slog.DebugContext(ctx, "no developers found after dedupe")
				return
			}

			pool := NewDeveloperInspectionPool(ctx, service.github, service.stackExchange)

			for i := range developers {
				for _, source := range availableSources {
					time.Sleep(defaultDelay)
					pool.Inspect(developers[i], source)
				}
			}

			return
		})
	} else {
		slog.DebugContext(ctx, "skipped developers inspection")
	}

	if inspection.Options.ExtractFullOrganizationInfo {
		wg.Go(func() {
			if len(orgs) == 0 {
				slog.DebugContext(ctx, "no organizations found after dedupe")
				return
			}

			return
		})
	} else {
		slog.DebugContext(ctx, "skipped organizations inspection")
	}

	wg.Wait()
	slog.InfoContext(ctx, "developers/organizations inspection finished successfully")
	return nil
}

// extractAllDevelopers collects all developers from all embedded structs:
// - Packages...
// - Packages.RegistryMetadata.Contributors (from registry info)
// - Repositories.GithubMetadata.Contributors (from github project contributors info)
// - Repositories.GithubMetadata.Owner (from github project info)
// - Repositories.GithubMetadata.Org (some owners can also be organizations)
// - Repositories.Commiters (from git history)
func extractAndDedupeAllDevelopers(ctx context.Context, inspection *entity.Inspection) ([]*entity.Developer, []*entity.Organization) {
	slog.DebugContext(ctx, "extracting and deduping developers from previously collected data...")

	// todo: dedupe on same name to username

	var (
		dedupes    int
		conflicts  int
		developers []*entity.Developer
		orgs       []*entity.Organization

		// deduping maps
		uniqueUsernames = make(map[string]*entity.Developer)
		uniqueOrgLogins = make(map[string]*entity.Organization)
	)

	var dedupe = func(devs []*entity.Developer) {
		if devs == nil || len(devs) == 0 {
			return
		}

		for i, d := range devs {
			username := strings.ToLower(d.Username)

			last, ok := uniqueUsernames[username]
			if !ok {
				uniqueUsernames[username] = devs[i]
				developers = append(developers, devs[i])

				continue
			}

			err := last.Merge(devs[i])
			if err != nil {
				slog.DebugContext(ctx, "conflict on organization dedupe",
					slog.String("username", last.Username),
					slog.String("error", err.Error()),
				)

				developers = append(developers, devs[i])
				conflicts++
				continue
			}

			devs[i] = last
			dedupes++
		}
	}

	for _, p := range inspection.Packages {
		contrib := p.RegistryMetadata.Contributors

		dedupe(contrib.Sponsors)
		dedupe(contrib.Maintainers)
		dedupe(contrib.CodeOwners)
		dedupe(contrib.Authors)
		dedupe(contrib.Contributors)
	}

	for _, r := range inspection.Repositories {
		dedupe(r.Commiters)

		if r.GithubMetadata == nil {
			continue
		}

		dedupe(r.GithubMetadata.Contributors)

		if r.GithubMetadata.Owner != nil {
			ownerUsername := r.GithubMetadata.Owner.Username
			last, ok := uniqueUsernames[ownerUsername]
			if !ok {
				dev := r.GithubMetadata.Owner.ToDeveloper()
				uniqueUsernames[ownerUsername] = dev
				developers = append(developers, dev)
			} else {
				err := last.AddGithubProfile(r.GithubMetadata.Owner)
				if err != nil {
					slog.DebugContext(ctx, "conflict on repository owner dedupe",
						slog.String("username", r.GithubMetadata.Owner.Username),
						slog.String("error", err.Error()),
					)

					developers = append(developers, r.GithubMetadata.Owner.ToDeveloper())
					conflicts++
				} else {
					r.GithubMetadata.Owner = last.GithubMetadata
					dedupes++
				}
			}
		}

		if r.GithubMetadata.Org != nil {
			orgLogin := strings.ToLower(r.GithubMetadata.Org.Login)
			last, ok := uniqueOrgLogins[orgLogin]
			if !ok {
				uniqueOrgLogins[orgLogin] = r.GithubMetadata.Org
				orgs = append(orgs, r.GithubMetadata.Org)
			} else {
				err := last.Merge(r.GithubMetadata.Org)
				if err != nil {
					slog.DebugContext(ctx, "conflict on organization dedupe",
						slog.String("login", r.GithubMetadata.Org.Login),
						slog.String("error", err.Error()),
					)

					orgs = append(orgs, r.GithubMetadata.Org)
					conflicts++
				} else {
					r.GithubMetadata.Org = last
					dedupes++
				}
			}
		}
	}

	slog.InfoContext(ctx, fmt.Sprintf("removed %d developers/organizations after dedupe", dedupes),
		slog.Int("total_developers", len(developers)),
		slog.Int("total_organizations", len(orgs)),
		slog.Int("conflicts", conflicts),
	)

	return developers, orgs
}

func (service *InspectionService) CollectViolations(ctx context.Context, inspection *entity.Inspection) ([]*entity.Violation, error) {
	//TODO implement me
	panic("implement me")
}

func (service *InspectionService) availableSources() []InspectionSource {
	var s []InspectionSource

	if service.github != nil {
		s = append(s, InspectionSource_GitHub)
	}

	if service.stackExchange != nil {
		s = append(s, InspectionSource_StackExchange)
	}

	return s
}
