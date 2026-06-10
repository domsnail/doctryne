package cli

import "github.com/domsnail/doctryne/internal/service"

type InspectionCommandLineHandler struct {
	service service.IInspectionService
}

func NewInspectionCommandLineHandler(service service.IInspectionService) *InspectionCommandLineHandler {
	return &InspectionCommandLineHandler{service: service}
}
