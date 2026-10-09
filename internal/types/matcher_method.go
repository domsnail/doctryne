package types

type MatcherMethod string

const (
	MatcherMethod_Unspecified MatcherMethod = "unspecified"

	MatcherMethod_ExactPurl    MatcherMethod = "exact_purl"
	MatcherMethod_ExactPackage MatcherMethod = "exact_package"
	MatcherMethod_ExactProduct MatcherMethod = "exact_product"
)
