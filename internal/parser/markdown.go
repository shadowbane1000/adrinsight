package parser

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

var adrFilePattern = regexp.MustCompile(`(?i)^ADR-(\d+).*\.md$`)

// MarkdownParser parses ADR files using goldmark.
type MarkdownParser struct{}

// NewMarkdownParser creates a new MarkdownParser.
func NewMarkdownParser() *MarkdownParser {
	return &MarkdownParser{}
}

// ParseDir reads all ADR-*.md files from dir and returns parsed ADRs.
func (p *MarkdownParser) ParseDir(dir string) ([]ADR, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("reading ADR directory: %w", err)
	}

	var adrs []ADR
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		matches := adrFilePattern.FindStringSubmatch(e.Name())
		if matches == nil {
			continue
		}
		num, _ := strconv.Atoi(matches[1])

		fp := filepath.Join(dir, e.Name())
		adr, err := p.parseFile(fp, num)
		if err != nil {
			return nil, fmt.Errorf("parsing %s: %w", e.Name(), err)
		}
		adrs = append(adrs, adr)
	}
	return adrs, nil
}

func (p *MarkdownParser) parseFile(path string, number int) (ADR, error) {
	src, err := os.ReadFile(path)
	if err != nil {
		return ADR{}, err
	}

	md := goldmark.New()
	doc := md.Parser().Parse(text.NewReader(src))

	adr := ADR{
		FilePath: path,
		Number:   number,
		Status:   "Unknown",
	}

	// Walk the AST to extract title and body structure.
	var bodyStart int
	_ = ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		if h, ok := n.(*ast.Heading); ok && h.Level == 1 && adr.Title == "" {
			adr.Title = headingText(h, src)
			// Body starts after the H1 line.
			if h.NextSibling() != nil {
				bodyStart = h.NextSibling().Lines().At(0).Start
			}
		}
		return ast.WalkContinue, nil
	})

	if adr.Title == "" {
		adr.Title = filepath.Base(path)
	}

	// Strip ADR number prefix from title if present (e.g., "ADR-001: Why Go" → "Why Go").
	if idx := strings.Index(adr.Title, ": "); idx != -1 {
		adr.Title = adr.Title[idx+2:]
	}

	// Extract frontmatter-style metadata from the body text.
	body := string(src[bodyStart:])
	adr.Body = body
	p.extractMetadata(&adr, body)

	return adr, nil
}

func (p *MarkdownParser) extractMetadata(adr *ADR, body string) {
	for _, line := range strings.Split(body, "\n") {
		line = strings.TrimSpace(line)
		// Strip bold markdown markers: **Status:** Accepted → Status: Accepted
		line = strings.ReplaceAll(line, "**", "")
		line = strings.TrimSpace(line)

		if kv := extractKV(line, "Status:"); kv != "" {
			adr.Status = kv
		} else if kv := extractKV(line, "Date:"); kv != "" {
			adr.Date = kv
		} else if kv := extractKV(line, "Deciders:"); kv != "" {
			adr.Deciders = kv
		}
	}
}

func extractKV(line, key string) string {
	if strings.HasPrefix(line, key) {
		return strings.TrimSpace(strings.TrimPrefix(line, key))
	}
	return ""
}

// ChunkADR splits an ADR into chunks by H2 headings.
func (p *MarkdownParser) ChunkADR(adr ADR) []Chunk {
	src := []byte(adr.Body)
	md := goldmark.New()
	doc := md.Parser().Parse(text.NewReader(src))

	type section struct {
		key   string
		start int
		end   int
	}

	var sections []section
	_ = ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		if h, ok := n.(*ast.Heading); ok && h.Level == 2 {
			key := headingText(h, src)
			start := 0
			if h.Lines().Len() > 0 {
				start = h.Lines().At(0).Start
			}
			// Close the previous section.
			if len(sections) > 0 {
				sections[len(sections)-1].end = start
			}
			sections = append(sections, section{key: key, start: start, end: len(src)})
		}
		return ast.WalkContinue, nil
	})

	if len(sections) == 0 {
		// No H2 headings — entire body is one chunk.
		content := strings.TrimSpace(adr.Body)
		if content == "" {
			return nil
		}
		return []Chunk{{
			ADRNumber:  adr.Number,
			SectionKey: "Full",
			Content:    content,
		}}
	}

	// Close the last section.
	if len(sections) > 0 {
		sections[len(sections)-1].end = len(src)
	}

	var chunks []Chunk
	for _, s := range sections {
		content := strings.TrimSpace(string(src[s.start:s.end]))
		// Strip the heading line itself from content.
		if idx := strings.Index(content, "\n"); idx != -1 {
			content = strings.TrimSpace(content[idx+1:])
		}
		if content == "" {
			continue
		}
		chunks = append(chunks, Chunk{
			ADRNumber:  adr.Number,
			SectionKey: s.key,
			Content:    content,
		})
	}
	return chunks
}

// headingText extracts the text content of a heading node without using
// the deprecated Node.Text() method.
func headingText(h *ast.Heading, src []byte) string {
	var b strings.Builder
	for c := h.FirstChild(); c != nil; c = c.NextSibling() {
		if t, ok := c.(*ast.Text); ok {
			b.Write(t.Value(src))
		}
	}
	return b.String()
}

// Verify MarkdownParser implements Parser.
var _ Parser = (*MarkdownParser)(nil)

