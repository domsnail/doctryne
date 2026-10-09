package http

import (
	"github.com/domsnail/doctryne/cfg"
	"github.com/domsnail/doctryne/internal/service"
)

type HandlerOptions struct {
	InspectionService            service.IInspectionService
	DeveloperService             service.IDeveloperService
	VulnerabilityService         service.IVulnerabilityService
	VulnerabilityDatabaseService service.IVulnerabilityDatabaseService
	VulnerabilityMatcherService  service.IVulnerabilityMatcherService

	Config *cfg.ServerConfig
}
