package entity

import (
	"time"

	"github.com/domsnail/doctryne/pkg/types"
	"github.com/google/uuid"
)

type Manifest struct {
	UUID     uuid.UUID
	Metadata ManifestMetadata

	Language types.Language
	Type     types.ManifestType

	Raw []byte
}

func NewManifest() *Manifest {
	return &Manifest{
		UUID: uuid.Must(uuid.NewV7()),
		Metadata: ManifestMetadata{
			UploadedAt: time.Now(),
			UpdatedAt:  time.Now(),
		},
	}
}

func (m *Manifest) WithAuthor(by, from string) *Manifest {
	m.Metadata.UploadedBy = by
	m.Metadata.UploadedFrom = from

	return m
}

func (m *Manifest) WithFile(filename string, checksum uint64) *Manifest {
	m.Metadata.Filename = filename
	m.Metadata.Checksum = checksum

	return m
}

func (m *Manifest) WithLanguageType(lang types.Language, tp types.ManifestType) *Manifest {
	m.Language = lang
	m.Type = tp

	return m
}

type ManifestMetadata struct {
	Filename string
	Checksum uint64

	UploadedBy   string
	UploadedFrom string

	UploadedAt time.Time
	UpdatedAt  time.Time
}
