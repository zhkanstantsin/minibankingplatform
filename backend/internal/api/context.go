package api

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"minibankingplatform/pkg/jwt"
)

type contextKey string

// UserClaimsKey is the context key for storing user claims.
const UserClaimsKey contextKey = "user_claims"

// ErrNoClaims is returned when no claims are found in context.
var ErrNoClaims = errors.New("no claims found in context")

// ContextWithClaims returns a new context with the given claims.
func ContextWithClaims(ctx context.Context, claims *jwt.Claims) context.Context {
	return context.WithValue(ctx, UserClaimsKey, claims)
}

// ClaimsFromContext retrieves claims from the context.
func ClaimsFromContext(ctx context.Context) (*jwt.Claims, error) {
	claims, ok := ctx.Value(UserClaimsKey).(*jwt.Claims)
	if !ok || claims == nil {
		return nil, ErrNoClaims
	}
	return claims, nil
}

// UserIDFromContext retrieves the user ID from context claims.
func UserIDFromContext(ctx context.Context) (uuid.UUID, error) {
	claims, err := ClaimsFromContext(ctx)
	if err != nil {
		return uuid.UUID{}, err
	}
	return claims.UserID, nil
}

// UserEmailFromContext retrieves the user email from context claims.
func UserEmailFromContext(ctx context.Context) (string, error) {
	claims, err := ClaimsFromContext(ctx)
	if err != nil {
		return "", err
	}
	return claims.Email, nil
}
