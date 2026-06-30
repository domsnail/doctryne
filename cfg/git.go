package cfg

type GitHistoryInspectionConfig struct {
	Disabled bool `json:"disabled" yaml:"disabled"`

	AlwaysFetch bool `json:"always_fetch" yaml:"always_fetch"`
	SaveToDisk  bool `json:"save_to_disk" yaml:"save_to_disk"`
	FullClone   bool `json:"full_clone" yaml:"full_clone"`
	MaxDepth    int  `json:"max_depth" yaml:"max_depth" env-default:"0"`

	Filepath string `json:"filepath" yaml:"filepath" env-default:"./downloaded"`
}
