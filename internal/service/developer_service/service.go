package developer_service

import (
	"context"

	"github.com/domsnail/doctryne/internal/entity"
	"github.com/domsnail/doctryne/internal/service"
	"github.com/google/uuid"
)

type DeveloperServiceImpl struct {
	repo service.IDeveloperRepository
}

func (service *DeveloperServiceImpl) GetDeveloperByUUID(ctx context.Context, uid uuid.UUID) (*entity.Developer, error) {
	return service.repo.SelectDeveloperByUUID(ctx, uid)
}

func (service *DeveloperServiceImpl) GetDeveloperByQueryFilter(ctx context.Context, filter entity.DevelopersQueryFilter) ([]*entity.Developer, error) {
	return service.repo.SelectDevelopersByQueryFilter(ctx, filter)
}

func (service *DeveloperServiceImpl) SaveDeveloper(ctx context.Context, developer *entity.Developer) (error, int64) {
	return service.repo.UpsertDeveloper(ctx, developer)
}

func (service *DeveloperServiceImpl) SaveDevelopers(ctx context.Context, developers []*entity.Developer) (error, int64) {
	return service.repo.UpsertDevelopers(ctx, developers)
}
