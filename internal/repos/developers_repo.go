package repos

import (
	"context"
	"errors"

	"github.com/domsnail/doctryne/internal/entity"
	"github.com/domsnail/doctryne/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DevelopersRepoImpl struct {
	orm *gorm.DB
}

func (repo *DevelopersRepoImpl) CreateDeveloper(ctx context.Context, developer *entity.Developer) error {
	if developer == nil {
		return errors.New("developer is nil")
	}

	model := models.NewDeveloperModel(developer)
	err := repo.orm.WithContext(ctx).Create(model).Error
	if err != nil {
		return err
	}

	developer.CreatedAt = model.CreatedAt
	developer.UpdatedAt = model.UpdatedAt
	return nil
}

func (repo *DevelopersRepoImpl) UpsertDeveloper(ctx context.Context, developer *entity.Developer) (error, int64) {
	if developer == nil {
		return errors.New("developer is nil"), 0
	}

	model := models.NewDeveloperModel(developer)
	query := repo.orm.WithContext(ctx).FirstOrCreate(model) // todo: use assign
	if query.Error != nil {
		return query.Error, query.RowsAffected
	}

	developer.CreatedAt = model.CreatedAt
	developer.UpdatedAt = model.UpdatedAt

	return nil, query.RowsAffected
}

func (repo *DevelopersRepoImpl) UpsertDevelopers(ctx context.Context, developers []*entity.Developer) (error, int64) {
	//TODO implement me
	panic("implement me")
}

func (repo *DevelopersRepoImpl) SelectDeveloperByUUID(ctx context.Context, uid uuid.UUID) (*entity.Developer, error) {
	if uid.String() == "" {
		return nil, errors.New("uuid is empty")
	}

	var model models.DeveloperModel
	err := repo.orm.WithContext(ctx).Where("uuid = ?", uid).First(&model).Error
	if err != nil {
		return nil, err
	}

	return model.ToEntity(), nil
}

func (repo *DevelopersRepoImpl) SelectDevelopersByQueryFilter(ctx context.Context, filter entity.DevelopersQueryFilter) ([]*entity.Developer, error) {
	//TODO implement me
	panic("implement me")
}
