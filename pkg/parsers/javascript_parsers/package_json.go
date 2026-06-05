package javascript_parsers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/domsnail/doctryne/cfg"
	"github.com/domsnail/doctryne/internal/entity"
	"github.com/domsnail/doctryne/pkg/npm"
	"github.com/domsnail/doctryne/pkg/types"
)

func (p *Parser) ParseManifest(ctx context.Context, manifest *entity.Manifest) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	var f npm.Package // todo: this needs testing

	if err := json.NewDecoder(bytes.NewBuffer(manifest.Raw)).Decode(&f); err != nil {
		return fmt.Errorf("failed to unmarshal manifest file: %w", err)
	}

	manifest.AddPackage(convert(f))
	return nil
}

// NOTE: convert function currently builds flat tree with only 1 layer of dependencies.
// Dependency tree is not built, all descendants will be places at the top layer
func convert(p npm.Package) *entity.Package {
	var (
		eco  = types.Ecosystem_NPM
		lang = types.Language_JavaScript // todo: how to check vs TypeScript?
	)

	var topPackage = entity.Package{
		Name:       p.Name,
		Version:    p.Version,
		Ecosystem:  eco,
		Language:   lang,
		Git:        p.GetGitURL(),
		RegistryID: p.ID,
		//Resolved:   nil, // todo: get from package-lock.json
		//Registry:   "",
		//Integrity:  "",
		License: p.License,
		Metadata: &entity.PackageMetadata{
			Description: p.Description,
			Homepage:    p.Homepage,
			Keywords:    p.Keywords,
		},
		PublishedAt:  p.GetCreatedAt(),
		ModifiedAt:   p.GetModifiedAt(),
		Dependencies: make([]*entity.Package, len(p.Dependencies)),
	}

	if cfg.GlobalConfig.Languages.JavaScript.CheckOptionalDependencies && len(p.OptionalDependencies) > 0 {
		for dep, ver := range p.OptionalDependencies {
			topPackage.Dependencies = append(topPackage.Dependencies, &entity.Package{
				Name:      dep,
				Version:   ver,
				Ecosystem: eco,
				Language:  lang,
				//Resolved:     nil, // todo: get from package-lock.json
				//Registry:     "",
				//Integrity:    "",
				//License:      "",
				Dependencies: nil, // todo: read method note
			})
		}
	}

	if cfg.GlobalConfig.Languages.JavaScript.CheckDevDependencies && len(p.DevDependencies) > 0 {
		for dep, ver := range p.DevDependencies {
			topPackage.Dependencies = append(topPackage.Dependencies, &entity.Package{
				Name:      dep,
				Version:   ver,
				Ecosystem: eco,
				Language:  lang,
				//Resolved:     nil, // todo: get from package-lock.json
				//Registry:     "",
				//Integrity:    "",
				//License:      "",
				Dependencies: nil, // todo: read method note
			})
		}
	}

	return &topPackage
}

// === Claimed from Trivy
// === Source: https://github.com/aquasecurity/trivy/blob/main/pkg/dependency/parser/nodejs/packagejson/types.go

type packageJson struct {
	Name                 string            `json:"name"`
	Version              string            `json:"version"`
	License              License           `json:"license"`
	Dependencies         map[string]string `json:"dependencies"`
	OptionalDependencies map[string]string `json:"optionalDependencies"`
	DevDependencies      map[string]string `json:"devDependencies"`
	Workspaces           any               `json:"workspaces"`
}

// License represents the npm "license" field, which historically
// supports several shapes:
//   - string:                 "MIT"
//   - object (legacy):        {"type": "MIT", "url": "..."}
//   - array of objects (legacy): [{"type": "MIT", ...}, {"type": "Apache-2.0", ...}]
//
// See https://docs.npmjs.com/cli/v11/configuring-npm/package-json#license
type License struct {
	names []string
}

type licenseObject struct {
	Type string `json:"type"`
}

func (l *License) UnmarshalJSON(data []byte) error {
	// "MIT"
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		if s != "" {
			l.names = []string{s}
		}

		return nil
	}

	// {"type": "MIT", ...}
	var obj licenseObject
	if err := json.Unmarshal(data, &obj); err == nil && obj.Type != "" {
		l.names = []string{obj.Type}
		return nil
	}

	// [{"type": "MIT", ...}, ...]
	var arr []licenseObject
	if err := json.Unmarshal(data, &arr); err == nil {
		for _, o := range arr {
			if o.Type != "" {
				l.names = append(l.names, o.Type)
			}
		}
		return nil
	}

	// Unknown shape — return empty list instead of failing the whole file.
	return nil
}

// Names returns the list of license names extracted from the field.
func (l License) Names() []string {
	return l.names
}
