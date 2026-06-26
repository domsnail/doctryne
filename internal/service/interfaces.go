package service

import (
	"context"
	"net/url"

	"github.com/domsnail/doctryne/internal/entity"
)

type IInspectionService interface {
	// InitInspection creates a new inspection and populates it with manifests found from target
	InitInspection(ctx context.Context, opts *entity.InspectionOptions) (*entity.Inspection, error)

	// InspectManifests goes through all the collected manifests, collects all the packages and repositories
	InspectManifests(ctx context.Context, inspection *entity.Inspection) error

	// InspectPackages queries all the collected packages from respectable registries/vcs, finds contributors, owners,
	// repository stats, commit history etc.
	InspectPackages(ctx context.Context, inspection *entity.Inspection) error

	InspectDevelopers(ctx context.Context, inspection *entity.Inspection) error

	CollectViolations(ctx context.Context, inspection *entity.Inspection) ([]*entity.Violation, error)
}

type IGithubService interface {
	GetRepositoryByName(ctx context.Context, owner string, name string) (*entity.Repository, error)
	GetRepositoryByURL(ctx context.Context, link *url.URL) (*entity.Repository, error)
	GetRepositoryContributors(ctx context.Context, owner string, name string) ([]*entity.Developer, error)

	GetUserOwnedRepositories(ctx context.Context, username string) ([]*entity.Repository, error)

	GetUserByUsername(ctx context.Context, username string) (*entity.Developer, error)

	GetUserActivity(ctx context.Context, username string) (*entity.Activity, error)

	GetOrganizationByName(ctx context.Context, name string) (*entity.Organization, error)
	GetOrganizationUsers(ctx context.Context, name string) ([]*entity.Developer, error)
}

type IManifestService interface {
	ProcessManifest(ctx context.Context, manifest *entity.Manifest) error
}

type IRepositoryAnalysisService interface {
}

type IPackageManagerService interface {
	GetPackage(ctx context.Context, name string) (*entity.Package, error)
}

type IManifestParser interface {
	ParseManifest(ctx context.Context, manifest *entity.Manifest) error
}

type IRegistryService interface {
	// GetPackageInfo queries all available data from relevant registry (npm, maven etc.), latest stats and other available data
	GetPackageInfo(ctx context.Context, pkg *entity.Package) error
}

type IGitHistoryService interface {
	InspectRepository(ctx context.Context, link *url.URL, branch string) (*entity.Repository, error)
}
