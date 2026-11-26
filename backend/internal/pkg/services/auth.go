package services

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/0x24CaptainParrot/gomarket/internal/config"
	"github.com/0x24CaptainParrot/gomarket/internal/models"
	"github.com/0x24CaptainParrot/gomarket/internal/pkg/repository"
	"github.com/0x24CaptainParrot/gomarket/internal/security"
	"github.com/golang-jwt/jwt/v4"
)

type Authorization interface {
	CreateUser(ctx context.Context, user models.User) (int, error)
	TokenGenerator
	TokenValidator
}

type TokenGenerator interface {
	GenerateTokenPair(ctx context.Context, login, password string) (*TokenPair, error)
	RefreshTokenPair(ctx context.Context, refreshToken string) (*TokenPair, error)
}

type TokenValidator interface {
	ValidateAccessToken(tokenString string) (*CustomClaims, error)
	ValidateRefreshToken(tokenString string) (*CustomClaims, error)
}

type CustomClaims struct {
	jwt.RegisteredClaims
	UserID int `json:"user_id"`
}

type TokenPair struct {
	RefreshToken string `json:"refresh_token"`
	AccessToken  string `json:"access_token"`
	TokenType    Bearer `json:"token_type"`
}

type AuthService struct {
	repo         repository.AuthorizationRepository
	tokenService *TokenService
	cfg          *config.Config
	keyPair      *config.KeyPair
}

type Bearer string

func NewAuthService(repo repository.AuthorizationRepository, cfg *config.Config) *AuthService {
	keyPair, _ := cfg.LoadKeys()

	return &AuthService{
		repo:         repo,
		tokenService: NewTokenService(),
		cfg:          cfg,
		keyPair:      keyPair,
	}
}

func (s *AuthService) CreateUser(ctx context.Context, user models.User) (int, error) {
	user.Login = strings.ToLower(strings.TrimSpace(user.Login))
	hashedPassword, err := security.HashPassword(user.Password, security.DefaultArgon2)
	if err != nil {
		return 0, fmt.Errorf("failed to hash password: %w", err)
	}

	user.Password = hashedPassword
	return s.repo.CreateUser(ctx, user)
}

func (s *AuthService) AuthenticateUser(ctx context.Context, login, password string) (*models.User, error) {
	user, err := s.repo.GetUserByLogin(ctx, strings.ToLower(strings.TrimSpace(login)))
	if err != nil {
		return nil, err
	}

	if !security.VerifyPassword(password, user.Password) {
		return nil, fmt.Errorf("invalid credentials")
	}

	return &user, nil
}

// func (s *AuthService) AuthenticateUser(ctx context.Context, login, password string) func(context.Context, string, string) (*models.User, error) {
// 	// ctx = ctx
// 	// login = login
// 	// password = password
// 	return func(ctx ctx, login login, password) (*models.User, error) {

// 	}
// }

// func (s *TokenService) GenerateTokenPair(authUser func(context.Context, string, string) (*models.User, error)) (*TokenPair, error) {

// }

func (s *AuthService) GenerateTokenPair(ctx context.Context, login, password string) (*TokenPair, error) {
	user, err := s.AuthenticateUser(ctx, login, password)
	if err != nil {
		return nil, err
	}

	accessToken, err := s.GenerateAccessToken(user.ID)
	if err != nil {
		return nil, fmt.Errorf("faile to generate access token: %w", err)
	}

	refreshToken, err := s.GenerateRefreshToken(user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return &TokenPair{
		RefreshToken: refreshToken,
		AccessToken:  accessToken,
		TokenType:    "Bearer",
	}, nil
}

func (s *AuthService) RefreshTokenPair(ctx context.Context, refreshToken string) (*TokenPair, error) {
	claims, err := s.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token")
	}

	user, err := s.repo.GetUserByID(ctx, claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	accessToken, err := s.GenerateAccessToken(user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	newRefreshToken, err := s.GenerateRefreshToken(user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		TokenType:    "Bearer",
	}, nil
}

type TokenService struct {
	TokenValidator
}

func NewTokenService() *TokenService {
	return &TokenService{}
}

func (s *AuthService) GenerateAccessToken(userId int) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, CustomClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.cfg.AuthCfg.AccessTTL)),
		},
		UserID: userId,
	})

	return token.SignedString(s.keyPair.PrivateKey)
}

func (s *AuthService) GenerateRefreshToken(userId int) (string, error) {
	refreshSecret, err := base64.StdEncoding.DecodeString(s.cfg.AuthCfg.RefreshSecretB64)
	if err != nil {
		return "", fmt.Errorf("failed to decode refresh secret: %w", err)
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, CustomClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.cfg.AuthCfg.RefreshTTL)),
		},
		UserID: userId,
	})

	return token.SignedString(refreshSecret)
}

func (s *AuthService) ValidateAccessToken(tokenString string) (*CustomClaims, error) {
	claims := &CustomClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims,
		func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return s.keyPair.PublicKey, nil
		},
	)

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("inavalid token")
	}

	return claims, nil
}

func (s *AuthService) ValidateRefreshToken(tokenString string) (*CustomClaims, error) {
	refreshSecret, err := base64.StdEncoding.DecodeString(s.cfg.AuthCfg.RefreshSecretB64)
	if err != nil {
		return nil, fmt.Errorf("failed to decode secret: %w", err)
	}

	claims := &CustomClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims,
		func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return refreshSecret, nil
		},
	)

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("token is not valid")
	}

	return claims, nil
}

// func (s *AuthService) GetRefreshSecret() (string, error) {
// 	return
// }
