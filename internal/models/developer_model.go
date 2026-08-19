package models

import (
	"database/sql"
	"strings"
	"time"

	"github.com/domsnail/doctryne/internal/entity"
	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type DeveloperModel struct {
	UUID uuid.UUID `gorm:"column:uuid;primaryKey;type:uuid"`

	Username sql.NullString              `gorm:"column:username;index"`
	FullName sql.NullString              `gorm:"column:full_name"`
	Emails   datatypes.JSONSlice[string] `gorm:"column:emails"`

	GithubID      *int64                                             `gorm:"column:github_id;index"`
	GithubProfile datatypes.JSONType[*entity.GithubDeveloperProfile] `gorm:"column:github_profile;type:jsonb"`

	StackExchangeAccountID *uint64                                                   `gorm:"column:stack_exchange_account_id;index"`
	StackExchangeProfile   datatypes.JSONType[*entity.StackExchangeDeveloperProfile] `gorm:"column:stack_exchange_profile;type:jsonb"`

	CreatedAt    time.Time  `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt    time.Time  `gorm:"column:updated_at;autoUpdateTime"`
	LastLookupAt *time.Time `gorm:"column:last_lookup_at"`
}

func NewDeveloperModel(developer *entity.Developer) *DeveloperModel {
	model := DeveloperModel{
		UUID:                   developer.UUID,
		Emails:                 developer.Emails,
		GithubID:               developer.GithubID,
		StackExchangeAccountID: developer.StackExchangeAccountID,
		LastLookupAt:           developer.LastLookupAt,
	}

	if developer.Username != "" {
		model.Username = sql.NullString{String: developer.Username, Valid: true}
	}

	if developer.Name != "" {
		model.FullName = sql.NullString{String: developer.Name, Valid: true}
	}

	if developer.GithubProfile != nil {
		model.GithubProfile = datatypes.NewJSONType[*entity.GithubDeveloperProfile](developer.GithubProfile)
	}

	if developer.StackExchangeProfile != nil {
		model.StackExchangeProfile = datatypes.NewJSONType[*entity.StackExchangeDeveloperProfile](developer.StackExchangeProfile)
	}

	return &model
}

func (model *DeveloperModel) ToEntity() *entity.Developer {
	developer := entity.Developer{
		UUID:                   model.UUID,
		Name:                   model.FullName.String,
		Username:               model.Username.String,
		Emails:                 model.Emails,
		GithubID:               model.GithubID,
		StackExchangeAccountID: model.StackExchangeAccountID,
		CreatedAt:              model.CreatedAt,
		UpdatedAt:              model.UpdatedAt,
		LastLookupAt:           model.LastLookupAt,
	}

	githubProfile := model.GithubProfile.Data()
	if githubProfile == nil {
		developer.GithubProfile = githubProfile
	}

	stackExchangeProfile := model.StackExchangeProfile.Data()
	if stackExchangeProfile == nil {
		developer.StackExchangeProfile = stackExchangeProfile
	}

	return &developer
}

func (model *DeveloperModel) TableName() string {
	return "developers"
}

func (model *DeveloperModel) BeforeSave() error {
	if model.Username.Valid {
		model.Username.String = strings.ToLower(strings.TrimSpace(model.Username.String))
	}

	return nil
}
