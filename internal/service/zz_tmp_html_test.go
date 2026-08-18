package service

import (
	"strings"
	"testing"

	"github.com/butschster/mcp-research/internal/domain"
)

func TestTmpMarkdownExportHTMLBlock(t *testing.T) {
	k := newMetaKit(t)
	entry, err := k.entry.Create(k.ctx, CreateEntryRequest{
		ResearchID: k.research.ID, SectionID: k.section.ID,
		Type:  domain.EntryBlocks,
		Title: "Chart",
		Content: `{"version":1,"blocks":[{"type":"html","data":{"title":"Revenue chart","caption":"Q3","html":"<div id=x>hi</div><script>alert(1)</script>"}}]}`,
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	file, err := k.entry.MarkdownExport(k.ctx, entry.ID)
	if err != nil {
		t.Fatalf("export: %v", err)
	}
	t.Logf("FILENAME %q", file.Filename)
	t.Logf("CONTENT:\n%s", file.Content)
	if strings.Contains(file.Content, "<script>") {
		t.Logf("=> INLINED")
	} else {
		t.Logf("=> NAMED")
	}
}
