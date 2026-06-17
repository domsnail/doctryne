package entity

import (
	"encoding/json"
	"net/url"
	"time"

	"github.com/domsnail/doctryne/pkg/types"
)

type Package struct {
	Name      string
	Version   string
	Ecosystem types.Ecosystem
	Language  types.Language

	// Resolved url from which dependency was downloaded
	Resolved   *url.URL
	Registry   string
	RegistryID string

	// Labels custom aggregation labels to ease indexing, filtering and description
	Labels     []types.Label
	IsDev      bool
	IsOptional bool

	// Integrity hash or check sum (SHA1, MD5 or other)
	Integrity string
	License   string

	AffiliatedDevelopers AffiliatedDevelopers

	Metadata *PackageMetadata

	Dependencies []*Package

	Raw json.RawMessage
}

type AffiliatedDevelopers struct {
	CodeOwners   []Developer
	Authors      []Developer
	Contributors []Developer
	Sponsors     []Developer
	Maintainers  []Developer
}

type PackageMetadata struct {
	Description string
	Homepage    string
	Keywords    []string

	Git *url.URL

	Contributors *PackageContributors
	Stats        *PackageStats

	PublishedAt time.Time
	ModifiedAt  time.Time
}

type PackageStats struct {
	Downloads PackageDownloads
}

type PackageDownloads struct {
	StartAt   time.Time
	EndAt     time.Time
	Downloads uint64
}

type PackageContributors struct {
	Authors      []Developer
	Contributors []Developer
	Maintainers  []Developer
}

func (pkg *Package) CountDevDependencies() int {
	var counter int

	for _, dep := range pkg.Dependencies {
		if dep.IsDev {
			counter++
		}
	}

	return counter
}

func (pkg *Package) CountOptionalDependencies() int {
	var counter int

	for _, dep := range pkg.Dependencies {
		if dep.IsOptional {
			counter++
		}
	}

	return counter
}
