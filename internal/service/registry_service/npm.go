package registry_service

import (
	"context"
	"log/slog"
	"time"

	"github.com/domsnail/doctryne/internal/entity"
)

func (service RegistryServiceImpl) queryNPM(ctx context.Context, pkg *entity.Package) error {
	fetched, raw, err := service.npm.GetPackage(ctx, pkg.Name)
	if err != nil {
		return err
	}

	// save for future use (todo: change to interface)
	pkg.Raw = raw

	contrib := entity.AffiliatedDevelopers{
		Authors: []entity.Developer{
			{Name: fetched.Author.Name, Emails: []string{fetched.Author.Email}},
		},
		Contributors: make([]entity.Developer, len(fetched.Contributors)),
		Maintainers:  make([]entity.Developer, len(fetched.Maintainers)),
	}

	for i := range fetched.Contributors {
		contrib.Contributors[i] = entity.Developer{
			Name:   fetched.Contributors[i].Name,
			Emails: []string{fetched.Contributors[i].Email},
		}
	}

	for i := range fetched.Maintainers {
		contrib.Maintainers[i] = entity.Developer{
			Name:   fetched.Maintainers[i].Name,
			Emails: []string{fetched.Maintainers[i].Email},
		}
	}

	// convert from npm to package
	pkg.RegistryMetadata = &entity.RegistryMetadata{
		RegistryID:   fetched.ID,
		Description:  fetched.Description,
		Homepage:     fetched.Homepage,
		Keywords:     fetched.Keywords,
		License:      fetched.License,
		Readme:       fetched.Readme,
		Git:          fetched.GetGitURL(),
		Contributors: &contrib,
		Stats:        nil,
		IsPrivate:    fetched.Private,
		PublishedAt:  fetched.GetCreatedAt(),
		ModifiedAt:   fetched.GetModifiedAt(),
	}

	stats, err := service.npm.GetPackageStats(ctx, pkg.Name, service.statsPeriod)
	if err != nil {
		slog.WarnContext(ctx, "failed to query npm package stats",
			slog.String("package_name", pkg.Name),
			slog.String("error", err.Error()),
		)
	} else {
		pkg.RegistryMetadata.Stats = &entity.PackageStats{
			Downloads: entity.PackageDownloads{
				Downloads: stats.Downloads,
			}}

		pkg.RegistryMetadata.Stats.Downloads.StartAt, _ = time.Parse("2006-01-02", stats.Start)
		pkg.RegistryMetadata.Stats.Downloads.EndAt, _ = time.Parse("2006-01-02", stats.End)
	}

	return nil
}
