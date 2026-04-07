package parser

import (
	"os"
	"path/filepath"
	"testing"
)

func testdataDir() string {
	// Tests run from the package directory; testdata is at repo root.
	return filepath.Join("..", "..", "testdata")
}

func TestParseDirValidADRs(t *testing.T) {
	p := NewMarkdownParser()
	adrs, err := p.ParseDir(testdataDir())
	if err != nil {
		t.Fatalf("ParseDir: %v", err)
	}

	// Should find ADR-001, ADR-002, ADR-099 (bad frontmatter), ADR-050 (empty body).
	// Should NOT find not-an-adr.md.
	if len(adrs) < 2 {
		t.Fatalf("expected at least 2 ADRs, got %d", len(adrs))
	}

	// Verify not-an-adr.md was skipped.
	for _, adr := range adrs {
		if filepath.Base(adr.FilePath) == "not-an-adr.md" {
			t.Error("not-an-adr.md should have been skipped")
		}
	}
}

func TestParseFileMetadata(t *testing.T) {
	p := NewMarkdownParser()
	adrs, err := p.ParseDir(testdataDir())
	if err != nil {
		t.Fatalf("ParseDir: %v", err)
	}

	var adr001 *ADR
	for i, a := range adrs {
		if a.Number == 1 {
			adr001 = &adrs[i]
			break
		}
	}
	if adr001 == nil {
		t.Fatal("ADR-001 not found")
	}

	tests := []struct {
		field string
		got   string
		want  string
	}{
		{"Title", adr001.Title, "Use Go for the Project"},
		{"Status", adr001.Status, "Accepted"},
		{"Date", adr001.Date, "2026-01-15"},
		{"Deciders", adr001.Deciders, "Alice, Bob"},
	}
	for _, tt := range tests {
		if tt.got != tt.want {
			t.Errorf("%s = %q, want %q", tt.field, tt.got, tt.want)
		}
	}
}

func TestChunkADRMultipleSections(t *testing.T) {
	p := NewMarkdownParser()
	adrs, err := p.ParseDir(testdataDir())
	if err != nil {
		t.Fatalf("ParseDir: %v", err)
	}

	var adr001 *ADR
	for i, a := range adrs {
		if a.Number == 1 {
			adr001 = &adrs[i]
			break
		}
	}
	if adr001 == nil {
		t.Fatal("ADR-001 not found")
	}

	chunks := p.ChunkADR(*adr001)
	if len(chunks) < 3 {
		t.Fatalf("expected at least 3 chunks (Context, Decision, Rationale, ...), got %d", len(chunks))
	}

	// Verify chunk section keys.
	keys := make(map[string]bool)
	for _, c := range chunks {
		keys[c.SectionKey] = true
		if c.ADRNumber != 1 {
			t.Errorf("chunk ADRNumber = %d, want 1", c.ADRNumber)
		}
		if c.Content == "" {
			t.Errorf("chunk %q has empty content", c.SectionKey)
		}
	}
	for _, want := range []string{"Context", "Decision", "Rationale"} {
		if !keys[want] {
			t.Errorf("missing expected section %q in chunks", want)
		}
	}
}

func TestChunkADRNoH2(t *testing.T) {
	// Create a temp ADR with no H2 headings.
	dir := t.TempDir()
	content := "# ADR-100: No Sections\n\n**Status:** Accepted\n\nJust a paragraph with no H2 headings.\n"
	if err := os.WriteFile(filepath.Join(dir, "ADR-100-no-sections.md"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	p := NewMarkdownParser()
	adrs, err := p.ParseDir(dir)
	if err != nil {
		t.Fatalf("ParseDir: %v", err)
	}
	if len(adrs) != 1 {
		t.Fatalf("expected 1 ADR, got %d", len(adrs))
	}

	chunks := p.ChunkADR(adrs[0])
	if len(chunks) != 1 {
		t.Fatalf("expected 1 chunk (full body), got %d", len(chunks))
	}
	if chunks[0].SectionKey != "Full" {
		t.Errorf("section key = %q, want %q", chunks[0].SectionKey, "Full")
	}
}

func TestChunkADREmptyBody(t *testing.T) {
	p := NewMarkdownParser()
	adrs, err := p.ParseDir(testdataDir())
	if err != nil {
		t.Fatalf("ParseDir: %v", err)
	}

	var adr050 *ADR
	for i, a := range adrs {
		if a.Number == 50 {
			adr050 = &adrs[i]
			break
		}
	}
	if adr050 == nil {
		t.Fatal("ADR-050 (empty body) not found")
	}

	chunks := p.ChunkADR(*adr050)
	// Empty body or metadata-only should produce zero or one chunk with just metadata.
	// The key point is it doesn't panic.
	_ = chunks
}

func TestParseDirMissingFrontmatter(t *testing.T) {
	p := NewMarkdownParser()
	adrs, err := p.ParseDir(testdataDir())
	if err != nil {
		t.Fatalf("ParseDir: %v", err)
	}

	var adr099 *ADR
	for i, a := range adrs {
		if a.Number == 99 {
			adr099 = &adrs[i]
			break
		}
	}
	if adr099 == nil {
		t.Fatal("ADR-099 (bad frontmatter) not found")
	}

	if adr099.Status != "Unknown" {
		t.Errorf("Status = %q, want %q for missing frontmatter", adr099.Status, "Unknown")
	}
	if adr099.Date != "" {
		t.Errorf("Date = %q, want empty for missing frontmatter", adr099.Date)
	}
}

func TestParseDirEmpty(t *testing.T) {
	dir := t.TempDir()
	p := NewMarkdownParser()
	adrs, err := p.ParseDir(dir)
	if err != nil {
		t.Fatalf("ParseDir on empty dir: %v", err)
	}
	if len(adrs) != 0 {
		t.Errorf("expected 0 ADRs from empty dir, got %d", len(adrs))
	}
}

func TestParseRelatedADRs(t *testing.T) {
	p := NewMarkdownParser()
	adrs, err := p.ParseDir(testdataDir())
	if err != nil {
		t.Fatalf("ParseDir: %v", err)
	}

	tests := []struct {
		adrNum       int
		wantRelCount int
		wantTargets  []int
	}{
		{101, 1, []int{1}},           // ADR-101 relates to ADR-001
		{102, 2, []int{101, 1}},      // ADR-102 relates to ADR-101 and ADR-001
		{103, 1, []int{102}},         // ADR-103 superseded by ADR-102 (from Status field)
		{1, 0, nil},                  // ADR-001 has no Related ADRs section
	}

	adrMap := make(map[int]ADR, len(adrs))
	for _, a := range adrs {
		adrMap[a.Number] = a
	}

	for _, tt := range tests {
		adr, ok := adrMap[tt.adrNum]
		if !ok {
			t.Errorf("ADR-%03d not found in parsed results", tt.adrNum)
			continue
		}
		if len(adr.RelatedADRs) != tt.wantRelCount {
			t.Errorf("ADR-%03d: got %d relationships, want %d", tt.adrNum, len(adr.RelatedADRs), tt.wantRelCount)
			continue
		}
		for i, target := range tt.wantTargets {
			if adr.RelatedADRs[i].TargetADR != target {
				t.Errorf("ADR-%03d rel[%d]: got target %d, want %d", tt.adrNum, i, adr.RelatedADRs[i].TargetADR, target)
			}
		}
	}
}

func TestParseDirNonexistent(t *testing.T) {
	p := NewMarkdownParser()
	_, err := p.ParseDir("/nonexistent/path/that/does/not/exist")
	if err == nil {
		t.Error("expected error for nonexistent directory, got nil")
	}
}
