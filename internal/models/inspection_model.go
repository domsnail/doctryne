package models

import (
	"time"

	"github.com/domsnail/doctryne/internal/entity"
	"github.com/domsnail/doctryne/pkg/types"
	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type InspectionModel struct {
	UUID     uuid.UUID `gorm:"column:uuid;primaryKey;type:uuid"`
	Revision uint32    `gorm:"column:revision;index"`

	ScanType types.ScanType `gorm:"column:scan_type;index"`

	Content datatypes.JSONType[*entity.Inspection] `gorm:"column:content;type:jsonb"`

	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func NewInspectionModel(inspection *entity.Inspection) *InspectionModel {
	model := InspectionModel{
		UUID:     inspection.UUID,
		Revision: inspection.Revision,
		ScanType: inspection.ScanType,
		Content:  datatypes.NewJSONType[*entity.Inspection](inspection),
	}

	return &model
}

func (model *InspectionModel) ToEntity() *entity.Inspection {
	content := model.Content.Data()
	if content == nil {
		return &entity.Inspection{
			UUID:      model.UUID,
			Revision:  model.Revision,
			ScanType:  model.ScanType,
			CreatedAt: model.CreatedAt,
			UpdatedAt: model.UpdatedAt,
		}
	}

	inspection := entity.Inspection{
		UUID:         model.UUID,
		Revision:     model.Revision,
		ScanType:     model.ScanType,
		Manifests:    content.Manifests,
		Packages:     content.Packages,
		Repositories: content.Repositories,
		Developers:   content.Developers,
		Options:      content.Options,
		UploadedBy:   content.UploadedBy,
		UploadedFrom: content.UploadedFrom,
		CreatedAt:    model.CreatedAt,
		UpdatedAt:    model.UpdatedAt,
	}

	return &inspection
}

func (model *InspectionModel) TableName() string {
	return "inspections"
}
