package grpc

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"

	"github.com/domsnail/doctryne/api/gen/go/inspection/v1"
	utils_v1 "github.com/domsnail/doctryne/api/gen/go/utils"
	"github.com/domsnail/doctryne/internal/entity"
	"github.com/domsnail/doctryne/internal/service"
	types2 "github.com/domsnail/doctryne/internal/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Handler struct {
	service service.IInspectionService
}

func (handler *Handler) GetInspectionByUUID(ctx context.Context, uid *utils_v1.UUID) (*inspection_v1.Inspection, error) {
	//TODO implement me
	panic("implement me")
}

func NewHandler(service service.IInspectionService) *Handler {
	return &Handler{service: service}
}

func (handler *Handler) Inspect(ctx context.Context, opts *inspection_v1.InspectionOptions) (*inspection_v1.Inspection, error) {
	if !opts.HasManifest() {
		return nil, status.Error(codes.InvalidArgument, "manifest file is not specified")
	} else if !opts.HasManifestType() {
		return nil, status.Error(codes.InvalidArgument, "manifest file type is not specified")
	} else if opts.GetScanType() == inspection_v1.ScanType_SCAN_TYPE_DIRPATH || opts.GetScanType() == inspection_v1.ScanType_SCAN_TYPE_FILEPATH {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("scan type '%s' not supported in server mode", opts.GetScanType().String()))
	}

	inspection, err := handler.service.InitInspection(ctx, &entity.InspectionOptions{
		ScanType: types2.ScanTypes_Enums[int32(opts.GetScanType())],
		Mode:     types2.InspectionModes_Enums[int32(opts.GetMode())],
		//
		Manifest:     bytes.NewReader(opts.GetManifest()),
		ManifestType: types2.ManifestType(opts.GetManifestType()),
		Lockfile:     bytes.NewReader(opts.GetLockfile()),
		LockfileType: types2.ManifestType(opts.GetLockfileType()),
		//
		ExtractFullContributorInfo: opts.GetLoadUserProfile(),
		// todo: add DeepRepositoryInspection
	})

	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "failed to init inspection: "+err.Error())
	}

	slog.InfoContext(ctx, "starting new inspection...",
		slog.String("inspection_uuid", inspection.UUID.String()),
		slog.String("uploaded_by", inspection.UploadedBy),
	)

	err = handler.service.InspectManifests(ctx, inspection)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "failed to inspect manifests: "+err.Error())
	}

	err = handler.service.InspectPackages(ctx, inspection)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "failed to inspect manifest packages: "+err.Error())
	}

	err = handler.service.InspectRepositories(ctx, inspection)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "failed to inspect package repositories: "+err.Error())
	}

	err = handler.service.InspectDevelopersAndOrganizations(ctx, inspection)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "failed to inspect manifest package developers: "+err.Error())
	}

	slog.DebugContext(ctx, "finished inspection, gathering security violations...",
		slog.String("inspection_uuid", inspection.UUID.String()),
	)

	violations, err := handler.service.CollectViolations(ctx, inspection)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "failed to collect manifest violations: "+err.Error())
	}

	slog.InfoContext(ctx, "finished inspection",
		slog.String("inspection_uuid", inspection.UUID.String()),
		slog.Int("violations_count", len(violations)),
	)

	// todo: policy match

	return nil, status.Error(codes.Unimplemented, "not implemented")
}
