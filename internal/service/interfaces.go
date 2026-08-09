package service

import (
	"context"
	"net/url"

	"github.com/domsnail/doctryne/internal/entity"
	"github.com/google/uuid"
)

type IInspectionService interface {
	// InitInspection creates a new inspection and populates it with manifests found from target
	InitInspection(ctx context.Context, opts *entity.InspectionOptions) (*entity.Inspection, error)

	// InspectManifests goes through all the collected manifests, collects all the packages and repositories
	InspectManifests(ctx context.Context, inspection *entity.Inspection) error

	// InspectPackages queries all the collected packages from respectable registries/vcs, finds contributors, owners,
	// download stats
	InspectPackages(ctx context.Context, inspection *entity.Inspection) error

	// InspectRepositories collects and dedupes all found vcs (git) repositories,
	// then downloads and analyzes commit history trees
	InspectRepositories(ctx context.Context, inspection *entity.Inspection) error

	InspectDevelopersAndOrganizations(ctx context.Context, inspection *entity.Inspection) error

	CollectViolations(ctx context.Context, inspection *entity.Inspection) ([]*entity.Violation, error)

	SaveInspection(ctx context.Context, inspection *entity.Inspection) error
	LoadInspection(ctx context.Context, uid uuid.UUID, rev uint32) (*entity.Inspection, error)
}

type IInspectionsRepository interface {
	CreateInspection(ctx context.Context, ins *entity.Inspection) error

	GetInspection(ctx context.Context, uid uuid.UUID) (*entity.Inspection, error)
	GetInspectionRevision(ctx context.Context, uid uuid.UUID, rev uint32) (*entity.Inspection, error)
}

type IGithubService interface {
	GetRepositoryByName(ctx context.Context, owner string, name string) (*entity.GitHubRepositoryMetadata, error)
	GetRepositoryByURL(ctx context.Context, link *url.URL) (*entity.GitHubRepositoryMetadata, error)
	GetRepositoryLanguages(ctx context.Context, owner string, name string) (map[string]int, error)
	GetRepositoryContributors(ctx context.Context, owner string, name string) ([]*entity.Developer, error)
	GetRepositoryIssues(ctx context.Context, owner string, name string) ([]*entity.GithubIssue, error)

	GetUserOwnedRepositories(ctx context.Context, username string) ([]*entity.GitHubRepositoryMetadata, error)

	GetProfileByUsername(ctx context.Context, username string) (*entity.GithubDeveloperProfile, error)
	GetProfileByID(ctx context.Context, id int64) (*entity.GithubDeveloperProfile, error)

	GetUserActivity(ctx context.Context, username string) (*entity.Activity, error)

	GetOrganizationByName(ctx context.Context, name string) (*entity.Organization, error)
	GetOrganizationUsers(ctx context.Context, name string) ([]*entity.GithubDeveloperProfile, error)
}

type IStackExchangeService interface {
	GetProfileByUsername(ctx context.Context, username string) (*entity.StackExchangeDeveloperProfile, error)
	GetProfileByID(ctx context.Context, id int64) (*entity.StackExchangeDeveloperProfile, error)
}

type IManifestService interface {
	ProcessManifest(ctx context.Context, manifest *entity.Manifest) error
}

type IRepositoryService interface {
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
