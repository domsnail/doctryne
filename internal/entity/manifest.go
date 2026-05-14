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
