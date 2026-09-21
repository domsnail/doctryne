package types

type CvssVersion string

const (
	CvssVersion_v2  CvssVersion = "2"
	CvssVersion_v30 CvssVersion = "3.0"
	CvssVersion_v31 CvssVersion = "3.1"
	CvssVersion_v4  CvssVersion = "4.0"
)

func (c CvssVersion) Severity(score float32) CvssSeverity {
	if score < 0.0 || score > 10.0 {
		panic("cvss score out of range")
	}

	switch c {
	case CvssVersion_v2:
		return c.cvss2severity(score)
	case CvssVersion_v30, CvssVersion_v31, CvssVersion_v4:
		return c.cvss3severity(score)
	}

	return Severity_None
}

func (c CvssVersion) cvss2severity(score float32) CvssSeverity {
	switch {
	case score < 4.0:
		return Severity_Low
	case score < 7.0:
		return Severity_Medium
	default:
		return Severity_High
	}
}

func (c CvssVersion) cvss3severity(score float32) CvssSeverity {
	switch {
	case score == 0.0:
		return Severity_None
	case score < 4.0:
		return Severity_Low
	case score < 7.0:
		return Severity_Medium
	case score < 9.0:
		return Severity_High
	default:
		return Severity_Critical
	}
}
