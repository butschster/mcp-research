package service

import (
	"context"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect"
)

type snapshotInterleaveHook struct {
	ran   atomic.Bool
	write func()
}

func (h *snapshotInterleaveHook) BeforeQuery(ctx context.Context, event *bun.QueryEvent) context.Context {
	if event.Operation() == "SELECT" && strings.Contains(event.Query, "FROM entry_revisions") && h.ran.CompareAndSwap(false, true) {
		h.write()
	}
	return ctx
}

func (*snapshotInterleaveHook) AfterQuery(context.Context, *bun.QueryEvent) {}

func TestLatestSnapshot_ConcurrentCommitCannotSplitBodyAndRevision(t *testing.T) {
	svc, _, ctx, research, section := revisionFixture(t)
	db := svc.revisions.DB()
	if db.Dialect().Name() == dialect.SQLite {
		t.Skip("SQLite serializes reads and writes through its single connection; this contract exercises server MVCC")
	}
	entry := mustEntry(t, svc, ctx, research.ID, section.ID, "Original body")
	var writeErr error
	hook := &snapshotInterleaveHook{write: func() {
		// Commit a real write between LatestSnapshot's entry and history reads.
		_, writeErr = svc.Update(ctx, entry.ID, UpdateEntryRequest{Content: ptr("Concurrent body")})
	}}
	db.AddQueryHook(hook)
	snapshot, revision, err := svc.LatestSnapshot(ctx, entry)
	if err != nil {
		t.Fatal(err)
	}
	if !hook.ran.Load() || writeErr != nil {
		t.Fatalf("interleaved write: ran=%v err=%v", hook.ran.Load(), writeErr)
	}
	if snapshot.Content != entry.Content || revision == nil || revision.Revision != 1 {
		t.Fatalf("snapshot straddled commit: body=%q revision=%+v", snapshot.Content, revision)
	}
	latest, err := svc.revisions.Latest(ctx, nil, entry.ID)
	if err != nil || latest == nil || latest.Revision != 2 {
		t.Fatalf("writer did not commit revision 2: %+v, %v", latest, err)
	}
}
