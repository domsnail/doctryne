package http

import (
	"log/slog"
	"net/http"

	"github.com/domsnail/doctryne/internal/entity"
)

func (h *Handler) handleManifestUpload(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseMultipartForm(1024 * 1024)
	if err != nil {
		slog.WarnContext(ctx, "failed to parse multipart form", slog.String("error", err.Error()))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("manifest_file")
	if err != nil {
		slog.WarnContext(ctx, "failed to read manifest file form data", slog.String("error", err.Error()))
		w.WriteHeader(http.StatusBadRequest)
		return
	} else if file == nil {
		slog.WarnContext(ctx, "empty form data passed")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	slog.DebugContext(ctx, "uploaded manifest file",
		slog.String("name", header.Filename),
		slog.Int64("size", header.Size),
	)

	opts := entity.InspectionOptions{
		ScanType:                    "",
		Manifest:                    nil,
		ManifestType:                "",
		Lockfile:                    nil,
		LockfileType:                "",
		Mode:                        "",
		ExtractFullOrganizationInfo: false,
		ExtractFullContributorInfo:  false,
		DeepRepositoryInspection:    false,
		InspectIssues:               false,
	}

	inspection, err := h.service.InitInspection(ctx)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	h.service.
		w.WriteHeader(http.StatusOK)
	return
}
