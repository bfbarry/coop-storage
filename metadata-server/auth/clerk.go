package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/clerk/clerk-sdk-go/v2/jwt"
	"github.com/clerk/clerk-sdk-go/v2/jwks"
)

type ClerkProvider struct {
	jwksClient *jwks.Client
}

func NewClerkProvider(secretKey string) *ClerkProvider {
	// Create a JWKS client with the Clerk secret key
	// This client will be used to fetch the JSON Web Key Set for token verification
	jwksClient := jwks.NewClient(&clerk.ClientConfig{
		BackendConfig: clerk.BackendConfig{
			Key: &secretKey,
		},
	})

	return &ClerkProvider{
		jwksClient: jwksClient,
	}
}

func (p *ClerkProvider) ValidateToken(ctx context.Context, token string) (*UserIdentity, error) {
	if token == "" {
		return nil, fmt.Errorf("token is empty")
	}

	// Verify the JWT token using Clerk's JWT verification
	claims, err := jwt.Verify(ctx, &jwt.VerifyParams{
		Token:      token,
		JWKSClient: p.jwksClient,
		Leeway:     10 * time.Second,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to verify token: %w", err)
	}

	// Extract user ID from the Subject claim (this is the Clerk user ID)
	userID := claims.Subject
	if userID == "" {
		return nil, fmt.Errorf("token missing user ID (sub claim)")
	}

	// Build a map of custom claims for flexibility
	rawClaims := make(map[string]any)
	if claims.Custom != nil {
		if customMap, ok := claims.Custom.(map[string]any); ok {
			rawClaims = customMap
		}
	}

	// Try to extract email and name from custom claims if present
	email := ""
	if emailVal, ok := rawClaims["email"].(string); ok {
		email = emailVal
	}

	name := ""
	if nameVal, ok := rawClaims["name"].(string); ok {
		name = nameVal
	}

	return &UserIdentity{
		UserID:    userID,
		Email:     email,
		Name:      name,
		RawClaims: rawClaims,
	}, nil
}
