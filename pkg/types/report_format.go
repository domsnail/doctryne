package types

type ReportFormat string

const (
	ReportFormat_Unspecified ReportFormat = "unspecified"

	ReportFormat_JSON        = "json"
	ReportFormat_TextTable   = "text_table"
	ReportFormat_CycloneDX   = "cyclonedx"
	ReportFormat_HexwayVampy = "hexway_vampy"
)

var ReportFormats = []ReportFormat{
	ReportFormat_Unspecified,
	ReportFormat_JSON,
	ReportFormat_TextTable,
	ReportFormat_CycloneDX,
	ReportFormat_HexwayVampy,
}
