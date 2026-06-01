package entity

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
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
			UpdatedAt: time.Now(),
		},
	}
}

func (m *Manifest) WithAuthor(by, from string) *Manifest {
	m.Metadata.UploadedBy = by
	m.Metadata.UploadedFrom = from

	return m
}

func (m *Manifest) SetFileContent(content io.Reader) error {
	if content != nil {
		var err error

		m.Raw, err = io.ReadAll(content)
		if err != nil {
			return err
		}

		m.Metadata.UpdatedAt = time.Now()
		m.Metadata.Checksum = "sha256:" + hex.EncodeToString(sha256.New().Sum(m.Raw))
	}

	return nil
}

func (m *Manifest) WithLanguageType(lang types.Language, tp types.ManifestType) *Manifest {
	m.Language = lang
	m.Type = tp

	return m
}

type ManifestMetadata struct {
	Filename string
	Checksum string

	UploadedBy   string
	UploadedFrom string

	UploadedAt time.Time
	UpdatedAt  time.Time
}
