package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/0x24CaptainParrot/gomarket/internal/config"
	"github.com/0x24CaptainParrot/gomarket/internal/models"
	"github.com/0x24CaptainParrot/gomarket/internal/pkg/repository"
	"github.com/0x24CaptainParrot/gomarket/internal/pkg/services"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
)

type AuthHandler struct {
	AuthService services.Authorization
	Cfg         *config.Config
	Validator   *validator.Validate
}

func NewAuthHandler(authService services.Authorization, cfg *config.Config) *AuthHandler {
	validator := validator.New()
	return &AuthHandler{
		AuthService: authService,
		Cfg:         cfg,
		Validator:   validator,
	}
}

func (a *AuthHandler) InitAuthRoutes() *chi.Mux {
	r := chi.NewRouter()

	r.Post("/register", a.RegisterUser)
	r.Post("/login", a.LoginUser)
	r.Post("/refresh", a.Refresh)
	r.Get("/me", a.Me)
	r.Post("/logout", a.Logout)

	return r
}

func (a *AuthHandler) RegisterUser(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") != "application/json" {
		writeErrorResponse(w, http.StatusUnsupportedMediaType, ErrUnsupportedMediaType)
		return
	}

	var userRegDTO models.RegisterUserDTO
	if err := json.NewDecoder(r.Body).Decode(&userRegDTO); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, err)
		return
	}

	userRegDTO.Login = strings.TrimSpace(strings.ToLower(userRegDTO.Login))
	if err := a.Validator.Struct(userRegDTO); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, err)
		return
	}

	user := models.User{
		Login:    userRegDTO.Login,
		Email:    userRegDTO.Email,
		Password: userRegDTO.Password,
	}

	ctx := r.Context()
	_, err := a.AuthService.CreateUser(ctx, user)
	if err != nil {
		if errors.Is(err, repository.ErrUserExists) {
			writeErrorResponse(w, http.StatusConflict, ErrLoginOccupied)
			return
		}
		writeErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	tokenPair, err := a.AuthService.GenerateTokenPair(ctx, user.Login, user.Password)
	if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, err)
		return
	}
	a.SetRefreshTokenCookie(w, tokenPair.RefreshToken)

	writeJSON(w, http.StatusCreated, TokenResponse{
		AccessToken: tokenPair.AccessToken,
		TokenType:   "Bearer",
		ExpiresIn:   int(a.Cfg.AuthCfg.AccessTTL.Seconds()),
	})
}

func (a *AuthHandler) LoginUser(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") != "application/json" {
		writeErrorResponse(w, http.StatusUnsupportedMediaType, ErrUnsupportedMediaType)
		return
	}

	var userLoginDTO models.LoginUserDTO
	if err := json.NewDecoder(r.Body).Decode(&userLoginDTO); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, err)
		return
	}

	if err := a.Validator.Struct(userLoginDTO); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, err)
		return
	}

	ctx := r.Context()
	tokenPair, err := a.AuthService.GenerateTokenPair(ctx, userLoginDTO.Login, userLoginDTO.Password)
	if err != nil {
		writeErrorResponse(w, http.StatusUnauthorized, ErrInvalidCredentials)
		return
	}
	a.SetRefreshTokenCookie(w, tokenPair.RefreshToken)

	writeJSON(w, http.StatusOK, TokenResponse{
		AccessToken: tokenPair.AccessToken,
		TokenType:   "Bearer",
		ExpiresIn:   int(a.Cfg.AuthCfg.AccessTTL.Seconds()),
	})
}

func (a *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		writeErrorResponse(w, http.StatusUnauthorized, ErrNoRefreshToken)
		return
	}

	ctx := r.Context()
	tokenPair, err := a.AuthService.RefreshTokenPair(ctx, cookie.Value)
	if err != nil {
		writeErrorResponse(w, http.StatusUnauthorized, err)
		return
	}
	a.SetRefreshTokenCookie(w, tokenPair.RefreshToken)

	writeJSON(w, http.StatusOK, TokenResponse{
		AccessToken: tokenPair.AccessToken,
		TokenType:   "Bearer",
		ExpiresIn:   int(a.Cfg.AuthCfg.AccessTTL.Seconds()),
	})
}

func (a *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {

}

func (a *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	a.SetRefreshTokenCookie(w, "")
	writeJSON(w, http.StatusOK, map[string]string{"message": "logged out successfully"})
}

func (a *AuthHandler) SetRefreshTokenCookie(w http.ResponseWriter, token string) {
	cookie := &http.Cookie{
		Name:     "refres_token",
		Value:    token,
		Path:     "/api/v1/auth",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}

	if token == "" {
		cookie.MaxAge = -1
	} else {
		cookie.MaxAge = int(a.Cfg.AuthCfg.RefreshTTL.Seconds())
	}

	http.SetCookie(w, cookie)
}
