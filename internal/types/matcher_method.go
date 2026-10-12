package types

type MatcherMethod string

const (
	MatcherMethod_Unspecified MatcherMethod = "unspecified"

	MatcherMethod_ByPurl    MatcherMethod = "match_purl"
	MatcherMethod_ByPackage MatcherMethod = "match_package"
)
