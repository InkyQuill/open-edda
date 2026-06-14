package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/mail"
	"strings"

	"git.inkyquill.net/inky/writer/store"
	"github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

const bcryptCost = 12

var (
	ErrInvalidEmail       = errors.New("invalid email")
	ErrPasswordTooShort   = errors.New("password must be at least 8 characters")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrEmailTaken         = errors.New("email already registered")
)

type Service struct {
	db      *sql.DB
	queries *store.Queries
	secret  string
}

func NewService(db *sql.DB, secret string) *Service {
	if err := ValidateSecret(secret); err != nil {
		panic("auth: " + err.Error())
	}
	return &Service{db: db, queries: store.New(db), secret: secret}
}

func (s *Service) Secret() string {
	return s.secret
}

func (s *Service) Register(ctx context.Context, email, password string) (AuthResponse, error) {
	email, err := validateEmail(email)
	if err != nil {
		return AuthResponse{}, err
	}
	if err := validatePassword(password); err != nil {
		return AuthResponse{}, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return AuthResponse{}, fmt.Errorf("hash password: %w", err)
	}

	authorID := newID("author")
	now := nowString()
	err = s.queries.CreateAuthor(ctx, store.CreateAuthorParams{
		ID:           authorID,
		Email:        email,
		PasswordHash: string(hash),
		CreatedAt:    now,
	})
	if err != nil {
		if isSQLiteUniqueConstraint(err) {
			return AuthResponse{}, ErrEmailTaken
		}
		return AuthResponse{}, fmt.Errorf("create author: %w", err)
	}

	token, err := GenerateToken(authorID, email, s.secret)
	if err != nil {
		return AuthResponse{}, fmt.Errorf("generate token: %w", err)
	}

	return AuthResponse{
		Token:  token,
		Author: AuthorPublic{ID: authorID, Email: email},
	}, nil
}

func (s *Service) Login(ctx context.Context, email, password string) (AuthResponse, error) {
	email, err := validateEmail(email)
	if err != nil {
		return AuthResponse{}, err
	}

	author, err := s.queries.GetAuthorByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return AuthResponse{}, ErrInvalidCredentials
		}
		return AuthResponse{}, fmt.Errorf("get author: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(author.PasswordHash), []byte(password)); err != nil {
		return AuthResponse{}, ErrInvalidCredentials
	}

	token, err := GenerateToken(author.ID, author.Email, s.secret)
	if err != nil {
		return AuthResponse{}, fmt.Errorf("generate token: %w", err)
	}

	return AuthResponse{
		Token:  token,
		Author: AuthorPublic{ID: author.ID, Email: author.Email},
	}, nil
}

func (s *Service) ValidateToken(tokenString string) (*Claims, error) {
	return ParseToken(tokenString, s.secret)
}

func validateEmail(email string) (string, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	if email == "" {
		return "", ErrInvalidEmail
	}
	if _, err := mail.ParseAddress(email); err != nil {
		return "", ErrInvalidEmail
	}
	return email, nil
}

func validatePassword(password string) error {
	if len(password) < 8 {
		return ErrPasswordTooShort
	}
	return nil
}

func isSQLiteUniqueConstraint(err error) bool {
	var sqliteErr sqlite3.Error
	return errors.As(err, &sqliteErr) && sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique
}
