package api

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/Sosokker/todolist-backend/internal/config"
	"github.com/Sosokker/todolist-backend/internal/domain"
	"github.com/Sosokker/todolist-backend/internal/service"
	"github.com/google/uuid"
)

type contextKey string

const UserIDKey contextKey = "userID"

var publicPaths = map[string]bool{
	"/auth/signup":          true,
	"/auth/login":           true,
	"/auth/google/login":    true,
	"/auth/google/callback": true,
}

func AuthMiddleware(authService service.AuthService, cfg *config.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestPath := r.URL.Path
			basePath := cfg.Server.BasePath
			relativePath := strings.TrimPrefix(requestPath, basePath)

			if _, isPublic := publicPaths[relativePath]; isPublic {
				slog.DebugContext(r.Context(), "Public path accessed, skipping auth", "path", requestPath)
				next.ServeHTTP(w, r)
				return
			}

			tokenString := extractToken(r, cfg)
			if tokenString == "" {
				slog.WarnContext(r.Context(), "Authentication failed: missing token", "path", requestPath)
				SendJSONError(w, domain.ErrUnauthorized, http.StatusUnauthorized, slog.Default())
				return
			}

			claims, err := authService.ValidateJWT(tokenString)
			if err != nil {
				slog.WarnContext(r.Context(), "Authentication failed: invalid token", "error", err, "path", requestPath)
				SendJSONError(w, domain.ErrUnauthorized, http.StatusUnauthorized, slog.Default())
				return
			}

			if claims.ID == uuid.Nil {
				slog.ErrorContext(r.Context(), "Authentication failed: Nil User ID in token claims", "path", requestPath)
				SendJSONError(w, domain.ErrUnauthorized, http.StatusUnauthorized, slog.Default())
				return
			}

			ctx := context.WithValue(r.Context(), UserIDKey, claims.ID)
			slog.DebugContext(ctx, "Authentication successful", "userId", claims.ID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func extractToken(r *http.Request, cfg *config.Config) string {
	bearerToken := r.Header.Get("Authorization")
	if parts := strings.Split(bearerToken, " "); len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
		slog.DebugContext(r.Context(), "Token found in Authorization header")
		return parts[1]
	}

	cookie, err := r.Cookie(cfg.JWT.CookieName)
	if err == nil {
		slog.DebugContext(r.Context(), "Token found in cookie", "cookieName", cfg.JWT.CookieName)
		return cookie.Value
	}
	if !errors.Is(err, http.ErrNoCookie) {
		slog.WarnContext(r.Context(), "Error reading auth cookie", "error", err, "cookieName", cfg.JWT.CookieName)
	} else {
		slog.DebugContext(r.Context(), "No token found in header or cookie")
	}

	return ""
}

func GetUserIDFromContext(ctx context.Context) (uuid.UUID, error) {
	userIDVal := ctx.Value(UserIDKey)
	if userIDVal == nil {
		slog.ErrorContext(ctx, "User ID not found in context. Middleware might not have run or failed.")
		return uuid.Nil, fmt.Errorf("user ID not found in context: %w", domain.ErrInternalServer)
	}

	userID, ok := userIDVal.(uuid.UUID)
	if !ok {
		slog.ErrorContext(ctx, "User ID in context has unexpected type", "type", fmt.Sprintf("%T", userIDVal))
		return uuid.Nil, fmt.Errorf("user ID in context has unexpected type: %w", domain.ErrInternalServer)
	}

	if userID == uuid.Nil {
		slog.ErrorContext(ctx, "Nil User ID found in context after authentication")
		return uuid.Nil, fmt.Errorf("invalid user ID (nil) found in context: %w", domain.ErrInternalServer)
	}

	return userID, nil
}
