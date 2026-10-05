package auth

import (
	"context"
	"log"
	"net/http"
	"strings"
)

type contextKey string               //private type so no other package can collide with this key
const iD_KEY contextKey = "identity" // TODO could expand on this

type AuthProvider interface {
	// validates token from request, returns normalized identity
	ValidateToken(ctx context.Context, token string) (*UserIdentity, error)
}

type UserIdentity struct {
	UserID    string
	Email     string
	Name      string
	RawClaims map[string]any // ? TBD
}

func extractBearerToken(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return ""
	}

	// Expected format: "Bearer <token>"
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return ""
	}

	return parts[1]
}

func AuthMiddleWare(provider AuthProvider) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := extractBearerToken(r)
			identity, err := provider.ValidateToken(r.Context(), token)
			if err != nil {
				log.Printf("auth: token validation failed: %v", err)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			ctx := context.WithValue(r.Context(), iD_KEY, identity)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetUserIdentity retrieves the authenticated user identity from the request context.
// Returns nil if no identity is present (e.g., unauthenticated request).
func GetUserIdentity(ctx context.Context) *UserIdentity {
	identity, ok := ctx.Value(iD_KEY).(*UserIdentity)
	if !ok {
		return nil
	}
	return identity
}
