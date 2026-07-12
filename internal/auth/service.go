package auth

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
)

var ErrInvalidCredentials = errors.New("credenciais invalidas")

type Service struct {
	repo            *Repository
	tokens          *TokenManager
	refreshTokenTTL time.Duration
	sessionTTL      time.Duration
}

type TokenResponse struct {
	AccessToken      string    `json:"access_token"`
	ExpiresAt        time.Time `json:"expires_at"`
	RefreshToken     string    `json:"refresh_token"`
	RefreshExpiresAt time.Time `json:"refresh_expires_at"`
	SessionExpiresAt time.Time `json:"session_expires_at"`
}

func NewService(repo *Repository, tokens *TokenManager, refreshTokenTTL, sessionTTL time.Duration) *Service {
	return &Service{
		repo:            repo,
		tokens:          tokens,
		refreshTokenTTL: refreshTokenTTL,
		sessionTTL:      sessionTTL,
	}
}

func (s *Service) Register(ctx context.Context, name, email, password string) (TokenResponse, error) {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return TokenResponse{}, err
	}

	user, err := s.repo.CreateUser(ctx, name, email, string(passwordHash))
	if err != nil {
		return TokenResponse{}, err
	}
	return s.issueTokenFamily(ctx, user)
}

func (s *Service) Login(ctx context.Context, email, password string) (TokenResponse, error) {
	user, err := s.repo.FindUserByEmail(ctx, email)
	if errors.Is(err, sql.ErrNoRows) {
		return TokenResponse{}, ErrInvalidCredentials
	}
	if err != nil {
		return TokenResponse{}, err
	}

	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) != nil {
		return TokenResponse{}, ErrInvalidCredentials
	}
	return s.issueTokenFamily(ctx, user)
}

func (s *Service) Refresh(ctx context.Context, refreshToken string) (TokenResponse, error) {
	oldHash := HashRefreshToken(refreshToken)
	session, err := s.repo.FindRefreshSession(ctx, oldHash)
	if err != nil {
		return TokenResponse{}, err
	}

	now := time.Now().UTC()
	if session.RevokedAt.Valid || session.ReplacedByHash.Valid {
		_ = s.repo.RevokeRefreshFamily(ctx, session.FamilyID)
		return TokenResponse{}, errors.New("refresh token revogado")
	}
	if now.After(session.ExpiresAt) || now.After(session.SessionExpiresAt) {
		_ = s.repo.RevokeRefreshToken(ctx, oldHash)
		return TokenResponse{}, errors.New("refresh token expirado")
	}

	user, err := s.repo.FindUserByID(ctx, session.UserID)
	if err != nil {
		return TokenResponse{}, err
	}

	newRefresh, newHash, err := NewRefreshToken()
	if err != nil {
		return TokenResponse{}, err
	}

	refreshExpiresAt := now.Add(s.refreshTokenTTL)
	if refreshExpiresAt.After(session.SessionExpiresAt) {
		refreshExpiresAt = session.SessionExpiresAt
	}

	if err := s.repo.RotateRefreshToken(ctx, oldHash, newHash, refreshExpiresAt); err != nil {
		return TokenResponse{}, err
	}

	access, accessExpiresAt, err := s.tokens.NewAccessToken(user.ID, user.Email)
	if err != nil {
		return TokenResponse{}, err
	}

	return TokenResponse{
		AccessToken:      access,
		ExpiresAt:        accessExpiresAt,
		RefreshToken:     newRefresh,
		RefreshExpiresAt: refreshExpiresAt,
		SessionExpiresAt: session.SessionExpiresAt,
	}, nil
}

func (s *Service) Logout(ctx context.Context, refreshToken string) error {
	return s.repo.RevokeRefreshToken(ctx, HashRefreshToken(refreshToken))
}

func (s *Service) issueTokenFamily(ctx context.Context, user User) (TokenResponse, error) {
	now := time.Now().UTC()
	refreshToken, refreshHash, err := NewRefreshToken()
	if err != nil {
		return TokenResponse{}, err
	}
	familyID, err := NewFamilyID()
	if err != nil {
		return TokenResponse{}, err
	}

	refreshExpiresAt := now.Add(s.refreshTokenTTL)
	sessionExpiresAt := now.Add(s.sessionTTL)
	if refreshExpiresAt.After(sessionExpiresAt) {
		refreshExpiresAt = sessionExpiresAt
	}

	if err := s.repo.CreateRefreshSession(ctx, user.ID, refreshHash, familyID, refreshExpiresAt, sessionExpiresAt); err != nil {
		return TokenResponse{}, err
	}

	access, accessExpiresAt, err := s.tokens.NewAccessToken(user.ID, user.Email)
	if err != nil {
		return TokenResponse{}, err
	}

	return TokenResponse{
		AccessToken:      access,
		ExpiresAt:        accessExpiresAt,
		RefreshToken:     refreshToken,
		RefreshExpiresAt: refreshExpiresAt,
		SessionExpiresAt: sessionExpiresAt,
	}, nil
}
