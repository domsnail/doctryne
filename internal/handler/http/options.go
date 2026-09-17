package http

import (
	"github.com/domsnail/doctryne/cfg"
	"github.com/domsnail/doctryne/internal/service"
)

type HandlerOptions struct {
	InspectionService    service.IInspectionService
	DeveloperService     service.IDeveloperService
	VulnerabilityService service.IVulnerabilityService

	Config *cfg.ServerConfig
}
