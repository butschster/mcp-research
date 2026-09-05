package service

import (
	"context"
	"strings"
	"testing"

	"github.com/dovod-app/app/internal/domain"
	"github.com/dovod-app/app/internal/storage"
)

const artifactHTML = `<!doctype html>
<html><head><meta charset="utf-8">
<title>Throughput by model</title>
<meta name="description" content="Tokens per second for four local models.">
</head><body><div id="chart"></div><script>console.log(1)</script></body></html>`

func TestEntryService_CreateArtifact(t *testing.T) {
	db := setupTestDB(t)
	svc := newEntryService(db, &mockNotifier{})
	ctx := context.Background()

	t.Run("defaults to markdown when type is omitted", func(t *testing.T) {
		r, sec := createTestResearchWithSection(t, db)
		entry, err := svc.Create(ctx, CreateEntryRequest{
			ResearchID: r.ID, SectionID: sec.ID, Content: "# Plain\n\nbody",
		})
		if err != nil {
			t.Fatalf("create: %v", err)
		}
		if entry.Type != domain.EntryMarkdown {
			t.Errorf("Type = %q, want %q", entry.Type, domain.EntryMarkdown)
		}
	})

	t.Run("takes title and description from the document", func(t *testing.T) {
		r, sec := createTestResearchWithSection(t, db)
		entry, err := svc.Create(ctx, CreateEntryRequest{
			ResearchID: r.ID, SectionID: sec.ID,
			Type: domain.EntryArtifact, Content: artifactHTML,
		})
		if err != nil {
			t.Fatalf("create: %v", err)
		}
		// artifact is an input alias: it is stored as a blocks document.
		if entry.Type != domain.EntryBlocks {
			t.Errorf("Type = %q, want %q", entry.Type, domain.EntryBlocks)
		}
		if entry.Title != "Throughput by model" {
			t.Errorf("Title = %q, want it from <title>", entry.Title)
		}
		if entry.Description != "Tokens per second for four local models." {
			t.Errorf("Description = %q, want it from <meta name=description>", entry.Description)
		}
	})

	t.Run("an explicit title wins over the document title", func(t *testing.T) {
		r, sec := createTestResearchWithSection(t, db)
		entry, err := svc.Create(ctx, CreateEntryRequest{
			ResearchID: r.ID, SectionID: sec.ID,
			Type: domain.EntryArtifact, Content: artifactHTML, Title: "Chosen name",
		})
		if err != nil {
			t.Fatalf("create: %v", err)
		}
		if entry.Title != "Chosen name" {
			t.Errorf("Title = %q, want %q", entry.Title, "Chosen name")
		}
	})

	t.Run("rejects an artifact with no title anywhere", func(t *testing.T) {
		r, sec := createTestResearchWithSection(t, db)
		_, err := svc.Create(ctx, CreateEntryRequest{
			ResearchID: r.ID, SectionID: sec.ID,
			Type: domain.EntryArtifact, Content: "<html><body><p>no title</p></body></html>",
		})
		if err == nil {
			t.Fatal("want an error, got nil")
		}
		if !strings.Contains(err.Error(), "title is required") {
			t.Errorf("error = %q, want it to mention the missing title", err)
		}
	})

	t.Run("does not derive a markdown title from HTML", func(t *testing.T) {
		r, sec := createTestResearchWithSection(t, db)
		entry, err := svc.Create(ctx, CreateEntryRequest{
			ResearchID: r.ID, SectionID: sec.ID,
			Type: domain.EntryArtifact, Content: artifactHTML,
		})
		if err != nil {
			t.Fatalf("create: %v", err)
		}
		// The markdown path would have produced the doctype line as a title.
		if strings.Contains(entry.Title, "<") || strings.Contains(entry.Title, "doctype") {
			t.Errorf("Title = %q, want no markup", entry.Title)
		}
	})

	t.Run("rejects an unknown type", func(t *testing.T) {
		r, sec := createTestResearchWithSection(t, db)
		_, err := svc.Create(ctx, CreateEntryRequest{
			ResearchID: r.ID, SectionID: sec.ID,
			Type: domain.EntryType("pdf"), Content: "whatever",
		})
		if err == nil {
			t.Fatal("want an error, got nil")
		}
		if !strings.Contains(err.Error(), "invalid entry_type") {
			t.Errorf("error = %q, want it to name the bad type", err)
		}
	})

	t.Run("survives a round trip through storage", func(t *testing.T) {
		r, sec := createTestResearchWithSection(t, db)
		created, err := svc.Create(ctx, CreateEntryRequest{
			ResearchID: r.ID, SectionID: sec.ID,
			Type: domain.EntryArtifact, Content: artifactHTML,
		})
		if err != nil {
			t.Fatalf("create: %v", err)
		}
		got, err := svc.Get(ctx, created.ID)
		if err != nil {
			t.Fatalf("get: %v", err)
		}
		if got.Type != domain.EntryBlocks {
			t.Errorf("Type after read = %q, want %q", got.Type, domain.EntryBlocks)
		}
		doc, err := NormalizeBlockDocument(got.Content)
		if err != nil {
			t.Fatalf("stored content is not a block document: %v", err)
		}
		if len(doc.Blocks) != 1 || doc.Blocks[0].Type != domain.BlockHTML {
			t.Fatalf("stored blocks = %v, want a single html block", doc.Blocks)
		}
		if str(doc.Blocks[0].Data, "html") != artifactHTML {
			t.Error("the HTML body was altered on the way through storage")
		}
	})

	t.Run("listing reports the type without loading content", func(t *testing.T) {
		r, sec := createTestResearchWithSection(t, db)
		if _, err := svc.Create(ctx, CreateEntryRequest{
			ResearchID: r.ID, SectionID: sec.ID,
			Type: domain.EntryArtifact, Content: artifactHTML,
		}); err != nil {
			t.Fatalf("create: %v", err)
		}
		list, err := svc.List(ctx, r.ID, sec.ID, storage.EntryFilter{})
		if err != nil {
			t.Fatalf("list: %v", err)
		}
		if len(list) != 1 {
			t.Fatalf("got %d entries, want 1", len(list))
		}
		if list[0].Type != domain.EntryBlocks {
			t.Errorf("Type in list = %q, want %q", list[0].Type, domain.EntryBlocks)
		}
	})
}

func TestEntryService_UpdateArtifactType(t *testing.T) {
	db := setupTestDB(t)
	svc := newEntryService(db, &mockNotifier{})
	ctx := context.Background()

	t.Run("switches a markdown entry to an artifact", func(t *testing.T) {
		r, sec := createTestResearchWithSection(t, db)
		entry, err := svc.Create(ctx, CreateEntryRequest{
			ResearchID: r.ID, SectionID: sec.ID, Content: "# Was markdown",
		})
		if err != nil {
			t.Fatalf("create: %v", err)
		}

		artifact := domain.EntryArtifact
		html := artifactHTML
		updated, err := svc.Update(ctx, entry.ID, UpdateEntryRequest{
			Type: &artifact, Content: &html,
		})
		if err != nil {
			t.Fatalf("update: %v", err)
		}
		if updated.Type != domain.EntryBlocks {
			t.Errorf("Type = %q, want %q", updated.Type, domain.EntryBlocks)
		}
		doc, err := NormalizeBlockDocument(updated.Content)
		if err != nil {
			t.Fatalf("stored content is not a block document: %v", err)
		}
		if str(doc.Blocks[0].Data, "html") != artifactHTML {
			t.Error("content was not replaced")
		}
	})

	t.Run("rejects an unknown type on update", func(t *testing.T) {
		r, sec := createTestResearchWithSection(t, db)
		entry, err := svc.Create(ctx, CreateEntryRequest{
			ResearchID: r.ID, SectionID: sec.ID, Content: "# Markdown",
		})
		if err != nil {
			t.Fatalf("create: %v", err)
		}
		bad := domain.EntryType("widget")
		if _, err := svc.Update(ctx, entry.ID, UpdateEntryRequest{Type: &bad}); err == nil {
			t.Fatal("want an error, got nil")
		}
	})

	t.Run("leaves the type alone when it is not given", func(t *testing.T) {
		r, sec := createTestResearchWithSection(t, db)
		entry, err := svc.Create(ctx, CreateEntryRequest{
			ResearchID: r.ID, SectionID: sec.ID,
			Type: domain.EntryArtifact, Content: artifactHTML,
		})
		if err != nil {
			t.Fatalf("create: %v", err)
		}
		title := "Renamed"
		updated, err := svc.Update(ctx, entry.ID, UpdateEntryRequest{Title: &title})
		if err != nil {
			t.Fatalf("update: %v", err)
		}
		if updated.Type != domain.EntryBlocks {
			t.Errorf("Type = %q, want it unchanged as %q", updated.Type, domain.EntryBlocks)
		}
	})
}

func TestHTMLMetadataExtraction(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		wantT    string
		wantDesc string
	}{
		{
			name:     "plain title and description",
			html:     `<head><title>Hello</title><meta name="description" content="A summary."></head>`,
			wantT:    "Hello",
			wantDesc: "A summary.",
		},
		{
			name:  "uppercase tags and attributes",
			html:  `<HEAD><TITLE>Shouty</TITLE><META NAME="description" CONTENT="Loud."></HEAD>`,
			wantT: "Shouty", wantDesc: "Loud.",
		},
		{
			name:  "title spanning lines is collapsed by the caller",
			html:  "<title>Two\nlines</title>",
			wantT: "Two\nlines",
		},
		{
			name:  "single quotes and attribute order",
			html:  `<meta content='Reversed order.' name='description'>`,
			wantT: "", wantDesc: "Reversed order.",
		},
		{
			name:  "no metadata at all",
			html:  `<html><body>nothing</body></html>`,
			wantT: "", wantDesc: "",
		},
		{
			name:  "title with attributes on the tag",
			html:  `<title data-x="1">With attrs</title>`,
			wantT: "With attrs",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := htmlTitle(tt.html); got != tt.wantT {
				t.Errorf("htmlTitle = %q, want %q", got, tt.wantT)
			}
			if got := htmlMetaDescription(tt.html); got != tt.wantDesc {
				t.Errorf("htmlMetaDescription = %q, want %q", got, tt.wantDesc)
			}
		})
	}
}
