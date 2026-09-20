package types

type CvssVersion string

const (
	CvssVersion_v2  CvssVersion = "2"
	CvssVersion_v30 CvssVersion = "3.0"
	CvssVersion_v31 CvssVersion = "3.1"
	CvssVersion_v4  CvssVersion = "4.0"
)

//func (c CvssVersion) Severity(score float32) CvssSeverity {
//	if score == 0.0 {
//		return Severity_None
//	}
//
//	switch c {
//	case CvssVersion_v2:
//
//	case CvssVersion_v30, CvssVersion_v31, CvssVersion_v4:
//		if score <= 6.9 {
//
//			if score <= 3.9 {
//				return Severity_Low
//			}
//
//			return Severity_Medium
//		}
//
//		if score > 8.9 {
//			return Severity_Critical
//		}
//
//		return Severity_High
//	}
//
//	return Severity_None
//}
