package http

import (
	"encoding/json"
	"net/http"

	"github.com/domsnail/doctryne/cfg"
	"github.com/domsnail/doctryne/internal/service"
)

type JsonHandler struct {
	vulnerabilities service.IVulnerabilityService

	config *cfg.ServerConfig
}

func newJSONHandler(opts *HandlerOptions) JsonHandler {
	if opts.InspectionService == nil || opts.DeveloperService == nil {
		panic("service is nil")
	}

	return JsonHandler{vulnerabilities: opts.VulnerabilityService, config: opts.Config}
}

func (h *JsonHandler) HandleMux(mux *http.ServeMux) {
	mux.HandleFunc("/vulnerabilities/{canonical_id}", h.handleVulnerabilityByCanonicalID)
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
