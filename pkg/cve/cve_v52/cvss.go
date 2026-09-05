package cve_v52

// CVSS metrics
// ref: https://cveproject.github.io/cve-schema/schema/docs/#oneOf_i0_containers_cna_metrics
type CVSS struct {
	Version string `json:"version" `

	BaseScore    float32  `json:"baseScore"`
	BaseSeverity Severity `json:"baseSeverity" `

	VectorString string `json:"vectorString" `
}

type CVSSv40 struct {
	CVSS
}

type CVSSv31 struct {
	CVSS
}

type CVSSv30 struct {
	CVSS
}

type CVSSv20 struct {
	CVSS
}

type (
	Severity string

	CVSSv2_Severity string
	CVSSv3_Severity string
	CVSSv4_Severity string
)

const (
	Severity_None     Severity = "NONE"
	Severity_Low      Severity = "LOW"
	Severity_Medium   Severity = "MEDIUM"
	Severity_High     Severity = "HIGH"
	Severity_Critical Severity = "CRITICAL"

	CVSSv2_Severity_Low    CVSSv2_Severity = "LOW"
	CVSSv2_Severity_Medium CVSSv2_Severity = "MEDIUM"
	CVSSv2_Severity_High   CVSSv2_Severity = "HIGH"

	CVSSv3_Severity_Low      CVSSv3_Severity = "LOW"
	CVSSv3_Severity_Medium   CVSSv3_Severity = "MEDIUM"
	CVSSv3_Severity_High     CVSSv3_Severity = "HIGH"
	CVSSv3_Severity_Critical CVSSv3_Severity = "CRITICAL"

	CVSSv4_Severity_Low      CVSSv4_Severity = "LOW"
	CVSSv4_Severity_Medium   CVSSv4_Severity = "MEDIUM"
	CVSSv4_Severity_High     CVSSv4_Severity = "HIGH"
	CVSSv4_Severity_Critical CVSSv4_Severity = "CRITICAL"
)
