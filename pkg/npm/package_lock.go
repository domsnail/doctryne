package npm

// ref: https://docs.npmjs.com/cli/v8/configuring-npm/package-lock-json
type PackageLock struct {
	Name            string `json:"name"`
	Version         string `json:"version"`
	LockfileVersion int    `json:"lockfileVersion"`
	Requires        bool   `json:"requires"`

	Packages map[string]*Dependency `json:"packages"`

	Funding Funding `json:"funding"`
}

type Funding struct {
	Type string `json:"type"`
	Url  string `json:"url"`
}

type Dependency struct {
	Version          string `json:"version"`
	Resolved         string `json:"resolved"`
	Integrity        string `json:"integrity"`
	License          string `json:"license"`
	HasInstallScript bool   `json:"hasInstallScript"`
	Optional         bool   `json:"optional"`
	Dev              bool   `json:"dev"`

	CPU []string `json:"cpu"`
	OS  []string `json:"os"`

	Dependencies         map[string]string `json:"dependencies"`
	PeerDependencies     map[string]string `json:"peerDependencies"`
	OptionalDependencies map[string]string `json:"optionalDependencies"`

	Engines map[string]string `json:"engines"`

	Bin map[string]string `json:"bin"`
}
