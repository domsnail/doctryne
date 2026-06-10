package grpc

import (
	"github.com/domsnail/doctryne/internal/service"
)

type InspectionGRPCHandler struct {
	service service.IInspectionService
}

func NewInspectionGRPCHandler(service service.IInspectionService) *InspectionGRPCHandler {
	return &InspectionGRPCHandler{service: service}
}
