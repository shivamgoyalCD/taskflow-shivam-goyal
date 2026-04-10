package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

const passwordHashCost = 12

var ErrInvalidCredentials = errors.New("invalid credentials")

type RegisterInput struct {
	Name     string
	Email    string
	Password string
}

type LoginInput struct {
	Email    string
	Password string
}

type AuthResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

type Service struct {
	repository Repository
	jwtManager *JWTManager
}

func NewService(repository Repository, jwtManager *JWTManager) *Service {
	return &Service{
		repository: repository,
		jwtManager: jwtManager,
	}
}

func (s *Service) Register(ctx context.Context, input RegisterInput) (AuthResponse, error) {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(input.Password), passwordHashCost)
	if err != nil {
		return AuthResponse{}, fmt.Errorf("hash password: %w", err)
	}

	user, err := s.repository.CreateUser(ctx, CreateUserParams{
		ID:           uuid.NewString(),
		Name:         strings.TrimSpace(input.Name),
		Email:        normalizeEmail(input.Email),
		PasswordHash: string(passwordHash),
	})
	if err != nil {
		return AuthResponse{}, err
	}

	token, err := s.jwtManager.GenerateToken(user)
	if err != nil {
		return AuthResponse{}, fmt.Errorf("generate register token: %w", err)
	}

	return AuthResponse{
		Token: token,
		User:  user,
	}, nil
}

func (s *Service) Login(ctx context.Context, input LoginInput) (AuthResponse, error) {
	user, err := s.repository.GetByEmail(ctx, normalizeEmail(input.Email))
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return AuthResponse{}, ErrInvalidCredentials
		}

		return AuthResponse{}, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		return AuthResponse{}, ErrInvalidCredentials
	}

	token, err := s.jwtManager.GenerateToken(user)
	if err != nil {
		return AuthResponse{}, fmt.Errorf("generate login token: %w", err)
	}

	return AuthResponse{
		Token: token,
		User:  user,
	}, nil
}

func (s *Service) ListUsers(ctx context.Context) ([]User, error) {
	users, err := s.repository.ListUsers(ctx)
	if err != nil {
		return nil, err
	}
	if users == nil {
		users = make([]User, 0)
	}

	return users, nil
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
