package entity

import (
	"errors"
	"io"
	"time"

	"github.com/domsnail/doctryne/pkg/types"
	"github.com/google/uuid"
)

// Manifest contains information about incoming application packages. Represents package.json, go.mod and other
// language manifests and/or software bill of materials.
type Manifest struct {
	UUID     uuid.UUID
	Metadata ManifestMetadata

	Language types.Language
	Type     types.ManifestType

	DiscoveredPackages []*Package

	Lockfile io.Reader
	Contents io.Reader
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

func (m *Manifest) WithFilename(filename string) *Manifest {
	m.Metadata.Filename = filename

	return m
}

func (m *Manifest) WithChecksum(checksum string) *Manifest {
	m.Metadata.Checksum = checksum

	return m
}

func (m *Manifest) WithLanguageType(lang types.Language, tp types.ManifestType) *Manifest {
	m.Language = lang
	m.Type = tp

	return m
}

func (m *Manifest) SetFileContent(reader io.Reader) error {
	if reader == nil {
		return errors.New("reader is nil")
	}

	m.Contents = reader
	return nil
}

func (m *Manifest) SetLockfileContent(reader io.Reader) error {
	if reader == nil {
		return errors.New("reader is nil")
	}

	m.Lockfile = reader
	return nil
}

func (m *Manifest) AddPackage(pkg *Package) {
	if m.DiscoveredPackages == nil {
		m.DiscoveredPackages = []*Package{}
	}

	m.DiscoveredPackages = append(m.DiscoveredPackages, pkg)
}

type ManifestMetadata struct {
	Filename string
	Checksum string

	UploadedBy   string
	UploadedFrom string

	UploadedAt time.Time
	UpdatedAt  time.Time
}

func (m *Manifest) CountDevDependencies() int {
	var counter int

	for _, pkg := range m.DiscoveredPackages {
		counter += pkg.CountDevDependencies()
	}

	return counter
}

func (m *Manifest) CountOptionalDependencies() int {
	var counter int

	for _, pkg := range m.DiscoveredPackages {
		counter += pkg.CountOptionalDependencies()
	}

	return counter
}
