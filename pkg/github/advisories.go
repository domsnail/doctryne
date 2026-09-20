package github

import "time"

type AdvisoriesQueryOptions struct {
	ModifiedAfter  *time.Time
	ModifiedBefore *time.Time

	CveID  string
	GhsaID string

	Sort      string
	Direction string
}

type AdvisoryRecord struct {
	GhsaId      string       `json:"ghsa_id"`
	CveId       string       `json:"cve_id"`
	Identifiers []Identifier `json:"identifiers"`

	Url string `json:"url"`
	//HtmlUrl string `json:"html_url"`

	Summary     string `json:"summary"`
	Description string `json:"description"`

	Type     AdvisoryType `json:"type"`
	Severity string       `json:"severity"`
	Cwes     []Cwe        `json:"cwes,omitempty"`

	RepositoryAdvisoryUrl string `json:"repository_advisory_url,omitempty"`
	SourceCodeLocation    string `json:"source_code_location,omitempty"`

	Vulnerabilities []Vulnerability `json:"vulnerabilities"`

	CvssSeverities CvssSeverities `json:"cvss_severities"`

	Epss Epss `json:"epss"`

	References []string `json:"references"`

	Credits []Credit `json:"credits,omitempty"`

	PublishedAt      time.Time `json:"published_at"`
	UpdatedAt        time.Time `json:"updated_at"`
	GithubReviewedAt time.Time `json:"github_reviewed_at"`

	NvdPublishedAt *time.Time `json:"nvd_published_at,omitempty"`
	WithdrawnAt    *time.Time `json:"withdrawn_at,omitempty"`
}

type Identifier struct {
	Value string `json:"value"`
	Type  string `json:"type"`
}

type Vulnerability struct {
	Package *Package `json:"package,omitempty"`

	VulnerableVersionRange string `json:"vulnerable_version_range"`
	FirstPatchedVersion    string `json:"first_patched_version"`

	VulnerableFunctions []string `json:"vulnerable_functions,omitempty"`
}

type Package struct {
	Ecosystem string `json:"ecosystem"`
	Name      string `json:"name"`
}

type Epss struct {
	Percentage float32 `json:"percentage"`
	Percentile float32 `json:"percentile"`
}

type Cwe struct {
	CweId string `json:"cwe_id"`
	Name  string `json:"name"`
}

type CvssSeverities struct {
	CvssV2 *Cvss `json:"cvss_v2,omitempty"`
	CvssV3 *Cvss `json:"cvss_v3,omitempty"`
	CvssV4 *Cvss `json:"cvss_v4,omitempty"`
}

func (cvss *CvssSeverities) IsEmpty() bool {
	return cvss.CvssV2 == nil && cvss.CvssV3 == nil && cvss.CvssV4 == nil
}

type Cvss struct {
	VectorString string  `json:"vector_string"`
	Score        float64 `json:"score"`
}

func (cvss *Cvss) IsEmpty() bool {
	return cvss.VectorString == "" && cvss.Score == 0
}

type Credit struct {
	Type CreditType `json:"type"`
	User User       `json:"user"`
}

type User struct {
	Id           int    `json:"id"`
	Login        string `json:"login"`
	Email        string `json:"email"`
	NodeId       string `json:"node_id"`
	Type         string `json:"type"`
	UserViewType string `json:"user_view_type"`
	SiteAdmin    bool   `json:"site_admin"`
}

type (
	AdvisoryType string
	CreditType   string
)

const (
	AdvisoryType_Reviewed   AdvisoryType = "reviewed"
	AdvisoryType_Unreviewed AdvisoryType = "unreviewed"
	AdvisoryType_Malware    AdvisoryType = "malware"

	CreditType_Analyst              CreditType = "analyst"
	CreditType_Finder               CreditType = "finder"
	CreditType_Reporter             CreditType = "reporter"
	CreditType_Coordinator          CreditType = "coordinator"
	CreditType_RemediationDeveloper CreditType = "remediation_developer"
	CreditType_RemediationReviewer  CreditType = "remediation_reviewer"
	CreditType_RemediationVerifier  CreditType = "remediation_verifier"
	CreditType_Tool                 CreditType = "tool"
	CreditType_Sponsor              CreditType = "sponsor"
	CreditType_Other                CreditType = "other"
)
