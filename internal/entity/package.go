package entity

import (
	"encoding/json"
	"net/url"
	"time"

	"github.com/domsnail/doctryne/pkg/types"
	"github.com/package-url/packageurl-go"
)

type Package struct {
	Name      string
	Version   string
	Ecosystem types.Ecosystem
	Language  types.Language

	// SBOM specific (identity)
	PackageURL packageurl.PackageURL
	CPE        string

	// Resolved url from which dependency was downloaded
	Resolved *url.URL
	Registry string

	// Labels custom aggregation labels to ease indexing, filtering and description
	Labels     []types.Label
	IsDev      bool
	IsOptional bool

	// Integrity hash or check sum (SHA1, MD5 or other)
	Integrity string

	RegistryMetadata *RegistryMetadata
	Git              *Git

	Dependencies []*Package

	Raw json.RawMessage
}

type AffiliatedDevelopers struct {
	CodeOwners   []*Developer
	Authors      []*Developer
	Contributors []*Developer
	Sponsors     []*Developer
	Maintainers  []*Developer
}

func (d *AffiliatedDevelopers) All() []*Developer {
	var devs []*Developer

	for _, dev := range d.CodeOwners {
		devs = append(devs, dev)
	}

	for _, dev := range d.Authors {
		devs = append(devs, dev)
	}

	for _, dev := range d.Contributors {
		devs = append(devs, dev)
	}

	for _, dev := range d.Sponsors {
		devs = append(devs, dev)
	}

	for _, dev := range d.Maintainers {
		devs = append(devs, dev)
	}

	return devs
}

type RegistryMetadata struct {
	RegistryID string
	IsPrivate  bool

	Description string
	Homepage    string
	Readme      string
	License     string
	Keywords    []string

	Git *url.URL

	Contributors *AffiliatedDevelopers
	Stats        *PackageStats

	PublishedAt time.Time
	ModifiedAt  time.Time
}

func (md *RegistryMetadata) IsPublished() bool {
	return !md.PublishedAt.IsZero()
}

type Git struct {
	Url *url.URL

	Repository *Repository
}

type PackageStats struct {
	Downloads PackageDownloads
}

type PackageDownloads struct {
	StartAt   time.Time
	EndAt     time.Time
	Downloads uint64
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

func (pkg *Package) GetGitURL() *url.URL {
	if pkg.Git != nil && pkg.Git.Url != nil {
		return pkg.Git.Url
	}

	if pkg.RegistryMetadata != nil && pkg.RegistryMetadata.Git != nil {
		return pkg.RegistryMetadata.Git
	}

	return nil
}
