package entity

import (
	"github.com/domsnail/doctryne/pkg/types"
	"github.com/google/uuid"
)

type Manifest struct {
	UUID     uuid.UUID
	Filename string
	Checksum uint64

	Language types.Language
	Type     types.ManifestType
}

func NewManifest() *Manifest {
	return &Manifest{
		UUID: uuid.Must(uuid.NewV7()),
	}
}

func (m *Manifest) WithFile(filename string, checksum uint64) *Manifest {
	m.Filename = filename
	m.Checksum = checksum

	return m
}

func (m *Manifest) WithLanguageType(lang types.Language, tp types.ManifestType) *Manifest {
	m.Language = lang
	m.Type = tp

	return m
}
