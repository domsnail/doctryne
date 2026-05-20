package service

import (
	"context"

	"github.com/domsnail/doctryne/internal/entity"
)

type IGithubService interface {
	GetRepositoryInfo(ctx context.Context) error
}

type IManifestAnalysisService interface {
	GetManifestInfo(ctx context.Context, file []byte) error
}

type IRepositoryAnalysisService interface {
}

type IPackageManagerService interface {
	GetPackage(ctx context.Context, name string) (*entity.Package, error)
}
