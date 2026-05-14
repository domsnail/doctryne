package service

import "context"

type IManifestAnalysisService interface {
	GetManifestInfo(ctx context.Context, file []byte) error
}

type IRepositoryAnalysisService interface {
}
