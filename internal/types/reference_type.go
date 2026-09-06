package types

type ReferenceType string

const (
	ReferenceType_Advisory           ReferenceType = "advisory"
	ReferenceType_Article            ReferenceType = "article"
	ReferenceType_Exploit            ReferenceType = "exploit"
	ReferenceType_Fix                ReferenceType = "fix"
	ReferenceType_GovernmentResource ReferenceType = "government_resource"
	ReferenceType_MediaCoverage      ReferenceType = "media_coverage"
	ReferenceType_Mitigation         ReferenceType = "mitigation"
	ReferenceType_Patch              ReferenceType = "patch"
	ReferenceType_Product            ReferenceType = "product"
	ReferenceType_ReleaseNotes       ReferenceType = "release_notes"
	ReferenceType_VendorAdvisory     ReferenceType = "vendor_advisory"

	ReferenceType_Other ReferenceType = "other"
)
