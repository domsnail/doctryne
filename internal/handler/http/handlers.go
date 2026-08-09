package http

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/domsnail/doctryne/internal/entity"
	"github.com/domsnail/doctryne/pkg/types"
	"github.com/domsnail/doctryne/web/templates"
	"github.com/google/uuid"
)

func (h *Handler) handleManifestUpload(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	opts, err := parseInspectionOptionsForm(ctx, r)
	if err != nil {
		slog.WarnContext(ctx, "failed to parse options from multipart form", slog.String("error", err.Error()))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	inspection, err := h.service.InitInspection(ctx, opts)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = h.service.InspectManifests(ctx, inspection)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = h.service.InspectPackages(ctx, inspection)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = h.service.SaveInspection(ctx, inspection)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// https://htmx.org/headers/hx-redirect/
	w.Header().Set("HX-Redirect", fmt.Sprintf("/inspections/%s/revisions/%d", inspection.UUID.String(), inspection.Revision))
	w.WriteHeader(http.StatusOK)
	return
}

func (h *Handler) handleInspectionPage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	inspectionUUID := r.PathValue("uuid")
	inspectionRevision := r.PathValue("revision")

	rev, err := strconv.ParseUint(inspectionRevision, 10, 32)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	uid, err := uuid.Parse(inspectionUUID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	inspection, err := h.service.LoadInspection(ctx, uid, uint32(rev))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	err = templates.Inspection_page(inspection).Render(ctx, w)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	return
}

func parseInspectionOptionsForm(ctx context.Context, req *http.Request) (*entity.InspectionOptions, error) {
	var opts = entity.InspectionOptions{
		ScanType: types.ScanType_Binary,
		Mode:     types.InspectionMode_Direct,
	}

	err := req.ParseMultipartForm(1024 * 1024)
	if err != nil {
		return nil, err
	}

	file, header, err := req.FormFile("manifest-file")
	if err != nil {
		return nil, err
	} else if file == nil {
		return nil, errors.New("no manifest file provided")
	}

	slog.DebugContext(ctx, "uploaded manifest file",
		slog.String("name", header.Filename),
		slog.Int64("size", header.Size),
	)

	opts.Manifest = file
	opts.ManifestType = types.ManifestType(req.FormValue("manifest-type"))

	opts.DeepRepositoryInspection = req.FormValue("deep-repository-inspection") == "on"
	opts.ExtractFullContributorInfo = req.FormValue("extract-full-contibutor-info") == "on"
	opts.ExtractFullOrganizationInfo = req.FormValue("extract-full-organization-info") == "on"

	return &opts, nil
}
