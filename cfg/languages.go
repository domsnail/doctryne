package cfg

type LanguagesConfig struct {
	JavaScript JavaScriptConfig `json:"javascript" yaml:"javascript"`
}

type JavaScriptConfig struct {
	CheckOptionalDependencies bool `json:"check_optional" yaml:"check_optional"`
	CheckDevDependencies      bool `json:"check_dev" yaml:"check_dev"`
}
