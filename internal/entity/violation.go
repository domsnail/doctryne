package entity

import (
	"github.com/domsnail/doctryne/internal/types"
)

type Violation struct {
	// Object describes what violates a rule
	Object string

	// Scope describes where Object was found (in emails, in developers or else)
	Scope string

	Rule    string
	Details string

	CvssScore float32
	Severity  types.CvssSeverity
	CWEId     string
}
