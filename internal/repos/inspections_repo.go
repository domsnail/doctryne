package repos

import (
	"context"
	"errors"

	"github.com/domsnail/doctryne/internal/entity"
	"github.com/domsnail/doctryne/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type InspectionsRepoImpl struct {
	orm *gorm.DB
}

func NewInspectionsRepoImpl(orm *gorm.DB) *InspectionsRepoImpl {
	return &InspectionsRepoImpl{orm: orm}
}

func (repo *InspectionsRepoImpl) CreateInspection(ctx context.Context, inspection *entity.Inspection) error {
	if inspection == nil {
		return errors.New("inspection is nil")
	}

	model := models.NewInspectionModel(inspection)
	err := repo.orm.WithContext(ctx).Create(model).Error
	if err != nil {
		return err
	}

	inspection.CreatedAt = model.CreatedAt
	inspection.UpdatedAt = model.UpdatedAt
	return nil
}

func (repo *InspectionsRepoImpl) SelectInspectionByUUID(ctx context.Context, uid uuid.UUID) (*entity.Inspection, error) {
	if uid.String() == "" {
		return nil, errors.New("uuid is empty")
	}

	var model models.InspectionModel
	err := repo.orm.WithContext(ctx).Where("uuid = ?", uid).First(&model).Order("revision DESC").Error
	if err != nil {
		return nil, err
	}

	return model.ToEntity(), nil
}

func (repo *InspectionsRepoImpl) SelectInspectionsByQueryFilter(ctx context.Context, filter entity.InspectionsQueryFilter) ([]*entity.Inspection, error) {
	var model []*models.InspectionModel

	err := repo.orm.WithContext(ctx).
		Select("uuid, scan_type, MAX(created_at) AS created_at, MAX(updated_at) AS updated_at").
		Group("uuid, scan_type").
		Find(&model).
		Offset(filter.Offset).
		Limit(filter.Limit).
		Error

	if err != nil {
		return nil, err
	}

	var inspections = make([]*entity.Inspection, len(model))
	for i, m := range model {
		inspections[i] = m.ToEntity()
	}

	return inspections, nil
}

func (repo *InspectionsRepoImpl) SelectInspectionRevisionByUUID(ctx context.Context, uid uuid.UUID, rev uint32) (*entity.Inspection, error) {
	if uid.String() == "" {
		return nil, errors.New("uuid is empty")
	}

	var model models.InspectionModel
	err := repo.orm.WithContext(ctx).Where("uuid = ? AND revision = ?", uid, rev).First(&model, uid.String()).Error
	if err != nil {
		return nil, err
	}

	return model.ToEntity(), nil
}
