package http

import (
	"net/http"

	"github.com/domsnail/doctryne/cfg"
	"github.com/domsnail/doctryne/internal/service"
)

type Handler struct {
	inspections service.IInspectionService
	developers  service.IDeveloperService

	config *cfg.ServerConfig
}

const pathPrefix = "/api/v1"

type HandlerOptions struct {
	InspectionService service.IInspectionService
	DeveloperService  service.IDeveloperService

	Config *cfg.ServerConfig
}

func NewHandler(opts *HandlerOptions) *Handler {
	if opts.InspectionService == nil || opts.DeveloperService == nil {
		panic("service is nil")
	}

	return &Handler{inspections: opts.InspectionService, developers: opts.DeveloperService, config: opts.Config}
}

func (h *Handler) HandleMux(mux *http.ServeMux) {
	mux.HandleFunc(pathPrefix+"/upload", h.handleManifestUpload)

	mux.HandleFunc("/inspections/{uuid}", h.handleInspectionPage)
	mux.HandleFunc("/inspections/{uuid}/revisions/{revision}", h.handleInspectionPage)
	mux.HandleFunc("/inspections", h.handleInspectionsPage)

	mux.HandleFunc("/developers/{uuid}", h.handleDeveloperPage)
	mux.HandleFunc("/developers/{uuid}/card", h.handleDeveloperCard)

	mux.HandleFunc("/", h.static())
}
