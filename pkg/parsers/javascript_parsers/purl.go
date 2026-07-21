package javascript_parsers

import (
	"strings"

	"github.com/domsnail/doctryne/pkg/npm"
	"github.com/package-url/packageurl-go"
)

var purlType = packageurl.TypeNPM

func getPackagePurl(pkg *npm.Package) *packageurl.PackageURL {
	if pkg == nil {
		return nil
	}

	var (
		name  = pkg.Name
		scope = ""
	)

	if strings.HasPrefix(pkg.Name, "@") {
		parts := strings.SplitN(pkg.Name, "/", 2)
		scope = parts[0]
		name = parts[1]
	}

	return packageurl.NewPackageURL(
		purlType,
		scope,
		name,
		pkg.Version,
		nil,
		"",
	)
}
