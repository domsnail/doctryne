package npm

type Stats struct {
	Package string `json:"package"`

	Downloads uint64 `json:"downloads"`

	Start string `json:"start"`
	End   string `json:"end"`
}

type PackageStatsPeriod string

const (
	PackageStatsPeriod_LastMonth PackageStatsPeriod = "last-month"
	PackageStatsPeriod_LastWeek  PackageStatsPeriod = "last-week"
	PackageStatsPeriod_LastYear  PackageStatsPeriod = "last-year"
)
