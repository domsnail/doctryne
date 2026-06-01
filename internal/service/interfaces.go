package service

import (
	"context"
	"net/url"

	"github.com/domsnail/doctryne/internal/entity"
)

type IInspectionService interface {
	CollectRepositoryInfo(ctx context.Context, opts entity.InspectionOptions) (*entity.Repository, error)
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

type IManifestAnalysisService interface {
	GetManifestInfo(ctx context.Context, file []byte) (entity.Manifest, error)
}

type IRepositoryAnalysisService interface {
}

type IPackageManagerService interface {
	GetPackage(ctx context.Context, name string) (*entity.Package, error)
}
