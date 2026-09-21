package http

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"uuid"

	"github.com/domsnail/doctryne/cfg"
	"github.com/domsnail/doctryne/internal/entity"
	"github.com/domsnail/doctryne/internal/service"
	types2 "github.com/domsnail/doctryne/internal/types"
	"github.com/domsnail/doctryne/web/templates"
)

type HtmlHandler struct {
	inspections service.IInspectionService
	developers  service.IDeveloperService

	config *cfg.ServerConfig
}

func newHTMLHandler(opts *HandlerOptions) HtmlHandler {
	if opts.InspectionService == nil || opts.DeveloperService == nil {
		panic("service is nil")
	}

	return HtmlHandler{inspections: opts.InspectionService, developers: opts.DeveloperService, config: opts.Config}
}

func (h *HtmlHandler) HandleMux(mux *http.ServeMux) {
	mux.HandleFunc("/upload", h.handleManifestUpload)

	mux.HandleFunc("/inspections/{uuid}", h.handleInspectionPage)
	mux.HandleFunc("/inspections/{uuid}/revisions/{revision}", h.handleInspectionPage)
	mux.HandleFunc("/inspections", h.handleInspectionsPage)

	mux.HandleFunc("/developers/{uuid}", h.handleDeveloperPage)
	mux.HandleFunc("/developers/{uuid}/card", h.handleDeveloperCard)

	mux.HandleFunc("/", h.static())
}

func (h *HtmlHandler) handleManifestUpload(w http.ResponseWriter, r *http.Request) {
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

	inspection, err := h.inspections.InitInspection(ctx, opts)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = h.inspections.InspectManifests(ctx, inspection)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = h.inspections.InspectPackages(ctx, inspection)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = h.inspections.InspectDevelopersAndOrganizations(ctx, inspection)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = h.inspections.SaveInspection(ctx, inspection)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// https://htmx.org/headers/hx-redirect/
	w.Header().Set("HX-Redirect", fmt.Sprintf("/inspections/%s/revisions/%d", inspection.UUID.String(), inspection.Revision))
	w.WriteHeader(http.StatusOK)
	return
}

func (h *HtmlHandler) handleInspectionPage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var (
		rev uint64
		err error
	)

	inspectionUUID := r.PathValue("uuid")
	inspectionRevision := r.PathValue("revision")
	if inspectionRevision != "" {
		rev, err = strconv.ParseUint(inspectionRevision, 10, 32)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}

	uid, err := uuid.Parse(inspectionUUID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	inspection, err := h.inspections.GetInspectionByUUID(ctx, uid, uint32(rev))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	err = templates.InspectionPage(inspection).Render(ctx, w)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	return
}

func (h *HtmlHandler) handleDeveloperPage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	developerUUID := r.PathValue("uuid")
	uid, err := uuid.Parse(developerUUID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	developer, err := h.developers.GetDeveloperByUUID(ctx, uid)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	err = templates.DeveloperPage(developer).Render(ctx, w)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	return
}

func (h *HtmlHandler) handleDeveloperCard(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	developerUUID := r.PathValue("uuid")
	uid, err := uuid.Parse(developerUUID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	developer, err := h.developers.GetDeveloperByUUID(ctx, uid)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	err = templates.DeveloperCard(developer).Render(ctx, w)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	return
}

func (h *HtmlHandler) handleInspectionsPage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	inspections, err := h.inspections.GetInspectionsByQueryFilter(ctx, entity.InspectionsQueryFilter{
		QueryFilter: entity.QueryFilter{
			Limit: 10,
		},
	})

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	err = templates.InspectionsPage(inspections).Render(ctx, w)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	return
}

func parseInspectionOptionsForm(ctx context.Context, req *http.Request) (*entity.InspectionOptions, error) {
	var opts = entity.InspectionOptions{
		ScanType: types2.ScanType_Binary,
		Mode:     types2.InspectionMode_Direct,
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
	opts.ManifestType = types2.ManifestType(req.FormValue("manifest-type"))
	opts.ManifestName = header.Filename

	opts.DeepRepositoryInspection = req.FormValue("deep-repository-inspection") == "on"
	opts.ExtractFullContributorInfo = req.FormValue("extract-full-contibutor-info") == "on"
	opts.ExtractFullOrganizationInfo = req.FormValue("extract-full-organization-info") == "on"

	return &opts, nil
}
