package inspect_service

import (
	"context"
	"errors"
	"log/slog"

	"github.com/domsnail/doctryne/internal/entity"
	"golang.org/x/sync/errgroup"
)

func (service *InspectionService) InspectManifests(ctx context.Context, inspection *entity.Inspection) error {
	if ctx.Err() != nil {
		return ctx.Err()
	} else if inspection == nil || len(inspection.Manifests) == 0 {
		return errors.New("no manifests found in inspection")
	}

	slog.DebugContext(ctx, "starting inspection manifests processing...",
		slog.Int("total_manifests", len(inspection.Manifests)),
	)

	group, groupCtx := errgroup.WithContext(ctx)
	for _, manifest := range inspection.Manifests {
		group.Go(func() error {
			slog.DebugContext(groupCtx, "starting manifest processing...",
				slog.String("filename", manifest.Metadata.Filename),
				slog.String("manifest_type", string(manifest.Type)),
			)

			err := service.manifests.ProcessManifest(groupCtx, manifest)
			if err != nil {
				slog.ErrorContext(groupCtx, "failed to process manifest",
					slog.String("filename", manifest.Metadata.Filename),
					slog.String("manifest_type", string(manifest.Type)),
					slog.String("error", err.Error()),
				)

				return err
			}

			slog.DebugContext(groupCtx, "manifest processed successfully",
				slog.String("filename", manifest.Metadata.Filename),
				slog.String("manifest_type", string(manifest.Type)),
				slog.Int("discovered_packages", len(manifest.DiscoveredPackages)),
			)

			return nil
		})
	}

	err := group.Wait()
	if err != nil {
		slog.WarnContext(ctx, "failed to process manifests",
			slog.String("error", err.Error()),
		)

		return err
	}

	slog.DebugContext(ctx, "manifest processing completed",
		slog.Int("manifests_processed", len(inspection.Manifests)),
	)

	return nil
}
