package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/Sosokker/todolist-backend/internal/auth"
	"github.com/Sosokker/todolist-backend/internal/config"
	"github.com/Sosokker/todolist-backend/internal/domain"
	"github.com/Sosokker/todolist-backend/internal/repository"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/oauth2"
)

type authService struct {
	userRepo        repository.UserRepository
	cfg             *config.Config
	googleOAuthProv auth.OAuthProvider
	logger          *slog.Logger
}

func NewAuthService(repo repository.UserRepository, cfg *config.Config) AuthService {
	logger := slog.Default().With("service", "auth")
	googleProvider := auth.NewGoogleOAuthProvider(cfg)
	return &authService{
		userRepo:        repo,
		cfg:             cfg,
		googleOAuthProv: googleProvider,
		logger:          logger,
	}
}

func (s *authService) Signup(ctx context.Context, creds SignupCredentials) (*domain.User, error) {
	if err := ValidateSignupInput(creds); err != nil {
		return nil, err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(creds.Password), bcrypt.DefaultCost)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to hash password", "error", err)
		return nil, domain.ErrInternalServer
	}

	newUser := &domain.User{
		Username:      creds.Username,
		Email:         creds.Email,
		PasswordHash:  string(hashedPassword),
		EmailVerified: false,
	}

	createdUser, err := s.userRepo.Create(ctx, newUser)
	if err != nil {
		if errors.Is(err, domain.ErrConflict) {
			_, emailErr := s.userRepo.GetByEmail(ctx, creds.Email)
			if emailErr == nil {
				return nil, fmt.Errorf("email already exists: %w", domain.ErrConflict)
			}
			return nil, fmt.Errorf("username already exists: %w", domain.ErrConflict)
		}
		slog.ErrorContext(ctx, "Failed to create user in db", "error", err)
		return nil, domain.ErrInternalServer
	}

	return createdUser, nil
}

func (s *authService) Login(ctx context.Context, creds LoginCredentials) (string, *domain.User, error) {
	if err := ValidateLoginInput(creds); err != nil {
		return "", nil, err
	}

	user, err := s.userRepo.GetByEmail(ctx, creds.Email)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return "", nil, fmt.Errorf("invalid email or password: %w", domain.ErrUnauthorized)
		}
		slog.ErrorContext(ctx, "Failed to get user by email", "error", err)
		return "", nil, domain.ErrInternalServer
	}

	if user.PasswordHash == "" && user.GoogleID != nil {
		return "", nil, fmt.Errorf("please log in using Google: %w", domain.ErrUnauthorized)
	}
	if user.PasswordHash == "" {
		slog.ErrorContext(ctx, "User found with empty password hash", "userId", user.ID)
		return "", nil, fmt.Errorf("account error, please contact support: %w", domain.ErrInternalServer)
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(creds.Password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return "", nil, fmt.Errorf("invalid email or password: %w", domain.ErrUnauthorized)
		}
		slog.ErrorContext(ctx, "Error comparing password hash", "error", err, "userId", user.ID)
		return "", nil, domain.ErrInternalServer
	}

	token, err := s.GenerateJWT(user)
	if err != nil {
		return "", nil, err
	}

	return token, user, nil
}

func (s *authService) GenerateJWT(user *domain.User) (string, error) {
	expirationTime := time.Now().Add(time.Duration(s.cfg.JWT.ExpiryMinutes) * time.Minute)
	claims := &auth.Claims{
		UserID: user.ID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   user.ID.String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.cfg.JWT.Secret))
	if err != nil {
		slog.Error("Failed to sign JWT token", "error", err, "userId", user.ID)
		return "", domain.ErrInternalServer
	}
	return tokenString, nil
}

func (s *authService) ValidateJWT(tokenString string) (*domain.User, error) {
	claims := &auth.Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.cfg.JWT.Secret), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, fmt.Errorf("token has expired: %w", domain.ErrUnauthorized)
		}
		if errors.Is(err, jwt.ErrTokenMalformed) {
			return nil, fmt.Errorf("token is malformed: %w", domain.ErrUnauthorized)
		}
		slog.Warn("JWT validation failed", "error", err)
		return nil, fmt.Errorf("invalid token: %w", domain.ErrUnauthorized)
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token: %w", domain.ErrUnauthorized)
	}

	user, err := s.userRepo.GetByID(context.Background(), claims.UserID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, fmt.Errorf("user associated with token not found: %w", domain.ErrUnauthorized)
		}
		slog.Error("Failed to fetch user for valid JWT", "error", err, "userId", claims.UserID)
		return nil, domain.ErrInternalServer
	}

	return user, nil
}

func (s *authService) GetGoogleAuthConfig() *oauth2.Config {
	return s.googleOAuthProv.GetOAuth2Config()
}

type GoogleUserInfo struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
}

func (s *authService) HandleGoogleCallback(ctx context.Context, code string) (string, *domain.User, error) {
	token, err := s.googleOAuthProv.ExchangeCode(ctx, code)
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to exchange google auth code via provider", "error", err)
		return "", nil, fmt.Errorf("google auth exchange failed: %w", domain.ErrUnauthorized)
	}

	userInfo, err := s.googleOAuthProv.FetchUserInfo(ctx, token)
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to fetch google user info via provider", "error", err)
		return "", nil, fmt.Errorf("failed to get user info from google: %w", domain.ErrUnauthorized)
	}

	if !userInfo.VerifiedEmail {
		return "", nil, fmt.Errorf("google email not verified: %w", domain.ErrUnauthorized)
	}

	user, err := s.userRepo.GetByGoogleID(ctx, userInfo.ID)
	if err != nil && !errors.Is(err, domain.ErrNotFound) {
		slog.ErrorContext(ctx, "Failed to check user by google ID", "error", err, "googleId", userInfo.ID)
		return "", nil, domain.ErrInternalServer
	}

	if user != nil {
		jwtToken, jwtErr := s.GenerateJWT(user)
		if jwtErr != nil {
			return "", nil, jwtErr
		}
		return jwtToken, user, nil
	}

	user, err = s.userRepo.GetByEmail(ctx, userInfo.Email)
	if err != nil && !errors.Is(err, domain.ErrNotFound) {
		slog.ErrorContext(ctx, "Failed to check user by email during google callback", "error", err, "email", userInfo.Email)
		return "", nil, domain.ErrInternalServer
	}

	if user != nil {
		if user.GoogleID != nil && *user.GoogleID != userInfo.ID {
			slog.WarnContext(ctx, "User email associated with different Google ID", "userId", user.ID, "existingGoogleId", *user.GoogleID, "newGoogleId", userInfo.ID)
			return "", nil, fmt.Errorf("email already linked to a different Google account: %w", domain.ErrConflict)
		}
		if user.GoogleID == nil {
			updateData := &domain.User{GoogleID: &userInfo.ID, EmailVerified: true}
			updatedUser, updateErr := s.userRepo.Update(ctx, user.ID, updateData)
			if updateErr != nil {
				slog.ErrorContext(ctx, "Failed to link Google ID to existing user", "error", updateErr, "userId", user.ID)
				return "", nil, domain.ErrInternalServer
			}
			user = updatedUser
		}

		jwtToken, jwtErr := s.GenerateJWT(user)
		if jwtErr != nil {
			return "", nil, jwtErr
		}
		return jwtToken, user, nil
	}

	newUser := &domain.User{
		Username:      userInfo.Name,
		Email:         userInfo.Email,
		PasswordHash:  "",
		EmailVerified: true,
		GoogleID:      &userInfo.ID,
	}

	createdUser, err := s.userRepo.Create(ctx, newUser)
	if err != nil {
		if errors.Is(err, domain.ErrConflict) {
			return "", nil, fmt.Errorf("failed to create user, potential conflict: %w", domain.ErrConflict)
		}
		slog.ErrorContext(ctx, "Failed to create new user from google info", "error", err)
		return "", nil, domain.ErrInternalServer
	}

	jwtToken, jwtErr := s.GenerateJWT(createdUser)
	if jwtErr != nil {
		return "", nil, jwtErr
	}
	return jwtToken, createdUser, nil
}
