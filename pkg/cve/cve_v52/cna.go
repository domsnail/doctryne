package cve_v52

import (
	cve_v40 "github.com/domsnail/doctryne/pkg/cve/cve_v4"
)

type CNA struct {
	Title string `json:"title"`

	ProblemTypes []ProblemType `json:"problemTypes"`

	RejectedReasons []Description `json:"rejectedReasons"`

	Metrics []Metrics `json:"metrics"`

	References []Reference `json:"references"`

	Affected []Affected `json:"affected"`
	Impacts  []Impact   `json:"impacts"`

	ProviderMetadata ProviderMetadata `json:"providerMetadata"`

	Descriptions   []Description `json:"descriptions"`
	Configurations []Description `json:"configurations" `
	Workarounds    []Description `json:"workarounds" `
	Solutions      []Description `json:"solutions"`
	Exploits       []Description `json:"exploits"`

	Timeline []Timeline `json:"timeline"`

	PublicAt   *Timestamp `json:"datePublic"`
	AssignedAt *Timestamp `json:"dateAssigned"`

	Tags []string `json:"tags"`

	Source  *Source       `json:"source,omitempty"`
	Credits []Description `json:"credits"`

	// legacy data
	LegacyV4Record cve_v40.Record `json:"legacyV4Record"`
}

type ProblemType struct {
	Descriptions []Description `json:"descriptions"`
}

type Source struct {
	Advisory  string `json:"advisory"`
	Discovery string `json:"discovery"`
}
