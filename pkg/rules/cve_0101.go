package rules

// Rule_CVE_0101 max CVSS rating of any CVE
type Rule_CVE_0101 struct {
	NumericRangeRule `yaml:",inline"`
}
