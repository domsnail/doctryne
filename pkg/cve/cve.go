package cve

type Cve struct {
	ID               string `json:"id"`
	SourceIdentifier string `json:"sourceIdentifier"`

	Published    Timestamp `json:"published"`
	LastModified Timestamp `json:"lastModified"`

	VulnStatus VulnStatus `json:"vulnStatus"` // https://nvd.nist.gov/vuln/vulnerability-status

	CveTags []CveTag `json:"cveTags"`

	Descriptions []Description `json:"descriptions"`
	Affected     []Affected    `json:"affected"`

	Metrics struct {
		CVSSMetricV40 []CVSSMetricsV40 `json:"cvssMetricV40"`
		CVSSMetricV31 []CVSSMetricsV31 `json:"cvssMetricV31"`
		CvssMetricV30 []CVSSMetricsV30 `json:"cvssMetricV30"`
		CVSSMetricV2  []CVSSMetricsV2  `json:"cvssMetricV2"`

		SSVCV203 []SSVCv203 `json:"ssvcV203"`
	} `json:"metrics"`

	Weaknesses     []Weakness      `json:"weaknesses"`
	Configurations []Configuration `json:"configurations"`
	References     []Reference     `json:"references"`

	CISA
}

type Description struct {
	Lang  string `json:"lang"`
	Value string `json:"value"`
}

type Affected struct {
	Source       string         `json:"source"`
	AffectedData []AffectedData `json:"affectedData"`
}

type AffectedData struct {
	Vendor  string `json:"vendor"`
	Product string `json:"product"`

	Versions []AffectedVersion `json:"versions"`
}

type AffectedVersion struct {
	Version     string `json:"version"`
	LessThan    string `json:"lessThan"`
	VersionType string `json:"versionType"`
	Status      string `json:"status"`

	Changes []struct {
		At     string `json:"at"`
		Status string `json:"status"`
	} `json:"changes"`
}

type Tag struct {
	SourceIdentifier string   `json:"sourceIdentifier"`
	Tags             []string `json:"tags"`
}

type SSVCv203 struct {
	Source string `json:"source"`

	SsvcData struct {
		ID        string    `json:"id"`
		Timestamp Timestamp `json:"timestamp"`

		Options []struct {
			Exploitation    string `json:"exploitation,omitempty"`
			Automatable     string `json:"automatable,omitempty"`
			TechnicalImpact string `json:"technicalImpact,omitempty"`
		} `json:"options"`

		Role    string `json:"role"`
		Version string `json:"version"`
	} `json:"ssvcData"`
}

type CVSSMetricsV2 struct {
	Source string `json:"source"`
	Type   string `json:"type"`

	CvssData struct {
		Version               string  `json:"version"`
		VectorString          string  `json:"vectorString"`
		BaseScore             float32 `json:"baseScore"`
		AccessVector          string  `json:"accessVector"`
		AccessComplexity      string  `json:"accessComplexity"`
		Authentication        string  `json:"authentication"`
		ConfidentialityImpact string  `json:"confidentialityImpact"`
		IntegrityImpact       string  `json:"integrityImpact"`
		AvailabilityImpact    string  `json:"availabilityImpact"`
	} `json:"cvssData"`

	BaseSeverity            string  `json:"baseSeverity"`
	ExploitabilityScore     float32 `json:"exploitabilityScore"`
	ImpactScore             float32 `json:"impactScore"`
	AcInsufInfo             bool    `json:"acInsufInfo"`
	ObtainAllPrivilege      bool    `json:"obtainAllPrivilege"`
	ObtainUserPrivilege     bool    `json:"obtainUserPrivilege"`
	ObtainOtherPrivilege    bool    `json:"obtainOtherPrivilege"`
	UserInteractionRequired bool    `json:"userInteractionRequired"`
}

type CVSSMetricsV30 struct {
	Source string `json:"source"`
	Type   string `json:"type"`

	ExploitabilityScore float32 `json:"exploitabilityScore"`
	ImpactScore         float32 `json:"impactScore"`

	CvssData struct {
		Version               string  `json:"version"`
		VectorString          string  `json:"vectorString"`
		BaseScore             float32 `json:"baseScore"`
		BaseSeverity          string  `json:"baseSeverity"`
		AttackVector          string  `json:"attackVector"`
		AttackComplexity      string  `json:"attackComplexity"`
		PrivilegesRequired    string  `json:"privilegesRequired"`
		UserInteraction       string  `json:"userInteraction"`
		Scope                 string  `json:"scope"`
		ConfidentialityImpact string  `json:"confidentialityImpact"`
		IntegrityImpact       string  `json:"integrityImpact"`
		AvailabilityImpact    string  `json:"availabilityImpact"`
	} `json:"cvssData"`
}

type CVSSMetricsV31 struct {
	Source string `json:"source"`
	Type   string `json:"type"`

	ExploitabilityScore float32 `json:"exploitabilityScore"`
	ImpactScore         float32 `json:"impactScore"`

	CvssData struct {
		Version               string  `json:"version"`
		VectorString          string  `json:"vectorString"`
		BaseScore             float32 `json:"baseScore"`
		BaseSeverity          string  `json:"baseSeverity"`
		AttackVector          string  `json:"attackVector"`
		AttackComplexity      string  `json:"attackComplexity"`
		PrivilegesRequired    string  `json:"privilegesRequired"`
		UserInteraction       string  `json:"userInteraction"`
		Scope                 string  `json:"scope"`
		ConfidentialityImpact string  `json:"confidentialityImpact"`
		IntegrityImpact       string  `json:"integrityImpact"`
		AvailabilityImpact    string  `json:"availabilityImpact"`
	} `json:"cvssData"`
}

type CVSSMetricsV40 struct {
	Source string `json:"source"`
	Type   string `json:"type"`

	CvssData struct {
		Version                           string  `json:"version"`
		VectorString                      string  `json:"vectorString"`
		BaseScore                         float32 `json:"baseScore"`
		BaseSeverity                      string  `json:"baseSeverity"`
		AttackVector                      string  `json:"attackVector"`
		AttackComplexity                  string  `json:"attackComplexity"`
		AttackRequirements                string  `json:"attackRequirements"`
		PrivilegesRequired                string  `json:"privilegesRequired"`
		UserInteraction                   string  `json:"userInteraction"`
		VulnConfidentialityImpact         string  `json:"vulnConfidentialityImpact"`
		VulnIntegrityImpact               string  `json:"vulnIntegrityImpact"`
		VulnAvailabilityImpact            string  `json:"vulnAvailabilityImpact"`
		SubConfidentialityImpact          string  `json:"subConfidentialityImpact"`
		SubIntegrityImpact                string  `json:"subIntegrityImpact"`
		SubAvailabilityImpact             string  `json:"subAvailabilityImpact"`
		ExploitMaturity                   string  `json:"exploitMaturity"`
		ConfidentialityRequirement        string  `json:"confidentialityRequirement"`
		IntegrityRequirement              string  `json:"integrityRequirement"`
		AvailabilityRequirement           string  `json:"availabilityRequirement"`
		ModifiedAttackVector              string  `json:"modifiedAttackVector"`
		ModifiedAttackComplexity          string  `json:"modifiedAttackComplexity"`
		ModifiedAttackRequirements        string  `json:"modifiedAttackRequirements"`
		ModifiedPrivilegesRequired        string  `json:"modifiedPrivilegesRequired"`
		ModifiedUserInteraction           string  `json:"modifiedUserInteraction"`
		ModifiedVulnConfidentialityImpact string  `json:"modifiedVulnConfidentialityImpact"`
		ModifiedVulnIntegrityImpact       string  `json:"modifiedVulnIntegrityImpact"`
		ModifiedVulnAvailabilityImpact    string  `json:"modifiedVulnAvailabilityImpact"`
		ModifiedSubConfidentialityImpact  string  `json:"modifiedSubConfidentialityImpact"`
		ModifiedSubIntegrityImpact        string  `json:"modifiedSubIntegrityImpact"`
		ModifiedSubAvailabilityImpact     string  `json:"modifiedSubAvailabilityImpact"`
		Safety                            string  `json:"Safety"`
		Automatable                       string  `json:"Automatable"`
		Recovery                          string  `json:"Recovery"`
		ValueDensity                      string  `json:"valueDensity"`
		VulnerabilityResponseEffort       string  `json:"vulnerabilityResponseEffort"`
		ProviderUrgency                   string  `json:"providerUrgency"`
	}
}

type Weakness struct {
	Source string `json:"source"`
	Type   string `json:"type"`

	Description []Description `json:"description"`
}

type Configuration struct {
	Operator string              `json:"operator,omitempty"`
	Nodes    []ConfigurationNode `json:"nodes"`
}

type ConfigurationNode struct {
	Operator string `json:"operator"`
	Negate   bool   `json:"negate"`

	CpeMatch []CPEMatch `json:"cpeMatch"`
}

type CPEMatch struct {
	Vulnerable            bool   `json:"vulnerable"`
	Criteria              string `json:"criteria"`
	VersionEndExcluding   string `json:"versionEndExcluding,omitempty"`
	MatchCriteriaId       string `json:"matchCriteriaId"`
	VersionStartIncluding string `json:"versionStartIncluding,omitempty"`
	VersionEndIncluding   string `json:"versionEndIncluding,omitempty"`
}

type CISA struct {
	CISAExploitAdd        string `json:"cisaExploitAdd"`
	CISAActionDue         string `json:"cisaActionDue"`
	CISARequiredAction    string `json:"cisaRequiredAction"`
	CISAVulnerabilityName string `json:"cisaVulnerabilityName"`
}

type Reference struct {
	Url    string   `json:"url"`
	Source string   `json:"source"`
	Tags   []string `json:"tags"`
}

type (
	VulnStatus string
	CveTag     string
)

const (
	VulnStatus_Received           VulnStatus = "Received"
	VulnStatus_AwaitingAnalysis   VulnStatus = "Awaiting Analysis"
	VulnStatus_UndergoingAnalysis VulnStatus = "Undergoing Analysis"
	VulnStatus_Analyzed           VulnStatus = "Analyzed"
	VulnStatus_Modified           VulnStatus = "Modified"
	VulnStatus_Deferred           VulnStatus = "Deferred" // The CVE is not currently scheduled for NVD enrichment efforts
	VulnStatus_Rejected           VulnStatus = "Rejected" // The CVE has been marked Rejected in the CVE List

	CveTag_Disputed                 CveTag = "disputed"
	CveTag_UnsupportedWhenAssigned  CveTag = "unsupported-when-assigned"
	CveTag_ExclusivelyHostedService CveTag = "exclusively-hosted-service"
)
