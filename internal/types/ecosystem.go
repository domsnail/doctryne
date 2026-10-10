package types

type Ecosystem string

const (
	Ecosystem_None        Ecosystem = ""
	Ecosystem_Unspecified Ecosystem = "unspecified"

	Ecosystem_Cargo           Ecosystem = "cargo"
	Ecosystem_Composer        Ecosystem = "composer"
	Ecosystem_Conan           Ecosystem = "conan"
	Ecosystem_Cpan            Ecosystem = "cpan"
	Ecosystem_Deb             Ecosystem = "deb"
	Ecosystem_Docker          Ecosystem = "docker"
	Ecosystem_Erlang          Ecosystem = "erlang"
	Ecosystem_Gem             Ecosystem = "gem"
	Ecosystem_Generic         Ecosystem = "generic"
	Ecosystem_Github          Ecosystem = "github"
	Ecosystem_GithubActions   Ecosystem = "actions"
	Ecosystem_Go              Ecosystem = "go"
	Ecosystem_Golang          Ecosystem = "golang"
	Ecosystem_Hex             Ecosystem = "hex"
	Ecosystem_Maven           Ecosystem = "maven"
	Ecosystem_Npm             Ecosystem = "npm"
	Ecosystem_Nuget           Ecosystem = "nuget"
	Ecosystem_Oci             Ecosystem = "oci"
	Ecosystem_Opam            Ecosystem = "opam"
	Ecosystem_Otp             Ecosystem = "otp"
	Ecosystem_Pip             Ecosystem = "pip"
	Ecosystem_Pub             Ecosystem = "pub"
	Ecosystem_Pypi            Ecosystem = "pypi"
	Ecosystem_Rubygems        Ecosystem = "rubygems"
	Ecosystem_Rust            Ecosystem = "rust"
	Ecosystem_SoftwareId      Ecosystem = "software-id"
	Ecosystem_Swift           Ecosystem = "swift"
	Ecosystem_WordpressPlugin Ecosystem = "wordpress-plugin"
)
