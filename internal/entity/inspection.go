package entity

import (
	"github.com/domsnail/doctryne/pkg/types"
	"github.com/google/uuid"
)

// Inspection is a resulting entity for a single scan of a target (one manifest = one inspection) from API or CLI
type Inspection struct {
	uid uuid.UUID

	// target can be a file, url, package or module name
	target   string
	manifest *Manifest

	packages     []*Package
	repositories []*Repository
	developers   []*Developer

	opts *InspectionOptions
}

func NewInspection(opts *InspectionOptions) *Inspection {
	ins := Inspection{
		target: opts.Target,
		uid:    uuid.Must(uuid.NewV7()),
		opts:   opts,
	}

	return &ins
}

func (ins *Inspection) WithManifest(manifest *Manifest) *Inspection {
	ins.manifest = manifest
	return ins
}

type InspectionOptions struct {
	Target string

	Mode types.InspectionMode

	// LoadUserProfiles is set to true, all users profiles (contributors) will be additionally
	// loaded (their profile, activity, repositories)
	LoadUserProfiles bool
}
