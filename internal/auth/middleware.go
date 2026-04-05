package auth

import (
	"net/http"

	"github.com/butschster/mcp-research/internal/domain"
)

// TokenValidator resolves a bearer token to a user.
type TokenValidator interface {
	ValidateToken(r *http.Request) (*domain.User, error)
}

// tokenValidatorFunc adapts a function to TokenValidator.
type tokenValidatorFunc func(r *http.Request) (*domain.User, error)

func (f tokenValidatorFunc) ValidateToken(r *http.Request) (*domain.User, error) {
	return f(r)
}

// RequireAuth returns middleware that rejects unauthenticated requests with 401.
func RequireAuth(validator TokenValidator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, err := validator.ValidateToken(r)
			if err != nil || user == nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"error":"unauthorized"}`))
				return
			}
			ctx := WithUser(r.Context(), user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// OptionalAuth sets user in context if token is present, but doesn't reject.
func OptionalAuth(validator TokenValidator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, _ := validator.ValidateToken(r)
			if user != nil {
				ctx := WithUser(r.Context(), user)
				r = r.WithContext(ctx)
			}
			next.ServeHTTP(w, r)
		})
	}
}
