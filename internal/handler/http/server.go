package http

import (
	"net/http"

	"github.com/domsnail/doctryne/cfg"
	"github.com/domsnail/doctryne/internal/service"
)

type Handler struct {
	service service.IInspectionService
	config  *cfg.Server
}

const pathPrefix = "/api/v1"

func NewHandler(service service.IInspectionService, config *cfg.Server) *Handler {
	if service == nil {
		panic("inspection service is nil")
	}

	return &Handler{service: service, config: config}
}

func (h *Handler) HandleMux(mux *http.ServeMux) {
	mux.HandleFunc(pathPrefix+"/upload", h.handleManifestUpload)

	mux.HandleFunc("/inspections/{uuid}/revisions", h.handleInspectionPage)
	mux.HandleFunc("/inspections/{uuid}/revisions/{revision}", h.handleInspectionPage)

	mux.HandleFunc("/", h.static())
}
