package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/domsnail/doctryne/internal/entity"
	types2 "github.com/domsnail/doctryne/internal/types"
	"github.com/domsnail/doctryne/pkg/npm"
)

type NodePackageManagerServiceImpl struct {
	c npm.Client
}

func (service *NodePackageManagerServiceImpl) GetPackage(ctx context.Context, name, version string) (*entity.Package, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	} else if name == "" {
		return nil, errors.New("name is required")
	}

	npmPkg, raw, err := service.c.GetPackage(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch package '%s': %w", name, err)
	} else if npmPkg == nil {
		return nil, nil
	}

	npmStats, err := service.c.GetPackageStats(ctx, name, time.Hour)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch package stats '%s': %w", name, err)
	} else if npmStats == nil {
		return nil, fmt.Errorf("no package stats found for package '%s'", name)
	}

	slog.DebugContext(ctx, "converting npm package...", slog.String("package_name", npmPkg.Name))
	var pkg = getPackage(npmPkg)
	//pkg.Stats = getPackageStats(npmStats)
	pkg.Raw = raw

	return pkg, nil
}

func getPackage(n *npm.Package) *entity.Package {
	pkg := entity.Package{
		Name:      n.Name,
		Ecosystem: types2.Ecosystem_NPM,
		Language:  types2.Language_JavaScript,
		RegistryMetadata: &entity.RegistryMetadata{
			RegistryID: n.ID,
			//Git:        n.GetGitURL(),
			License:     n.License,
			Description: n.Description,
			Homepage:    n.Homepage,
			Keywords:    n.Keywords,
		},
		//Contributors: getPackageContributors(n),
		//PublishedAt:  n.GetCreatedAt(),
		//ModifiedAt:   n.GetModifiedAt(),
	}

	return &pkg
}

func getPackageContributors(n *npm.Package) *entity.AffiliatedDevelopers {
	var contrib = entity.AffiliatedDevelopers{
		Authors:      make([]*entity.Developer, 0),
		Contributors: make([]*entity.Developer, 0),
		Maintainers:  make([]*entity.Developer, 0),
	}

	contrib.Authors = append(contrib.Authors, &entity.Developer{
		Name:   n.Author.Name,
		Emails: []string{n.Author.Email},
	})

	for _, item := range n.Contributors {
		contrib.Authors = append(contrib.Contributors, &entity.Developer{
			Name:   item.Name,
			Emails: []string{item.Email},
		})
	}

	for _, item := range n.Maintainers {
		contrib.Authors = append(contrib.Maintainers, &entity.Developer{
			Name:   item.Name,
			Emails: []string{item.Email},
		})
	}

	for _, version := range n.Versions {
		contrib.Authors = append(contrib.Authors, &entity.Developer{
			Name:   version.Author.Name,
			Emails: []string{version.Author.Email},
		})

		for _, item := range version.Contributors {
			contrib.Authors = append(contrib.Contributors, &entity.Developer{
				Name:   item.Name,
				Emails: []string{item.Email},
			})
		}

		for _, item := range version.Maintainers {
			contrib.Authors = append(contrib.Maintainers, &entity.Developer{
				Name:   item.Name,
				Emails: []string{item.Email},
			})
		}
	}

	return &contrib
}

func getPackageStats(n *npm.Stats) *entity.PackageStats {
	var stats = entity.PackageStats{
		Downloads: entity.PackageDownloads{
			Downloads: n.Downloads,
		},
	}

	stats.Downloads.StartAt, _ = time.Parse("2006-01-02", n.Start)
	stats.Downloads.EndAt, _ = time.Parse("2006-01-02", n.End)

	return &stats
}
