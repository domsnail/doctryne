package entity

import (
	"net/url"

	"github.com/domsnail/doctryne/pkg/types"
)

type Dependency struct {
	Name string

	Version string
	Ref     string

	// Resolved url from which dependency was downloaded
	Resolved *url.URL

	IsRequired      bool
	IsDevDependency bool

	// Labels custom aggregation labels to ease indexing, filtering and description
	Labels []types.Label

	// Integrity hash or check sum (SHA1, MD5 or other)
	Integrity string
	License   string
}
