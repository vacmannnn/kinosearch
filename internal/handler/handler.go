package handler

import (
	"log/slog"

	"github.com/vacmannnn/kinosearch/internal/service"
)

type Handler struct {
	service *service.Service
	logger  *slog.Logger
}

func New(service *service.Service, logger *slog.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}
