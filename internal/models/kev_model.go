package models

import "time"

type KnownExploitVulnerabilitiesModel struct {
	CanonicalID string `gorm:"column:canonical_id;type:citext;primary_key"`

	IsExploitAvailable bool `gorm:"column:is_exploit_available"`
	IsListed           bool `gorm:"column:is_listed"`

	AddedAt time.Time `gorm:"column:added_at;type:timestamptz"`
	DueDate time.Time `gorm:"column:due_date;type:timestamptz"`

	RequiredAction string `gorm:"column:required_action"`
}
