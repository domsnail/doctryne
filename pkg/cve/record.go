package cve

import (
	"errors"
	"time"

	"github.com/domsnail/doctryne/pkg/cve/cve_v52"
)

type RecordsQueryOptions struct {
	ModificationDateStart *time.Time // lastModStartDate
	ModificationDateEnd   *time.Time // lastModEndDate (max date range is 120 days)

	KevDateStart *time.Time // kevStartDate
	KevDateEnd   *time.Time // kevEndDate

	PublishedDateStart *time.Time // pubStartDate
	PublishedDateEnd   *time.Time // pubEndDate

	VirtualMatchString string
	KeywordExactMatch  string
	KeywordSearch      []string
	SourceIdentifier   string

	CpeName string
	CveIDs  []string
	CveTag  string

	CweID string

	CvssV2Metrics, CvssV3Metrics, CvssV4Metrics string

	CvssV2Severity cve_v52.CVSSv2_Severity
	CvssV3Severity cve_v52.CVSSv3_Severity
	CvssV4Severity cve_v52.CVSSv4_Severity

	EventName     string
	HasCertAlerts bool
	HasKEV        bool // https://www.cisa.gov/known-exploited-vulnerabilities-catalog
	HasOVAL       bool

	IsVulnerable bool // requires CpeName
	NoRejected   bool

	versionEnd, versionEndType     string
	versionStart, versionStartType string
}

func (opts RecordsQueryOptions) Validate() error {
	if opts.ModificationDateStart != nil && opts.ModificationDateEnd != nil {
		if opts.ModificationDateEnd.Before(*opts.ModificationDateStart) {
			return errors.New("modification date start must be before modification date end")
		}

		if opts.ModificationDateEnd.Sub(*opts.ModificationDateStart) > time.Hour*24*120 {
			return errors.New("modification date delta must be less than 120 days, see: https://nvd.nist.gov/developers/vulnerabilities")
		}
	}

	if opts.PublishedDateStart != nil && opts.PublishedDateEnd != nil {
		if opts.ModificationDateEnd.Before(*opts.ModificationDateStart) {
			return errors.New("published date start must be before published date end")
		}

		if opts.ModificationDateEnd.Sub(*opts.ModificationDateStart) > time.Hour*24*120 {
			return errors.New("published date delta must be less than 120 days, see: https://nvd.nist.gov/developers/vulnerabilities")
		}
	}

	if opts.KevDateStart != nil && opts.KevDateEnd != nil {
		if opts.KevDateStart.Before(*opts.KevDateEnd) {
			return errors.New("kev date start must be before modification date end")
		}
	}

	if opts.IsVulnerable && opts.CpeName == "" {
		return errors.New("cpe_name must not be empty when using is_vulnerable")
	}

	return nil
}

type RecordsQueryResponse struct {
	Format    string    `json:"format"`
	Version   string    `json:"version"`
	Timestamp Timestamp `json:"timestamp"`

	ResultsPerPage uint32 `json:"resultsPerPage"`
	StartIndex     uint32 `json:"startIndex"`
	TotalResults   uint32 `json:"totalResults"`

	Vulnerabilities []Vulnerability `json:"vulnerabilities"`
}

type Vulnerability struct {
	CVE Cve `json:"cve"`
}
