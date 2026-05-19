package entity

import (
	"net/url"
	"time"

	"github.com/domsnail/doctryne/pkg/types"
)

type Package struct {
	Name string

	Version string
	Ref     string

	// Resolved url from which dependency was downloaded
	Resolved   *url.URL
	Registry   string
	RegistryID string

	// Labels custom aggregation labels to ease indexing, filtering and description
	Labels []types.Label

	// Integrity hash or check sum (SHA1, MD5 or other)
	Integrity string
	License   string

	Stats        *PackageStats
	Metadata     *PackageMetadata
	Contributors *PackageContributors

	PublishedAt time.Time
	ModifiedAt  time.Time
}

type PackageMetadata struct {
	Description string
	Homepage    string
	Keywords    []string
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
	Authors      []Person
	Contributors []Person
	Maintainers  []Person
}
