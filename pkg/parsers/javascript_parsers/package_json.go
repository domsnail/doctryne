package javascript_parsers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"maps"
	"net/url"
	"strings"

	"github.com/domsnail/doctryne/cfg"
	"github.com/domsnail/doctryne/internal/entity"
	types2 "github.com/domsnail/doctryne/internal/types"
	"github.com/domsnail/doctryne/pkg/npm"
)

func (p *Parser) ParseManifest() (*entity.Package, error) {
	if p.ctx.Err() != nil {
		return nil, p.ctx.Err()
	} else if p.file == nil {
		return nil, errors.New("manifest file is nil")
	}

	var (
		f npm.Package // todo: this needs testing
		l npm.PackageLock
	)

	if err := json.NewDecoder(p.file).Decode(&f); err != nil {
		return nil, fmt.Errorf("failed to unmarshal manifest file: %w", err)
	}

	if p.lockfile != nil {
		if err := json.NewDecoder(p.lockfile).Decode(&l); err != nil {
			return nil, fmt.Errorf("failed to unmarshal lockfile: %w", err)
		}

		return convertWithLockfile(p.ctx, f, l)
	}

	return convert(p.ctx, f)
}

// NOTE: convert function currently builds flat tree with only 1 layer of dependencies.
// Dependency tree is not built, all descendants will be places at the top layer
func convert(ctx context.Context, p npm.Package) (*entity.Package, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	var (
		eco  = types2.Ecosystem_NPM
		lang = types2.Language_JavaScript // todo: how to check vs TypeScript?
	)

	var rootPkg = entity.Package{
		Name:       p.Name,
		Version:    p.Version,
		Ecosystem:  eco,
		Language:   lang,
		PackageURL: getPackagePurl(&p),
		//Resolved:   nil, // todo: get from package-lock.json
		//Registry:   "",
		//Integrity:  "",
		Labels: []types2.Label{types2.Label_Root},
		RegistryMetadata: &entity.RegistryMetadata{
			RegistryID:   p.ID,
			Description:  p.Description,
			Homepage:     p.Homepage,
			Keywords:     p.Keywords,
			License:      p.License,
			Contributors: new(entity.AffiliatedDevelopers),
			Git:          p.GetGitURL(),
			PublishedAt:  p.GetCreatedAt(),
			ModifiedAt:   p.GetModifiedAt(),
		},
		Dependencies: make([]*entity.Package, len(p.Dependencies)),
	}

	if p.Author.Name != "" {
		var developer entity.Developer
		developer.Username = p.Author.Name

		if p.Author.Email != "" {
			developer.Emails = append(developer.Emails, p.Author.Email)
		}

		rootPkg.RegistryMetadata.Contributors.Authors = append(rootPkg.RegistryMetadata.Contributors.Authors, &developer)
	}

	var counter = 0
	for dep, ver := range p.Dependencies {
		rootPkg.Dependencies[counter] = &entity.Package{
			Name:      dep,
			Version:   ver,
			Ecosystem: eco,
			Language:  lang,
			//Resolved:     nil, // todo: get from package-lock.json
			//Registry:     "",
			//Integrity:    "",
			//License:      "",
			Dependencies: nil, // todo: read method note
		}

		counter++
	}

	if cfg.GlobalConfig.Languages.JavaScript.CheckOptionalDependencies && len(p.OptionalDependencies) > 0 {
		for dep, ver := range p.OptionalDependencies {
			rootPkg.Dependencies = append(rootPkg.Dependencies, &entity.Package{
				Name:      dep,
				Version:   ver,
				Ecosystem: eco,
				Language:  lang,
				//Resolved:     nil, // todo: get from package-lock.json
				//Registry:     "",
				//Integrity:    "",
				//License:      "",
				IsOptional:   true,
				Dependencies: nil, // todo: read method note
			})
		}
	}

	if cfg.GlobalConfig.Languages.JavaScript.CheckDevDependencies && len(p.DevDependencies) > 0 {
		for dep, ver := range p.DevDependencies {
			rootPkg.Dependencies = append(rootPkg.Dependencies, &entity.Package{
				Name:      dep,
				Version:   ver,
				Ecosystem: eco,
				Language:  lang,
				//Resolved:     nil, // todo: get from package-lock.json
				//Registry:     "",
				//Integrity:    "",
				//License:      "",
				IsDev:        true,
				Dependencies: nil, // todo: read method note
			})
		}
	}

	return &rootPkg, nil
}

func convertWithLockfile(ctx context.Context, p npm.Package, l npm.PackageLock) (pkg *entity.Package, err error) {
	if p.Version != l.Version {
		slog.WarnContext(ctx, "manifest and lockfile versions do not match")
	}

	var (
		eco  = types2.Ecosystem_NPM
		lang = types2.Language_JavaScript // todo: how to check vs TypeScript?
	)

	// removing root package from dependency list
	delete(l.Packages, "")

	// remove dev and optional dependencies if needed
	slog.Debug("dropping dependencies from lockfile...",
		slog.Bool("remove_dev", cfg.GlobalConfig.Languages.JavaScript.CheckDevDependencies),
		slog.Bool("remove_optional", cfg.GlobalConfig.Languages.JavaScript.CheckOptionalDependencies),
	)

	maps.DeleteFunc(l.Packages, func(s string, d *npm.Dependency) bool {
		if d.Dev {
			return !cfg.GlobalConfig.Languages.JavaScript.CheckDevDependencies // todo: switch to scan options
		}

		if d.Optional {
			return !cfg.GlobalConfig.Languages.JavaScript.CheckOptionalDependencies // todo: switch to scan options
		}

		return false
	})

	pkg = &entity.Package{
		Name:      p.Name,
		Version:   p.Version,
		Ecosystem: eco,
		Language:  lang,
		Labels:    []types2.Label{types2.Label_Root},
		RegistryMetadata: &entity.RegistryMetadata{
			RegistryID:  p.ID,
			Description: p.Description,
			Homepage:    p.Homepage,
			Keywords:    p.Keywords,
			License:     p.License,
			Git:         p.GetGitURL(),
			PublishedAt: p.GetCreatedAt(),
			ModifiedAt:  p.GetModifiedAt(),
		},
		Dependencies: make([]*entity.Package, len(l.Packages)),
	}

	var counter = 0
	for key, depPkg := range l.Packages {
		v := entity.Package{
			Name:       strings.TrimLeft(key, "node_modules/"),
			Version:    depPkg.Version,
			Ecosystem:  eco,
			Language:   lang,
			Integrity:  depPkg.Integrity, // todo: define algorithm
			IsDev:      depPkg.Dev,
			IsOptional: depPkg.Optional,
		}

		v.Resolved, err = url.Parse(depPkg.Resolved)
		if err != nil {
			slog.WarnContext(ctx, "could not parse resolved url", slog.String("error", err.Error()))
		} else {
			v.Registry = v.Resolved.Hostname()
		}

		pkg.Dependencies[counter] = &v
		counter++
	}

	return pkg, nil
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
