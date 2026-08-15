package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"github.com/butschster/mcp-research/internal/domain"
)

// sharePrefix marks a token as a share link, so one pasted into the wrong field
// is recognisable rather than mysterious.
const sharePrefix = "mrs_"

// GenerateShareToken returns (plaintext, hash). The plaintext is shown once, at
// creation; only the hash is stored, so a leaked database hands out no working
// links. 256 bits of randomness is what makes the URL itself the credential.
func GenerateShareToken() (string, string) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		panic(fmt.Sprintf("crypto/rand: %v", err))
	}
	plain := sharePrefix + hex.EncodeToString(b)
	return plain, HashAPIKey(plain)
}

// Share is the capability a share link carries into a request. It authorizes a
// scope, not a person: there is no user, no account and nothing to sign in as.
//
// It is deliberately a separate context value from the user rather than a
// synthetic User. Every handler in this product decides what to do by reading
// the user, and a handler that reads the user and finds nobody already refuses
// — which is what makes an unmodified handler safe when a share reaches it.
// Fabricating a user here would have inverted that: every existing route would
// have started letting share visitors in, silently.
type Share struct {
	// ID identifies the row, for the view bookkeeping and for revocation.
	ID string
	// ResearchID is the single research this token reaches. Everything else in
	// the product is out of scope by construction.
	ResearchID string
	// Include is what the creator chose to expose beyond the content itself.
	Include domain.ShareInclude
}

type shareContextKey struct{}

// WithShare stores the share capability in the context.
func WithShare(ctx context.Context, sh *Share) context.Context {
	return context.WithValue(ctx, shareContextKey{}, sh)
}

// ShareFromContext returns the share capability, or nil when the caller is not
// a share visitor.
func ShareFromContext(ctx context.Context) *Share {
	sh, _ := ctx.Value(shareContextKey{}).(*Share)
	return sh
}
