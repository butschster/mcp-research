package storage

import (
	"context"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/dovod-app/app/internal/domain"
	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

func contractUser(t *testing.T, db *bun.DB) *domain.User {
	t.Helper()
	id := uuid.NewString()
	u := &domain.User{ID: id, Email: id + "@test.invalid", Name: "O'Reilly ? 日本語", PasswordHash: "test"}
	if err := NewUserRepository(db).Create(context.Background(), u); err != nil {
		t.Fatal(err)
	}
	return u
}

func contractEntry(t *testing.T, db *bun.DB, tags []string) *domain.Entry {
	t.Helper()
	ctx := context.Background()
	r := &domain.Research{ID: uuid.NewString(), Name: "contract", TeamID: "team-local", Status: domain.ResearchActive}
	if err := NewResearchRepository(db).Create(ctx, r); err != nil {
		t.Fatal(err)
	}
	s := &domain.Section{ID: uuid.NewString(), ResearchID: r.ID, Name: "section", Status: domain.SectionDraft}
	if err := NewSectionRepository(db).Create(ctx, s); err != nil {
		t.Fatal(err)
	}
	e := &domain.Entry{ID: uuid.NewString(), ResearchID: r.ID, SectionID: s.ID, Title: "O'Reilly ?", Tags: tags}
	if err := NewEntryRepository(db).Create(ctx, e); err != nil {
		t.Fatal(err)
	}
	return e
}

func TestBunContract_APIKeysOwnershipAndLifecycle(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()
	owner, other := contractUser(t, db), contractUser(t, db)
	repo := NewAPIKeyRepository(db)
	key := &domain.APIKey{ID: uuid.NewString(), UserID: owner.ID, Name: "quoted '? key", KeyPrefix: "test-key"}
	if err := repo.Create(ctx, key, "hash-1"); err != nil {
		t.Fatal(err)
	}
	got, err := repo.FindByHash(ctx, "hash-1")
	if err != nil || got == nil || got.ID != key.ID || got.Name != key.Name || got.LastUsedAt != nil || got.ExpiresAt != nil {
		t.Fatalf("key roundtrip: %+v %v", got, err)
	}
	if err := repo.Delete(ctx, key.ID, other.ID); err == nil {
		t.Fatal("non-owner deleted a key")
	}
	if keys, err := repo.ListByUser(ctx, other.ID); err != nil || len(keys) != 0 {
		t.Fatalf("cross-user key leak: %+v %v", keys, err)
	}
	repo.TouchLastUsed(ctx, key.ID)
	expires := time.Now().UTC().Add(time.Hour).Truncate(time.Second)
	if _, err := db.NewUpdate().Table("api_keys").Set("expires_at=?", expires.Format(time.DateTime)).Where("id=?", key.ID).Exec(ctx); err != nil {
		t.Fatal(err)
	}
	keys, err := repo.ListByUser(ctx, owner.ID)
	if err != nil || len(keys) != 1 || keys[0].LastUsedAt == nil || keys[0].ExpiresAt == nil || !keys[0].ExpiresAt.Equal(expires) {
		t.Fatalf("key timestamps: %+v %v", keys, err)
	}
	got, err = repo.FindByHash(ctx, "hash-1")
	if err != nil || got == nil || got.LastUsedAt == nil || got.ExpiresAt == nil || !got.ExpiresAt.Equal(expires) {
		t.Fatalf("key lookup timestamps: %+v %v", got, err)
	}
	duplicate := &domain.APIKey{ID: uuid.NewString(), UserID: other.ID, Name: "duplicate"}
	if err := repo.Create(ctx, duplicate, "hash-1"); err == nil {
		t.Fatal("duplicate hash accepted")
	}
	if err := repo.Delete(ctx, key.ID, owner.ID); err != nil {
		t.Fatal(err)
	}
	if got, err := repo.FindByHash(ctx, "hash-1"); err != nil || got != nil {
		t.Fatalf("deleted key still usable: %+v %v", got, err)
	}
}

func TestBunContract_OAuthLifecycle(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()
	owner, other := contractUser(t, db), contractUser(t, db)
	repo := NewOAuthRepository(db)
	client := &domain.OAuthClient{ID: uuid.NewString(), UserID: owner.ID, Name: "quoted '? client", RedirectURIs: []string{"https://example.invalid/callback?x=1&y=2", "http://localhost/callback"}}
	if err := repo.CreateClient(ctx, client, "secret-hash"); err != nil {
		t.Fatal(err)
	}
	got, secret, err := repo.FindClientByID(ctx, client.ID)
	if err != nil || got == nil || got.UserID != owner.ID || got.Name != client.Name || secret != "secret-hash" || !reflect.DeepEqual(got.RedirectURIs, client.RedirectURIs) {
		t.Fatalf("client roundtrip: %+v %v", got, err)
	}
	if list, err := repo.ListClientsByUser(ctx, owner.ID); err != nil || len(list) != 1 || list[0].ID != client.ID {
		t.Fatalf("client list: %+v %v", list, err)
	}
	if list, err := repo.ListClientsByUser(ctx, other.ID); err != nil || len(list) != 0 {
		t.Fatalf("cross-user client leak: %+v %v", list, err)
	}
	public := &domain.OAuthClient{ID: uuid.NewString(), Name: "public"}
	if err := repo.CreateClient(ctx, public, ""); err != nil {
		t.Fatal(err)
	}
	if got, _, err := repo.FindClientByID(ctx, public.ID); err != nil || got == nil || got.UserID != "" {
		t.Fatalf("nullable client owner: %+v %v", got, err)
	}
	if got, _, err := repo.FindClientByID(ctx, "missing"); err != nil || got != nil {
		t.Fatalf("missing client: %+v %v", got, err)
	}

	future := time.Now().UTC().Add(time.Hour).Truncate(time.Second)
	past := future.Add(-2 * time.Hour)
	code := OAuthCode{Code: "code-1", ClientID: client.ID, UserID: owner.ID, RedirectURI: client.RedirectURIs[0], Scope: "research:read", CodeChallenge: "challenge", CodeChallengeMethod: "S256", ExpiresAt: future}
	if err := repo.CreateCode(ctx, &code); err != nil {
		t.Fatal(err)
	}
	if got, err := repo.FindCode(ctx, code.Code); err != nil || got == nil || !reflect.DeepEqual(*got, code) {
		t.Fatalf("code roundtrip: %+v %v", got, err)
	}
	expired := code
	expired.Code, expired.ExpiresAt = "expired-code", past
	if err := repo.CreateCode(ctx, &expired); err != nil {
		t.Fatal(err)
	}
	if err := repo.CleanExpiredCodes(ctx); err != nil {
		t.Fatal(err)
	}
	if got, err := repo.FindCode(ctx, expired.Code); err != nil || got != nil {
		t.Fatalf("expired code retained: %+v %v", got, err)
	}
	if got, err := repo.FindCode(ctx, code.Code); err != nil || got == nil {
		t.Fatalf("live code removed: %+v %v", got, err)
	}
	if err := repo.DeleteCode(ctx, code.Code); err != nil {
		t.Fatal(err)
	}
	if got, err := repo.FindCode(ctx, code.Code); err != nil || got != nil {
		t.Fatalf("deleted code retained: %+v %v", got, err)
	}

	token := OAuthToken{ID: uuid.NewString(), ClientID: client.ID, UserID: owner.ID, AccessTokenHash: "access-hash", RefreshTokenHash: "refresh-hash", Scope: "research:read", ExpiresAt: future}
	if err := repo.CreateToken(ctx, &token); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		lookup func(context.Context, string) (*OAuthToken, error)
		hash   string
	}{
		{repo.FindByAccessTokenHash, token.AccessTokenHash},
		{repo.FindByRefreshTokenHash, token.RefreshTokenHash},
	} {
		if got, err := tc.lookup(ctx, tc.hash); err != nil || got == nil || !reflect.DeepEqual(*got, token) {
			t.Fatalf("token roundtrip: %+v %v", got, err)
		}
	}
	oldToken := token
	oldToken.ID, oldToken.AccessTokenHash, oldToken.RefreshTokenHash, oldToken.ExpiresAt = uuid.NewString(), "old-access", "old-refresh", past
	if err := repo.CreateToken(ctx, &oldToken); err != nil {
		t.Fatal(err)
	}
	if err := repo.CleanExpiredTokens(ctx); err != nil {
		t.Fatal(err)
	}
	if got, err := repo.FindByAccessTokenHash(ctx, oldToken.AccessTokenHash); err != nil || got != nil {
		t.Fatalf("expired token retained: %+v %v", got, err)
	}
	if got, err := repo.FindByAccessTokenHash(ctx, token.AccessTokenHash); err != nil || got == nil {
		t.Fatalf("live token removed: %+v %v", got, err)
	}
	if err := repo.DeleteToken(ctx, token.ID); err != nil {
		t.Fatal(err)
	}
	if got, err := repo.FindByAccessTokenHash(ctx, token.AccessTokenHash); err != nil || got != nil {
		t.Fatalf("revoked access token usable: %+v %v", got, err)
	}
	if got, err := repo.FindByRefreshTokenHash(ctx, token.RefreshTokenHash); err != nil || got != nil {
		t.Fatalf("revoked refresh token usable: %+v %v", got, err)
	}
}

func TestBunContract_TagsLinksAndBlockState(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()
	longTag := strings.Repeat("日本語?", 100)
	e := contractEntry(t, db, []string{"quoted'", longTag})
	other := contractEntry(t, db, []string{"other-research"})
	entries := NewEntryRepository(db)
	second := &domain.Entry{ID: uuid.NewString(), ResearchID: e.ResearchID, SectionID: e.SectionID, Tags: []string{longTag}}
	if err := entries.Create(ctx, second); err != nil {
		t.Fatal(err)
	}
	tags, err := entries.FindTagsByResearch(ctx, e.ResearchID)
	if err != nil || len(tags) != 2 || tags[0].Tag != longTag || tags[0].Count != 2 || tags[1].Tag != "quoted'" || tags[1].Count != 1 {
		t.Fatalf("tag aggregation: %+v %v", tags, err)
	}
	if found, err := entries.FindByResearch(ctx, e.ResearchID, EntryFilter{Tag: longTag}); err != nil || len(found) != 2 {
		t.Fatalf("long tag filter: %+v %v", found, err)
	}
	links := NewExternalLinkRepository(db)
	link := domain.ExternalLink{ID: uuid.NewString(), SourceType: "entry", SourceID: e.ID, ResearchID: e.ResearchID, URL: "https://example.invalid/?q='test'", Title: "quoted'", Domain: "example.invalid"}
	if err := links.ReplaceForSource(ctx, "entry", e.ID, []domain.ExternalLink{link}); err != nil {
		t.Fatal(err)
	}
	if got, err := links.FindByResearch(ctx, e.ResearchID); err != nil || len(got) != 1 || got[0].URL != link.URL || got[0].EntryCode != e.Code || got[0].EntryTitle != e.Title {
		t.Fatalf("links with entry join: %+v %v", got, err)
	}
	if got, err := links.FindByResearch(ctx, other.ResearchID); err != nil || len(got) != 0 {
		t.Fatalf("cross-research link leak: %+v %v", got, err)
	}
	// A duplicate in the replacement must roll back both its delete and insert.
	if err := links.ReplaceForSource(ctx, "entry", e.ID, []domain.ExternalLink{link, link}); err == nil {
		t.Fatal("duplicate replacement should fail")
	}
	if got, err := links.FindBySource(ctx, "entry", e.ID); err != nil || len(got) != 1 || got[0].ID != link.ID {
		t.Fatalf("failed replacement lost old links: %+v %v", got, err)
	}
	if err := links.ReplaceForSource(ctx, "entry", e.ID, nil); err != nil {
		t.Fatal(err)
	}
	if got, err := links.FindBySource(ctx, "entry", e.ID); err != nil || len(got) != 0 {
		t.Fatalf("clear links: %+v %v", got, err)
	}
	blocks := NewBlockRepository(db)
	if err := blocks.ReplaceForEntry(ctx, nil, e.ID, []BlockRow{{ResearchID: e.ResearchID, BlockID: "one", Type: "paragraph", Data: `{"text":"keep"}`, State: `{}`}}); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 2; i++ {
		if exists, err := blocks.SetState(ctx, nil, e.ID, "one", `{"checked":true}`); err != nil || !exists {
			t.Fatalf("unchanged state must report existing block: %v %v", exists, err)
		}
	}
	if exists, err := blocks.SetState(ctx, nil, e.ID, "missing", `{}`); err != nil || exists {
		t.Fatalf("missing block: %v %v", exists, err)
	}
	if got, err := blocks.FindByEntry(ctx, nil, e.ID); err != nil || len(got) != 1 || got[0].Data != `{"text":"keep"}` || got[0].State != `{"checked":true}` {
		t.Fatalf("state update changed block data: %+v %v", got, err)
	}
}
