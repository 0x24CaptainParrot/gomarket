package handlers

import (
	"net/http"

	"github.com/0x24CaptainParrot/gomarket/internal/config"
	"github.com/0x24CaptainParrot/gomarket/internal/pkg/services"
	"github.com/go-chi/chi/v5"
)

type AuthHandler struct {
	AuthService services.Authorization
	Cfg         *config.Config
}

func NewAuthHandler(authService services.Authorization, cfg *config.Config) *AuthHandler {
	return &AuthHandler{
		AuthService: authService,
		Cfg:         cfg,
	}
}

func (a *AuthHandler) InitAuthRoutes() *chi.Mux {
	r := chi.NewRouter()

	r.Post("/register", a.RegisterUser)
	r.Post("/login", a.LoginUser)
	r.Post("/refresh", a.Refresh)
	r.Post("/me", a.Me)

	return r
}

func (a *AuthHandler) RegisterUser(w http.ResponseWriter, r *http.Request) {

}

func (a *AuthHandler) LoginUser(w http.ResponseWriter, r *http.Request) {

}

func (a *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {

}

func (a *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {

}
