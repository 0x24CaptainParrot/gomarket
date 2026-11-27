package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
)

type ErrorResponse struct {
	Error   string `json:"error"`
	Details string `json:"details,omitempty"`
}

type TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expire_in"`
}

var (
	ErrUnsupportedMediaType = errors.New("unsupported media type")
	ErrLoginOccupied        = errors.New("user with the given login exists")
	ErrInvalidCredentials   = errors.New("invalid credential")
	ErrNoRefreshToken       = errors.New("refresh token was not given")
)

func writeErrorResponse(w http.ResponseWriter, statusCode int, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	errResp := ErrorResponse{
		Error: http.StatusText(statusCode),
	}
	if err != nil {
		errResp.Details = err.Error()
	}

	json.NewEncoder(w).Encode(errResp)
}

func writeJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}
