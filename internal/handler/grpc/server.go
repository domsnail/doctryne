package grpc

import (
	"context"

	"github.com/domsnail/doctryne/internal/entity"
	"github.com/domsnail/doctryne/internal/service"
)

type InspectionGRPCHandler struct {
	service service.IInspectionService
}

func NewInspectionGRPCHandler(service service.IInspectionService) *InspectionGRPCHandler {
	entity.NewInspection()

	return &InspectionGRPCHandler{service: service}
}

func (handler *InspectionGRPCHandler) Inspect(ctx context.Context, opts *entity.InspectionOptions) error {

}
