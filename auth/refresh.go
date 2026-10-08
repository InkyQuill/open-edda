package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"time"
)

const refreshLifetime = 30 * 24 * time.Hour
const refreshRenewWindow = 7 * 24 * time.Hour

var ErrInvalidRefresh = errors.New("invalid or expired refresh token")

func refreshHash(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}

func (s *Service) newSession(ctx context.Context, author AuthorPublic) (AuthResponse, error) {
	token, err := GenerateToken(author.ID, author.Email, s.secret)
	if err != nil {
		return AuthResponse{}, err
	}
	refresh := rand.Text() + rand.Text()
	expires := time.Now().Add(refreshLifetime).Unix()
	if _, err := s.db.ExecContext(ctx, "INSERT INTO refresh_sessions(token_hash,author_id,expires_at) VALUES(?,?,?)", refreshHash(refresh), author.ID, expires); err != nil {
		return AuthResponse{}, err
	}
	return AuthResponse{Token: token, RefreshToken: refresh, RefreshExpiresAt: expires, Author: author}, nil
}

func (s *Service) Refresh(ctx context.Context, token string) (AuthResponse, error) {
	return s.refreshAt(ctx, token, time.Now())
}

func (s *Service) refreshAt(ctx context.Context, token string, now time.Time) (AuthResponse, error) {
	if len(token) < 32 || len(token) > 256 {
		return AuthResponse{}, ErrInvalidRefresh
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return AuthResponse{}, err
	}
	defer tx.Rollback()
	// Acquire the SQLite write lock before reading so concurrent rotations see
	// either the old session or the committed replacement, never a stale snapshot.
	if _, err := tx.ExecContext(ctx, "UPDATE refresh_sessions SET expires_at=expires_at WHERE token_hash=?", refreshHash(token)); err != nil {
		return AuthResponse{}, err
	}
	var author AuthorPublic
	var expires int64
	err = tx.QueryRowContext(ctx, `SELECT a.id,a.email,s.expires_at FROM refresh_sessions s JOIN authors a ON a.id=s.author_id WHERE s.token_hash=? AND s.expires_at>?`, refreshHash(token), now.Unix()).Scan(&author.ID, &author.Email, &expires)
	if errors.Is(err, sql.ErrNoRows) {
		return AuthResponse{}, ErrInvalidRefresh
	}
	if err != nil {
		return AuthResponse{}, err
	}
	refresh := token
	if expires-now.Unix() < int64(refreshRenewWindow/time.Second) {
		refresh = rand.Text() + rand.Text()
		expires = now.Add(refreshLifetime).Unix()
		result, err := tx.ExecContext(ctx, "UPDATE refresh_sessions SET token_hash=?,expires_at=? WHERE token_hash=?", refreshHash(refresh), expires, refreshHash(token))
		if err != nil {
			return AuthResponse{}, err
		}
		count, err := result.RowsAffected()
		if err != nil {
			return AuthResponse{}, err
		}
		if count != 1 {
			return AuthResponse{}, ErrInvalidRefresh
		}
	}
	access, err := GenerateToken(author.ID, author.Email, s.secret)
	if err != nil {
		return AuthResponse{}, err
	}
	if _, err := tx.ExecContext(ctx, "DELETE FROM refresh_sessions WHERE expires_at<=?", now.Unix()); err != nil {
		return AuthResponse{}, err
	}
	if err := tx.Commit(); err != nil {
		return AuthResponse{}, err
	}
	return AuthResponse{Token: access, RefreshToken: refresh, RefreshExpiresAt: expires, Author: author}, nil
}

func (s *Service) RevokeRefresh(ctx context.Context, token string) error {
	if token == "" {
		return nil
	}
	_, err := s.db.ExecContext(ctx, "DELETE FROM refresh_sessions WHERE token_hash=?", refreshHash(token))
	return err
}
