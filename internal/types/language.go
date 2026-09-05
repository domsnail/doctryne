package types

type Language string

const (
	Language_Unspecified Language = "unspecified"

	Language_Golang     Language = "golang"
	Language_JavaScript Language = "javascript"
	Language_TypeScript Language = "typescript"
)

var Languages = []Language{
	Language_Unspecified,
	//
	Language_Golang,
	//
	Language_JavaScript,
	Language_TypeScript,
}
