package handlers

import (
	"github.com/0x24CaptainParrot/gomarket/internal/config"
	"github.com/0x24CaptainParrot/gomarket/internal/pkg/services"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	AuthHandler *AuthHandler
	Services    *services.Service
}

func NewHandler(services *services.Service, cfg *config.Config) *Handler {
	return &Handler{
		AuthHandler: NewAuthHandler(services.AuthService, cfg),
		Services:    services,
	}
}

func (h *Handler) InitApiRoutes() *chi.Mux {
	r := chi.NewRouter()

	r.Route("/api/v1", func(r chi.Router) {
		r.Mount("auth/", h.AuthHandler.InitAuthRoutes())

		r.Group(func(r chi.Router) {

		})
	})

	return r
}
