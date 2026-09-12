package authservice

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/ladev707/3dbuilder-backend/internal/config"
	authaction "github.com/ladev707/3dbuilder-backend/internal/src/auth/actions"
	authmodel "github.com/ladev707/3dbuilder-backend/internal/src/auth/models"
	authrepository "github.com/ladev707/3dbuilder-backend/internal/src/auth/repository"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInactivePrincipal  = errors.New("account is inactive")
	ErrInvalidToken       = errors.New("invalid token")
)

type Repository interface {
	FindPrincipalByEmail(context.Context, authmodel.Gate, string) (authmodel.Principal, error)
	FindPrincipalByID(context.Context, authmodel.Gate, uuid.UUID) (authmodel.Principal, error)
	LoadAuthorization(context.Context, *authmodel.Principal) error
	CreateSession(context.Context, authmodel.Session) error
	RotateSession(context.Context, uuid.UUID, string, authmodel.Session) error
	RevokeSession(context.Context, uuid.UUID, uuid.UUID, authmodel.Gate) error
	ListRoles(context.Context, authmodel.Gate) ([]authmodel.Role, error)
	ListPermissions(context.Context, authmodel.Gate) ([]authmodel.Permission, error)
	CreateRole(context.Context, authmodel.Gate, string, string) (authmodel.Role, error)
	AssignPermission(context.Context, authmodel.Gate, uuid.UUID, uuid.UUID) error
	RemovePermission(context.Context, authmodel.Gate, uuid.UUID, uuid.UUID) error
	AssignRole(context.Context, authmodel.Gate, uuid.UUID, uuid.UUID) error
	RemoveRole(context.Context, authmodel.Gate, uuid.UUID, uuid.UUID) error
}

type Claims struct {
	Gate        authmodel.Gate `json:"gate"`
	TokenType   string         `json:"token_type"`
	Roles       []string       `json:"roles,omitempty"`
	Permissions []string       `json:"permissions,omitempty"`
	jwt.RegisteredClaims
}

type Service struct {
	repository Repository
	config     config.AuthConfig
	now        func() time.Time
}

func New(repository Repository, cfg config.AuthConfig) *Service {
	return &Service{repository: repository, config: cfg, now: time.Now}
}

func (s *Service) Login(ctx context.Context, gate authmodel.Gate, request authmodel.LoginRequest) (authmodel.LoginResponse, error) {
	if !gate.Valid() {
		return authmodel.LoginResponse{}, ErrInvalidCredentials
	}
	principal, err := s.repository.FindPrincipalByEmail(ctx, gate, request.Email)
	if err != nil {
		if errors.Is(err, authrepository.ErrPrincipalNotFound) {
			return authmodel.LoginResponse{}, ErrInvalidCredentials
		}
		return authmodel.LoginResponse{}, err
	}
	if authaction.CheckPassword(principal.PasswordHash, request.Password) != nil {
		return authmodel.LoginResponse{}, ErrInvalidCredentials
	}
	if !principal.Active {
		return authmodel.LoginResponse{}, ErrInactivePrincipal
	}
	if err := s.repository.LoadAuthorization(ctx, &principal); err != nil {
		return authmodel.LoginResponse{}, err
	}

	tokens, session, err := s.issueTokenPair(principal)
	if err != nil {
		return authmodel.LoginResponse{}, err
	}
	if err := s.repository.CreateSession(ctx, session); err != nil {
		return authmodel.LoginResponse{}, err
	}
	return authmodel.LoginResponse{Tokens: tokens, Principal: principal}, nil
}

func (s *Service) Refresh(ctx context.Context, gate authmodel.Gate, rawRefreshToken string) (authmodel.LoginResponse, error) {
	claims, err := s.ParseToken(rawRefreshToken, "refresh")
	if err != nil || claims.Gate != gate {
		return authmodel.LoginResponse{}, ErrInvalidToken
	}
	principalID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return authmodel.LoginResponse{}, ErrInvalidToken
	}
	oldSessionID, err := uuid.Parse(claims.ID)
	if err != nil {
		return authmodel.LoginResponse{}, ErrInvalidToken
	}

	principal, err := s.repository.FindPrincipalByID(ctx, gate, principalID)
	if err != nil || !principal.Active {
		return authmodel.LoginResponse{}, ErrInvalidToken
	}
	if err := s.repository.LoadAuthorization(ctx, &principal); err != nil {
		return authmodel.LoginResponse{}, err
	}
	tokens, session, err := s.issueTokenPair(principal)
	if err != nil {
		return authmodel.LoginResponse{}, err
	}
	if err := s.repository.RotateSession(ctx, oldSessionID, hashToken(rawRefreshToken), session); err != nil {
		if errors.Is(err, authrepository.ErrInvalidSession) {
			return authmodel.LoginResponse{}, ErrInvalidToken
		}
		return authmodel.LoginResponse{}, err
	}
	return authmodel.LoginResponse{Tokens: tokens, Principal: principal}, nil
}

func (s *Service) Logout(ctx context.Context, gate authmodel.Gate, rawRefreshToken string) error {
	claims, err := s.ParseToken(rawRefreshToken, "refresh")
	if err != nil || claims.Gate != gate {
		return ErrInvalidToken
	}
	principalID, principalErr := uuid.Parse(claims.Subject)
	sessionID, sessionErr := uuid.Parse(claims.ID)
	if principalErr != nil || sessionErr != nil {
		return ErrInvalidToken
	}
	return s.repository.RevokeSession(ctx, sessionID, principalID, gate)
}

func (s *Service) CurrentPrincipal(ctx context.Context, claims *Claims) (authmodel.Principal, error) {
	principalID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return authmodel.Principal{}, ErrInvalidToken
	}
	principal, err := s.repository.FindPrincipalByID(ctx, claims.Gate, principalID)
	if err != nil || !principal.Active {
		return authmodel.Principal{}, ErrInvalidToken
	}
	if err := s.repository.LoadAuthorization(ctx, &principal); err != nil {
		return authmodel.Principal{}, err
	}
	return principal, nil
}

func (s *Service) ListRoles(ctx context.Context, gate authmodel.Gate) ([]authmodel.Role, error) {
	if !gate.Valid() {
		return nil, fmt.Errorf("invalid gate")
	}
	return s.repository.ListRoles(ctx, gate)
}

func (s *Service) ListPermissions(ctx context.Context, gate authmodel.Gate) ([]authmodel.Permission, error) {
	if !gate.Valid() {
		return nil, fmt.Errorf("invalid gate")
	}
	return s.repository.ListPermissions(ctx, gate)
}

func (s *Service) CreateRole(ctx context.Context, gate authmodel.Gate, request authmodel.CreateRoleRequest) (authmodel.Role, error) {
	if !gate.Valid() {
		return authmodel.Role{}, fmt.Errorf("invalid gate")
	}
	return s.repository.CreateRole(ctx, gate, strings.ToLower(strings.TrimSpace(request.Name)), strings.TrimSpace(request.Description))
}

func (s *Service) AssignPermission(ctx context.Context, gate authmodel.Gate, roleID, permissionID uuid.UUID) error {
	if !gate.Valid() {
		return fmt.Errorf("invalid gate")
	}
	return s.repository.AssignPermission(ctx, gate, roleID, permissionID)
}

func (s *Service) RemovePermission(ctx context.Context, gate authmodel.Gate, roleID, permissionID uuid.UUID) error {
	if !gate.Valid() {
		return fmt.Errorf("invalid gate")
	}
	return s.repository.RemovePermission(ctx, gate, roleID, permissionID)
}

func (s *Service) AssignRole(ctx context.Context, gate authmodel.Gate, principalID, roleID uuid.UUID) error {
	if !gate.Valid() {
		return fmt.Errorf("invalid gate")
	}
	return s.repository.AssignRole(ctx, gate, principalID, roleID)
}

func (s *Service) RemoveRole(ctx context.Context, gate authmodel.Gate, principalID, roleID uuid.UUID) error {
	if !gate.Valid() {
		return fmt.Errorf("invalid gate")
	}
	return s.repository.RemoveRole(ctx, gate, principalID, roleID)
}

func (s *Service) ParseToken(rawToken, expectedType string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(
		rawToken,
		claims,
		func(token *jwt.Token) (any, error) {
			return []byte(s.config.Secret), nil
		},
		jwt.WithIssuer(s.config.Issuer),
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
	)
	if err != nil || !token.Valid || claims.TokenType != expectedType || !claims.Gate.Valid() {
		return nil, ErrInvalidToken
	}
	if _, err := uuid.Parse(claims.Subject); err != nil {
		return nil, ErrInvalidToken
	}
	return claims, nil
}

func (s *Service) issueTokenPair(principal authmodel.Principal) (authmodel.TokenPair, authmodel.Session, error) {
	now := s.now().UTC()
	accessExpiry := now.Add(s.config.AccessTokenTTL)
	refreshExpiry := now.Add(s.config.RefreshTokenTTL)
	sessionID := uuid.New()

	accessClaims := Claims{
		Gate:        principal.Gate,
		TokenType:   "access",
		Roles:       principal.Roles,
		Permissions: principal.Permissions,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.config.Issuer,
			Subject:   principal.ID.String(),
			ID:        uuid.NewString(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(accessExpiry),
		},
	}
	refreshClaims := Claims{
		Gate:      principal.Gate,
		TokenType: "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.config.Issuer,
			Subject:   principal.ID.String(),
			ID:        sessionID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(refreshExpiry),
		},
	}
	accessToken, err := s.sign(accessClaims)
	if err != nil {
		return authmodel.TokenPair{}, authmodel.Session{}, fmt.Errorf("sign access token: %w", err)
	}
	refreshToken, err := s.sign(refreshClaims)
	if err != nil {
		return authmodel.TokenPair{}, authmodel.Session{}, fmt.Errorf("sign refresh token: %w", err)
	}

	return authmodel.TokenPair{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
			TokenType:    "Bearer",
			ExpiresIn:    int64(s.config.AccessTokenTTL.Seconds()),
			ExpiresAt:    accessExpiry,
		}, authmodel.Session{
			ID:               sessionID,
			PrincipalID:      principal.ID,
			Gate:             principal.Gate,
			RefreshTokenHash: hashToken(refreshToken),
			ExpiresAt:        refreshExpiry,
		}, nil
}

func (s *Service) sign(claims Claims) (string, error) {
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(s.config.Secret))
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
