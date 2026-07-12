package auth

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

type User struct {
	ID           int64
	Name         string
	Email        string
	PasswordHash string
}

type RefreshSession struct {
	ID               int64
	UserID           int64
	TokenHash        string
	FamilyID         string
	ExpiresAt        time.Time
	SessionExpiresAt time.Time
	RevokedAt        sql.NullTime
	ReplacedByHash   sql.NullString
}

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateUser(ctx context.Context, name, email, passwordHash string) (User, error) {
	var user User
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO users (name, email, password_hash)
		VALUES ($1, $2, $3)
		RETURNING id, name, email, password_hash
	`, name, email, passwordHash).Scan(&user.ID, &user.Name, &user.Email, &user.PasswordHash)
	return user, err
}

func (r *Repository) FindUserByEmail(ctx context.Context, email string) (User, error) {
	var user User
	err := r.db.QueryRowContext(ctx, `
		SELECT id, name, email, password_hash
		FROM users
		WHERE email = $1
	`, email).Scan(&user.ID, &user.Name, &user.Email, &user.PasswordHash)
	return user, err
}

func (r *Repository) CreateRefreshSession(ctx context.Context, userID int64, tokenHash, familyID string, expiresAt, sessionExpiresAt time.Time) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO refresh_tokens (user_id, token_hash, family_id, expires_at, session_expires_at)
		VALUES ($1, $2, $3, $4, $5)
	`, userID, tokenHash, familyID, expiresAt, sessionExpiresAt)
	return err
}

func (r *Repository) FindRefreshSession(ctx context.Context, tokenHash string) (RefreshSession, error) {
	var session RefreshSession
	err := r.db.QueryRowContext(ctx, `
		SELECT id, user_id, token_hash, family_id, expires_at, session_expires_at, revoked_at, replaced_by_hash
		FROM refresh_tokens
		WHERE token_hash = $1
	`, tokenHash).Scan(
		&session.ID,
		&session.UserID,
		&session.TokenHash,
		&session.FamilyID,
		&session.ExpiresAt,
		&session.SessionExpiresAt,
		&session.RevokedAt,
		&session.ReplacedByHash,
	)
	return session, err
}

func (r *Repository) FindUserByID(ctx context.Context, userID int64) (User, error) {
	var user User
	err := r.db.QueryRowContext(ctx, `
		SELECT id, name, email, password_hash
		FROM users
		WHERE id = $1
	`, userID).Scan(&user.ID, &user.Name, &user.Email, &user.PasswordHash)
	return user, err
}

func (r *Repository) RotateRefreshToken(ctx context.Context, oldHash, newHash string, expiresAt time.Time) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var session RefreshSession
	err = tx.QueryRowContext(ctx, `
		SELECT id, user_id, family_id, session_expires_at, revoked_at, replaced_by_hash
		FROM refresh_tokens
		WHERE token_hash = $1
		FOR UPDATE
	`, oldHash).Scan(&session.ID, &session.UserID, &session.FamilyID, &session.SessionExpiresAt, &session.RevokedAt, &session.ReplacedByHash)
	if err != nil {
		return err
	}

	if session.RevokedAt.Valid || session.ReplacedByHash.Valid {
		_, _ = tx.ExecContext(ctx, `UPDATE refresh_tokens SET revoked_at = now() WHERE family_id = $1 AND revoked_at IS NULL`, session.FamilyID)
		return errors.New("refresh token reutilizado")
	}

	_, err = tx.ExecContext(ctx, `
		UPDATE refresh_tokens
		SET revoked_at = now(), replaced_by_hash = $1
		WHERE id = $2
	`, newHash, session.ID)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO refresh_tokens (user_id, token_hash, family_id, expires_at, session_expires_at)
		VALUES ($1, $2, $3, $4, $5)
	`, session.UserID, newHash, session.FamilyID, expiresAt, session.SessionExpiresAt)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *Repository) RevokeRefreshToken(ctx context.Context, tokenHash string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE refresh_tokens
		SET revoked_at = now()
		WHERE token_hash = $1 AND revoked_at IS NULL
	`, tokenHash)
	return err
}

func (r *Repository) RevokeRefreshFamily(ctx context.Context, familyID string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE refresh_tokens
		SET revoked_at = now()
		WHERE family_id = $1 AND revoked_at IS NULL
	`, familyID)
	return err
}
