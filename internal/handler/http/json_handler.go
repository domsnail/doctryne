package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"uuid"

	"github.com/domsnail/doctryne/cfg"
	"github.com/domsnail/doctryne/internal/entity"
	"github.com/domsnail/doctryne/internal/service"
)

type JsonHandler struct {
	vulnerabilities       service.IVulnerabilityService
	vulnerabilityDatabase service.IVulnerabilityDatabaseService

	config *cfg.ServerConfig
}

func newJSONHandler(opts *HandlerOptions) JsonHandler {
	if opts.InspectionService == nil || opts.DeveloperService == nil {
		panic("service is nil")
	}

	return JsonHandler{
		vulnerabilities:       opts.VulnerabilityService,
		vulnerabilityDatabase: opts.VulnerabilityDatabaseService,
		config:                opts.Config,
	}
}

func (h *JsonHandler) HandleMux(mux *http.ServeMux) {
	mux.HandleFunc("/vulnerabilities/{canonical_id}", h.handleVulnerabilityByCanonicalID)

	mux.HandleFunc("/vulnerabilities/databases/latest", h.handleLatestVulnerabilityDatabaseUpdates)
	mux.HandleFunc("/vulnerabilities/databases/updates", h.handleGetVulnerabilityDatabaseUpdatesByQueryFilter)
	mux.HandleFunc("/vulnerabilities/databases/updates/{uuid}", h.handleVulnerabilityDatabaseUpdateByUUID)
}

func (h *JsonHandler) handleLatestVulnerabilityDatabaseUpdates(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	latest, err := h.vulnerabilities.GetLatestVulnerabilityDatabaseUpdates(ctx)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	payload, err := json.Marshal(latest)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(payload)
	return
}

func (h *JsonHandler) handleVulnerabilityByCanonicalID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	canonicalID := r.PathValue("canonical_id")
	vulnerability, err := h.vulnerabilities.GetVulnerabilityByCanonicalID(ctx, canonicalID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	} else if vulnerability == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	payload, err := json.Marshal(vulnerability)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(payload)
	return
}

func (h *JsonHandler) handleVulnerabilityDatabaseUpdateByUUID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	uuidPathValue := r.PathValue("uuid")
	uid, err := uuid.Parse(uuidPathValue)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var update entity.VulnerabilitiesDatabaseUpdate
	switch r.Method {
	case http.MethodGet:
		update, err = h.vulnerabilities.GetVulnerabilityDatabaseUpdateByUUID(ctx, uid)
	case http.MethodPost:
		update, err = h.vulnerabilityDatabase.RunVulnerabilityDatabaseUpdateByUUID(ctx, uid)
	default:
		h.error(w, errors.New("method not allowed"), http.StatusMethodNotAllowed)
		return
	}

	if err != nil {
		h.error(w, err, http.StatusBadRequest)
		return
	}

	payload, err := json.Marshal(update)
	if err != nil {
		h.error(w, err, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(payload)
	return
}

func (h *JsonHandler) handleGetVulnerabilityDatabaseUpdatesByQueryFilter(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if r.Method != http.MethodGet {
		h.error(w, errors.New("method not allowed"), http.StatusMethodNotAllowed)
		return
	}

	filter := entity.VulnerabilitiesDatabaseUpdateQueryFilter{}
	err := filter.FromQuery(r.URL.Query())
	if err != nil {
		h.error(w, err, http.StatusBadRequest)
		return
	}

	updates, err := h.vulnerabilities.GetVulnerabilityDatabaseUpdatesByQueryFilter(ctx, filter)
	if err != nil {
		h.error(w, err, http.StatusBadRequest)
		return
	}

	payload, err := json.Marshal(updates)
	if err != nil {
		h.error(w, err, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(payload)
	return
}

func (h *JsonHandler) error(w http.ResponseWriter, err error, code int) {
	customError := entity.Error{
		StatusCode: code,
		Message:    err.Error(),
		Details:    "",
	}

	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(customError)
	return
}
