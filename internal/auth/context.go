package auth

import (
	"context"

	"github.com/dovod-app/app/internal/domain"
)

type contextKey struct{}

// WithUser stores the authenticated user in the context.
func WithUser(ctx context.Context, user *domain.User) context.Context {
	return context.WithValue(ctx, contextKey{}, user)
}

// UserFromContext returns the authenticated user, or nil if not authenticated.
func UserFromContext(ctx context.Context) *domain.User {
	u, _ := ctx.Value(contextKey{}).(*domain.User)
	return u
}

// UserIDFromContext returns the user ID or "" if not authenticated.
func UserIDFromContext(ctx context.Context) string {
	if u := UserFromContext(ctx); u != nil {
		return u.ID
	}
	return ""
}
