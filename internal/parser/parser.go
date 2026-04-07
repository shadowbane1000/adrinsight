package parser

// ADR represents a parsed Architecture Decision Record.
type ADR struct {
	FilePath    string
	Number      int
	Title       string
	Date        string
	Status      string
	Deciders    string
	Body        string
	RelatedADRs []RawRelationship
}

// RawRelationship is a parsed but unclassified relationship from a Related ADRs section.
type RawRelationship struct {
	TargetADR   int
	Description string
}

// Chunk represents a section of an ADR suitable for embedding.
type Chunk struct {
	ADRNumber  int
	SectionKey string
	Content    string
}

// Parser parses ADR markdown files from a directory.
type Parser interface {
	ParseDir(dir string) ([]ADR, error)
	ChunkADR(adr ADR) []Chunk
}
