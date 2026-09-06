package service

import (
	"context"
	"crypto/subtle"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/dovod-app/app/internal/auth"
	"github.com/dovod-app/app/internal/domain"
	"github.com/dovod-app/app/internal/storage"
	"github.com/google/uuid"
)

var (
	// ErrShareUnavailable is the single answer for every reason a link does not
	// work: revoked, expired, never existed, or malformed. Telling them apart
	// would turn the URL bar into an oracle — "this token was valid once" and
	// "this research exists" are both information a stranger has no claim to.
	ErrShareUnavailable = errors.New("this link is no longer available")
	// ErrShareLocked is returned when the link is real and the password is
	// missing or wrong. It is deliberately distinct from the above: whoever is
	// holding a working URL already knows the link exists, and a password
	// prompt that renders as "not found" is unusable.
	ErrShareLocked = errors.New("this link is password protected")
	// ErrShareScope guards the columns that exist for narrower shares but are
	// not issued yet. Refusing is better than enforcing them halfway.
	ErrShareScope = errors.New("only a whole-research share can be created")
	// ErrSharePassword rejects a password too short to be worth the prompt.
	ErrSharePassword = errors.New("a share password must be at least 6 characters")
	// ErrShareInactive refuses an edit to a link that has been revoked or has
	// expired. It is deliberately distinguishable — this is the owner's own
	// management surface, where "the link died while you had the dialog open"
	// is the useful answer, unlike the public prefix where every refusal must
	// look the same. It wraps ErrConflict so the handlers map it to 409.
	ErrShareInactive = fmt.Errorf("%w: this link is no longer live", ErrConflict)
)

// shareMaxTTLDays bounds how far ahead a link may be set to expire. A share
// that outlives the project it was made for is the failure mode here, and ten
// years of "never noticed it was still live" is not a feature.
const shareMaxTTLDays = 3650

type CreateShareRequest struct {
	Label   string
	Include domain.ShareInclude
	// ExpiresInDays is nil for a link that never expires on its own. That is a
	// real choice — a published report should not break — and it is why
	// revocation has to work regardless of expiry.
	ExpiresInDays *int
	Password      string
}

// ShareResult carries the token, and it is the only place the token ever
// appears. It is returned from Create and from nowhere else: the database holds
// a SHA-256, and there is no route that can turn one back into a link.
type ShareResult struct {
	Share *domain.Share
	Token string
}

type ShareService struct {
	shares *storage.ShareRepository
	access *Access
	events EventNotifier
	log    *slog.Logger
}

func NewShareService(shares *storage.ShareRepository, access *Access, events EventNotifier, log *slog.Logger) *ShareService {
	return &ShareService{shares: shares, access: access, events: events, log: log}
}

// Create issues a link. It takes write access rather than a new owner-only
// tier: an editor can already export the whole research to a file and send it,
// so a link is not a capability they did not have. A viewer cannot, and that
// matters — otherwise the weakest role in a team could republish everything the
// team owns.
func (s *ShareService) Create(ctx context.Context, researchID string, req CreateShareRequest) (*ShareResult, error) {
	if err := s.access.Write(ctx, researchID); err != nil {
		return nil, err
	}

	password := strings.TrimSpace(req.Password)
	if password != "" && len(password) < 6 {
		return nil, ErrSharePassword
	}
	passwordHash := ""
	if password != "" {
		h, err := auth.HashPassword(password)
		if err != nil {
			return nil, err
		}
		passwordHash = h
	}

	var expires *time.Time
	if req.ExpiresInDays != nil {
		days := *req.ExpiresInDays
		if days < 1 {
			days = 1
		}
		if days > shareMaxTTLDays {
			days = shareMaxTTLDays
		}
		t := time.Now().UTC().Add(time.Duration(days) * 24 * time.Hour)
		expires = &t
	}

	token, tokenHash := auth.GenerateShareToken()
	share := &domain.Share{
		ID:         uuid.New().String(),
		ResearchID: researchID,
		UserID:     auth.UserIDFromContext(ctx),
		Scope:      domain.ShareResearch,
		Label:      normalizeTitle(req.Label),
		Include:    req.Include,
		ExpiresAt:  expires,
	}
	if err := s.shares.Create(ctx, share, tokenHash, passwordHash); err != nil {
		return nil, err
	}

	emit(ctx, s.events, Event{Type: "share.created", ResearchID: researchID, EntityID: share.ID, Entity: "share"})
	return &ShareResult{Share: share, Token: token}, nil
}

// List returns every share ever made for a research, live and dead. It takes
// the same write access Create does: the pair "hand out a link" and "take it
// back" belongs to the same people, and a list you cannot act on is noise.
func (s *ShareService) List(ctx context.Context, researchID string) ([]*domain.Share, error) {
	if err := s.access.Write(ctx, researchID); err != nil {
		return nil, err
	}
	shares, err := s.shares.ListByResearch(ctx, researchID)
	if err != nil {
		return nil, err
	}
	if shares == nil {
		shares = []*domain.Share{}
	}
	return shares, nil
}

// ActiveCount is how many live links a research has, for the research page.
//
// It answers zero rather than an error for anyone who could not manage them
// anyway: a viewer, or a share visitor reading the page through one of the
// links being counted. Neither should be told the number, and neither has a
// control to act on it.
func (s *ShareService) ActiveCount(ctx context.Context, researchID string) int {
	if err := s.access.Write(ctx, researchID); err != nil {
		return 0
	}
	n, err := s.shares.CountActive(ctx, researchID)
	if err != nil {
		return 0
	}
	return n
}

// Revoke kills a link immediately. Everything that consults a share does so per
// request, so there is no cached verdict to wait out — except the WebSocket,
// which re-checks on its own timer and closes within the minute.
func (s *ShareService) Revoke(ctx context.Context, shareID string) error {
	share, err := s.shares.FindByID(ctx, shareID)
	if err != nil {
		return err
	}
	if share == nil {
		return ErrNotFound
	}
	if err := s.access.Write(ctx, share.ResearchID); err != nil {
		return err
	}
	if err := s.shares.Revoke(ctx, shareID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// Already revoked. Saying so is better than a 500 and better than a
			// silent success that moves the timestamp.
			return ErrNotFound
		}
		return err
	}
	emit(ctx, s.events, Event{Type: "share.revoked", ResearchID: share.ResearchID, EntityID: shareID, Entity: "share"})
	return nil
}

// UpdateShareRequest is what an owner may change about a link after issuing it.
//
// Include is the whole set of flags, not a patch: a body naming one flag that
// silently left the other three at false would narrow a link the owner meant
// to widen. A nil Include leaves the flags alone; a nil Label leaves the label.
type UpdateShareRequest struct {
	Label   *string
	Include *domain.ShareInclude
}

// Update changes a live link's label or contents without reissuing it.
//
// It exists because a published link is an address people have already
// saved: revoking it to change what it shows breaks every place it was pasted.
// Widening is the sensitive direction — everyone holding the URL sees more, at
// once — and that is a decision for the person, confirmed in the UI; the
// service only requires the same write access Create and Revoke do.
func (s *ShareService) Update(ctx context.Context, shareID string, req UpdateShareRequest) (*domain.Share, error) {
	share, err := s.shares.FindByID(ctx, shareID)
	if err != nil {
		return nil, err
	}
	if share == nil {
		return nil, ErrNotFound
	}
	if err := s.access.Write(ctx, share.ResearchID); err != nil {
		return nil, err
	}
	if !share.Active(time.Now().UTC()) {
		return nil, ErrShareInactive
	}

	label := share.Label
	if req.Label != nil {
		label = normalizeTitle(*req.Label)
	}
	include := share.Include
	if req.Include != nil {
		include = *req.Include
	}
	if err := s.shares.Update(ctx, shareID, label, include); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// Revoked between the read above and the write.
			return nil, ErrShareInactive
		}
		return nil, err
	}

	updated, err := s.shares.FindByID(ctx, shareID)
	if err != nil {
		return nil, err
	}
	if updated == nil {
		return nil, ErrNotFound
	}
	emit(ctx, s.events, Event{Type: "share.updated", ResearchID: share.ResearchID, EntityID: shareID, Entity: "share"})
	return updated, nil
}

// Resolve turns a token into the capability it carries, or refuses.
//
// unlock is the value a visitor got back from Unlock, and is checked only when
// the share has a password. Everything else — revoked, expired, unknown — is
// one indistinguishable ErrShareUnavailable.
func (s *ShareService) Resolve(ctx context.Context, token, unlock string) (*domain.Share, error) {
	share, err := s.find(ctx, token)
	if err != nil {
		return nil, err
	}
	if share.HasPassword {
		hash, err := s.shares.PasswordHash(ctx, share.ID)
		if err != nil {
			return nil, err
		}
		want := unlockValue(share.ID, hash)
		if unlock == "" || subtle.ConstantTimeCompare([]byte(unlock), []byte(want)) != 1 {
			return nil, ErrShareLocked
		}
	}
	return share, nil
}

// Unlock exchanges the password for a value the visitor sends with each
// subsequent request.
//
// It is derived from the share id and the stored hash rather than being a row
// of its own: it then costs no storage, dies with the share, and stops working
// the moment the password changes. Its power is exactly the password's, which
// is the right amount — it is not a session, and there is nobody to have one.
func (s *ShareService) Unlock(ctx context.Context, token, password string) (string, error) {
	share, err := s.find(ctx, token)
	if err != nil {
		return "", err
	}
	if !share.HasPassword {
		// Nothing to unlock. Answering with a usable value keeps the client
		// from having to special-case a link that lost its password.
		return "", nil
	}
	hash, err := s.shares.PasswordHash(ctx, share.ID)
	if err != nil {
		return "", err
	}
	if auth.CheckPassword(hash, password) != nil {
		return "", ErrShareLocked
	}
	return unlockValue(share.ID, hash), nil
}

// Seen records that the link was opened. Called once per page load rather than
// once per request, so the count means visits and not fetches.
func (s *ShareService) Seen(ctx context.Context, shareID string) {
	s.shares.TouchSeen(ctx, shareID)
}

// Describe returns a share by id for a caller who already holds the capability
// — the visitor whose request the id came out of, and nobody else.
//
// It takes no permission because there is none to take: whoever presented the
// token has already been granted this exact share, and the only thing this adds
// is the label and the creator's name for the banner. A nil result is a share
// that vanished between two queries, which the caller renders as "no metadata"
// rather than as an error.
func (s *ShareService) Describe(ctx context.Context, shareID string) *domain.Share {
	share, err := s.shares.FindByID(ctx, shareID)
	if err != nil {
		return nil
	}
	return share
}

// Capability converts a resolved share into what the request layers read.
func Capability(share *domain.Share) *auth.Share {
	return &auth.Share{ID: share.ID, ResearchID: share.ResearchID, Include: share.Include}
}

// Scope is Resolve with the reasons collapsed, for the WebSocket.
//
// A socket has no status codes to distinguish "revoked" from "wrong password"
// with, and re-checks itself on a timer where a nil means "close this
// connection". Both the handshake and that re-check ask this.
func (s *ShareService) Scope(ctx context.Context, token, unlock string) *auth.Share {
	share, err := s.Resolve(ctx, token, unlock)
	if err != nil {
		return nil
	}
	return Capability(share)
}

func (s *ShareService) find(ctx context.Context, token string) (*domain.Share, error) {
	if token == "" {
		return nil, ErrShareUnavailable
	}
	share, err := s.shares.FindByHash(ctx, auth.HashAPIKey(token))
	if err != nil {
		return nil, err
	}
	if share == nil || !share.Active(time.Now().UTC()) {
		return nil, ErrShareUnavailable
	}
	if share.Scope != domain.ShareResearch {
		// A scope this build does not enforce is a scope it must not serve.
		return nil, ErrShareUnavailable
	}
	return share, nil
}

// unlockValue is domain-separated so it can never be mistaken for, or replayed
// as, any other hashed secret in the product.
func unlockValue(shareID, passwordHash string) string {
	if passwordHash == "" {
		return ""
	}
	return auth.HashAPIKey("share-unlock:" + shareID + ":" + passwordHash)
}
