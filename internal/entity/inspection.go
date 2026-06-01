package entity

import (
	"io"

	"github.com/domsnail/doctryne/pkg/types"
	"github.com/google/uuid"
)

// Inspection is a resulting entity for a single scan of a target (one manifest = one inspection) from API or CLI
type Inspection struct {
	UUID uuid.UUID

	// target can be a binary file, url, package or module name
	Target   io.Reader
	ScanType types.ScanType

	Manifest *Manifest

	Packages     []*Package
	Repositories []*Repository
	Developers   []*Developer

	Options *InspectionOptions
}

type InspectionOptions struct {
	ScanType types.ScanType
	Target   io.Reader

	Mode types.InspectionMode

	// LoadUserProfiles is set to true, all users profiles (contributors) will be additionally
	// loaded (their profile, activity, repositories)
	LoadUserProfiles bool
}

func NewInspection(opts *InspectionOptions) *Inspection {
	ins := Inspection{
		UUID:    uuid.Must(uuid.NewV7()),
		Target:  opts.Target,
		Options: opts,
	}

	// resolve target content

	return &ins
}

func (ins *Inspection) WithManifest(m *Manifest) *Inspection {
	ins.Manifest = m
	return ins
}
