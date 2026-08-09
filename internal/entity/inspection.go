package entity

import (
	"io"

	"github.com/domsnail/doctryne/pkg/types"
	"github.com/google/uuid"
)

// Inspection is a resulting entity for a scan of a target (one inspection = multiple manifests) from API or CLI.
// If Inspection runs in dir mode multiple manifests can be found
type Inspection struct {
	UUID     uuid.UUID `json:"uuid"`
	Revision uint32    `json:"revision"`

	// target can be a binary file, url or directory path
	// todo: idea: maybe replace with slice of targets with target_type field?
	Target         io.Reader `json:"-"`
	TargetLockfile io.Reader `json:"-"`

	ScanType types.ScanType `json:"scan_type"`

	Manifests []*Manifest `json:"manifests"`

	Packages     []*Package    `json:"packages"`
	Repositories []*Repository `json:"repositories"`
	Developers   []*Developer  `json:"developers"`

	Options *InspectionOptions `json:"options"`

	UploadedBy   string `json:"uploaded_by"`
	UploadedFrom string `json:"uploaded_from"`
}

type InspectionOptions struct {
	ScanType types.ScanType

	Manifest     io.Reader `json:"-"`
	ManifestType types.ManifestType

	Lockfile     io.Reader `json:"-"`
	LockfileType types.ManifestType

	Mode types.InspectionMode

	ExtractFullOrganizationInfo bool
	ExtractFullContributorInfo  bool
	DeepRepositoryInspection    bool
	InspectIssues               bool
}

func NewInspection(opts *InspectionOptions) *Inspection {
	ins := Inspection{
		UUID:           uuid.Must(uuid.NewV7()),
		Target:         opts.Manifest,
		TargetLockfile: opts.Lockfile,
		ScanType:       opts.ScanType,
		Options:        opts,
		Manifests:      []*Manifest{},
	}

	return &ins
}

func (i *Inspection) WithAuthor(by, from string) *Inspection {
	i.UploadedBy = by
	i.UploadedFrom = from

	return i
}

func (i *Inspection) AddManifest(m *Manifest) {
	if i.Manifests == nil {
		i.Manifests = []*Manifest{}
	}

	i.Manifests = append(i.Manifests, m)
}
