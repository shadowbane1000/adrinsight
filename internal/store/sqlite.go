package store

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	sqlite_vec "github.com/asg017/sqlite-vec-go-bindings/cgo"
	_ "github.com/mattn/go-sqlite3"
)

func init() {
	sqlite_vec.Auto()
}

// SQLiteStore implements Store using SQLite with sqlite-vec.
type SQLiteStore struct {
	db *sql.DB
}

// NewSQLiteStore opens or creates a SQLite database at dbPath.
func NewSQLiteStore(dbPath string) (*SQLiteStore, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}
	return &SQLiteStore{db: db}, nil
}

// Reset drops and recreates the chunks and vec_chunks tables.
func (s *SQLiteStore) Reset(ctx context.Context) error {
	stmts := []string{
		"DROP TABLE IF EXISTS chunks",
		"DROP TABLE IF EXISTS vec_chunks",
		"DROP TABLE IF EXISTS fts_chunks",
		`CREATE TABLE chunks (
			id         INTEGER PRIMARY KEY AUTOINCREMENT,
			adr_number INTEGER NOT NULL,
			adr_title  TEXT    NOT NULL,
			adr_status TEXT    NOT NULL DEFAULT 'Unknown',
			adr_path   TEXT    NOT NULL,
			section    TEXT    NOT NULL,
			content    TEXT    NOT NULL
		)`,
		"CREATE VIRTUAL TABLE vec_chunks USING vec0(embedding float[768])",
		"CREATE VIRTUAL TABLE fts_chunks USING fts5(content)",
	}
	for _, stmt := range stmts {
		if _, err := s.db.ExecContext(ctx, stmt); err != nil {
			return fmt.Errorf("executing %q: %w", truncate(stmt, 60), err)
		}
	}
	return nil
}

// StoreChunks inserts chunk metadata and embeddings.
func (s *SQLiteStore) StoreChunks(ctx context.Context, chunks []ChunkRecord) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	metaStmt, err := tx.PrepareContext(ctx,
		`INSERT INTO chunks (adr_number, adr_title, adr_status, adr_path, section, content)
		 VALUES (?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return fmt.Errorf("prepare meta insert: %w", err)
	}
	defer func() { _ = metaStmt.Close() }()

	vecStmt, err := tx.PrepareContext(ctx,
		"INSERT INTO vec_chunks (rowid, embedding) VALUES (?, ?)")
	if err != nil {
		return fmt.Errorf("prepare vec insert: %w", err)
	}
	defer func() { _ = vecStmt.Close() }()

	ftsStmt, err := tx.PrepareContext(ctx,
		"INSERT INTO fts_chunks (rowid, content) VALUES (?, ?)")
	if err != nil {
		return fmt.Errorf("prepare fts insert: %w", err)
	}
	defer func() { _ = ftsStmt.Close() }()

	for _, c := range chunks {
		res, err := metaStmt.ExecContext(ctx,
			c.ADRNumber, c.ADRTitle, c.ADRStatus, c.ADRPath, c.Section, c.Content)
		if err != nil {
			return fmt.Errorf("insert chunk meta: %w", err)
		}
		rowID, err := res.LastInsertId()
		if err != nil {
			return fmt.Errorf("last insert id: %w", err)
		}

		blob, err := sqlite_vec.SerializeFloat32(c.Embedding)
		if err != nil {
			return fmt.Errorf("serialize embedding: %w", err)
		}
		if _, err := vecStmt.ExecContext(ctx, rowID, blob); err != nil {
			return fmt.Errorf("insert vec: %w", err)
		}
		if _, err := ftsStmt.ExecContext(ctx, rowID, c.Content); err != nil {
			return fmt.Errorf("insert fts: %w", err)
		}
	}
	return tx.Commit()
}

// Search finds the topK most similar chunks to the query vector.
func (s *SQLiteStore) Search(ctx context.Context, query []float32, topK int) ([]SearchResult, error) {
	blob, err := sqlite_vec.SerializeFloat32(query)
	if err != nil {
		return nil, fmt.Errorf("serialize query: %w", err)
	}

	rows, err := s.db.QueryContext(ctx,
		`SELECT rowid, distance
		 FROM vec_chunks
		 WHERE embedding MATCH ?
		 ORDER BY distance
		 LIMIT ?`, blob, topK)
	if err != nil {
		return nil, fmt.Errorf("search query: %w", err)
	}
	defer func() { _ = rows.Close() }()

	type vecResult struct {
		rowID    int64
		distance float64
	}
	var vecs []vecResult
	for rows.Next() {
		var vr vecResult
		if err := rows.Scan(&vr.rowID, &vr.distance); err != nil {
			return nil, fmt.Errorf("scan vec result: %w", err)
		}
		vecs = append(vecs, vr)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("vec rows: %w", err)
	}

	var results []SearchResult
	for _, vr := range vecs {
		var r SearchResult
		err := s.db.QueryRowContext(ctx,
			`SELECT adr_number, adr_title, adr_path, section, content
			 FROM chunks WHERE id = ?`, vr.rowID).Scan(
			&r.ADRNumber, &r.ADRTitle, &r.ADRPath, &r.Section, &r.Content)
		if err != nil {
			return nil, fmt.Errorf("lookup chunk %d: %w", vr.rowID, err)
		}
		r.Score = vr.distance
		results = append(results, r)
	}
	return results, nil
}

// ListADRs returns distinct ADR metadata from the chunks table.
func (s *SQLiteStore) ListADRs(ctx context.Context) ([]ADRSummary, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT DISTINCT adr_number, adr_title, adr_status, adr_path
		 FROM chunks ORDER BY adr_number`)
	if err != nil {
		return nil, fmt.Errorf("list ADRs: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var adrs []ADRSummary
	for rows.Next() {
		var a ADRSummary
		if err := rows.Scan(&a.Number, &a.Title, &a.Status, &a.Path); err != nil {
			return nil, fmt.Errorf("scan ADR: %w", err)
		}
		adrs = append(adrs, a)
	}
	return adrs, rows.Err()
}

// IsEmpty checks whether the chunks table has any rows.
func (s *SQLiteStore) IsEmpty(ctx context.Context) (bool, error) {
	var count int
	err := s.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM chunks").Scan(&count)
	if err != nil {
		// Table may not exist yet — treat as empty.
		return true, nil
	}
	return count == 0, nil
}

// prepareFTSQuery converts a natural language query into an FTS5 OR query
// by stripping stop words and joining remaining terms with OR.
func prepareFTSQuery(query string) string {
	stopWords := map[string]bool{
		"a": true, "an": true, "and": true, "are": true, "as": true, "at": true,
		"be": true, "been": true, "but": true, "by": true, "can": true, "case": true,
		"could": true, "do": true, "does": true, "for": true, "from": true,
		"had": true, "has": true, "have": true, "how": true, "i": true,
		"if": true, "in": true, "into": true, "is": true, "it": true, "its": true,
		"like": true, "made": true, "may": true, "must": true,
		"need": true, "needed": true, "needs": true, "no": true, "not": true,
		"of": true, "on": true, "or": true, "our": true, "over": true,
		"should": true, "so": true, "some": true, "such": true,
		"than": true, "that": true, "the": true, "their": true,
		"them": true, "then": true, "there": true, "these": true, "they": true,
		"this": true, "to": true, "up": true, "us": true,
		"was": true, "we": true, "were": true, "what": true,
		"when": true, "where": true, "which": true, "who": true, "why": true,
		"will": true, "with": true, "would": true, "you": true, "your": true,
	}

	words := strings.Fields(strings.ToLower(query))
	var terms []string
	for _, w := range words {
		// Strip non-alphanumeric characters.
		cleaned := strings.Map(func(r rune) rune {
			if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
				return r
			}
			return -1
		}, w)
		if cleaned == "" || stopWords[cleaned] {
			continue
		}
		terms = append(terms, cleaned)
	}

	if len(terms) == 0 {
		return query // fall back to original
	}
	return strings.Join(terms, " OR ")
}

// SearchFTS performs full-text keyword search using FTS5 BM25 ranking.
func (s *SQLiteStore) SearchFTS(ctx context.Context, query string, topK int) ([]SearchResult, error) {
	ftsQuery := prepareFTSQuery(query)
	rows, err := s.db.QueryContext(ctx,
		`SELECT fts.rowid, fts.rank
		 FROM fts_chunks fts
		 WHERE fts_chunks MATCH ?
		 ORDER BY rank
		 LIMIT ?`, ftsQuery, topK)
	if err != nil {
		return nil, fmt.Errorf("fts search: %w", err)
	}
	defer func() { _ = rows.Close() }()

	type ftsHit struct {
		rowID int64
		rank  float64
	}
	var hits []ftsHit
	for rows.Next() {
		var h ftsHit
		if err := rows.Scan(&h.rowID, &h.rank); err != nil {
			return nil, fmt.Errorf("scan fts result: %w", err)
		}
		hits = append(hits, h)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("fts rows: %w", err)
	}

	if len(hits) == 0 {
		return nil, nil
	}

	// Normalize BM25 scores to 0-1 (rank is negative; closer to 0 = better).
	maxAbs := 0.0
	for _, h := range hits {
		a := -h.rank
		if a > maxAbs {
			maxAbs = a
		}
	}

	var results []SearchResult
	for _, h := range hits {
		var r SearchResult
		err := s.db.QueryRowContext(ctx,
			`SELECT adr_number, adr_title, adr_path, section, content
			 FROM chunks WHERE id = ?`, h.rowID).Scan(
			&r.ADRNumber, &r.ADRTitle, &r.ADRPath, &r.Section, &r.Content)
		if err != nil {
			return nil, fmt.Errorf("lookup chunk %d: %w", h.rowID, err)
		}
		if maxAbs > 0 {
			r.Score = 1.0 - (-h.rank / maxAbs)
		} else {
			r.Score = 1.0
		}
		results = append(results, r)
	}
	return results, nil
}

// HybridSearch combines vector and FTS5 keyword search with weighted merge.
func (s *SQLiteStore) HybridSearch(ctx context.Context, queryVec []float32, queryText string, topK int, vecWeight, kwWeight float64) ([]SearchResult, error) {
	// Run vector search.
	vecResults, err := s.Search(ctx, queryVec, topK*2)
	if err != nil {
		return nil, fmt.Errorf("vector search: %w", err)
	}

	// Normalize vector distances to 0-1 scores (lower distance = higher score).
	if len(vecResults) > 0 {
		maxDist := 0.0
		for _, r := range vecResults {
			if r.Score > maxDist {
				maxDist = r.Score
			}
		}
		if maxDist > 0 {
			for i := range vecResults {
				vecResults[i].Score = 1.0 - (vecResults[i].Score / maxDist)
			}
		}
	}

	// Run FTS search.
	ftsResults, err := s.SearchFTS(ctx, queryText, topK*2)
	if err != nil {
		// FTS failure is non-fatal — fall back to vector only.
		ftsResults = nil
	}

	// If no FTS results, return vector-only (keyword weight becomes 0).
	if len(ftsResults) == 0 {
		// Deduplicate by ADR number.
		return deduplicateByADR(vecResults, topK), nil
	}

	// Weighted merge: build a map keyed by chunk rowid (approximated by ADR+section).
	type mergeKey struct {
		adrNumber int
		section   string
	}
	merged := make(map[mergeKey]*SearchResult)

	for i := range vecResults {
		k := mergeKey{vecResults[i].ADRNumber, vecResults[i].Section}
		r := vecResults[i]
		r.Score = r.Score * vecWeight
		merged[k] = &r
	}

	for i := range ftsResults {
		k := mergeKey{ftsResults[i].ADRNumber, ftsResults[i].Section}
		if existing, ok := merged[k]; ok {
			existing.Score += ftsResults[i].Score * kwWeight
		} else {
			r := ftsResults[i]
			r.Score = r.Score * kwWeight
			merged[k] = &r
		}
	}

	// Collect and sort by combined score descending.
	var combined []SearchResult
	for _, r := range merged {
		combined = append(combined, *r)
	}
	sortByScoreDesc(combined)

	return deduplicateByADR(combined, topK), nil
}

func deduplicateByADR(results []SearchResult, topK int) []SearchResult {
	seen := make(map[int]bool)
	var deduped []SearchResult
	for _, r := range results {
		if seen[r.ADRNumber] {
			continue
		}
		seen[r.ADRNumber] = true
		deduped = append(deduped, r)
		if len(deduped) >= topK {
			break
		}
	}
	return deduped
}

func sortByScoreDesc(results []SearchResult) {
	for i := 1; i < len(results); i++ {
		for j := i; j > 0 && results[j].Score > results[j-1].Score; j-- {
			results[j], results[j-1] = results[j-1], results[j]
		}
	}
}

// Close closes the database connection.
func (s *SQLiteStore) Close() error {
	return s.db.Close()
}

func truncate(s string, n int) string {
	if len(s) > n {
		return s[:n]
	}
	return s
}

// Verify SQLiteStore implements Store.
var _ Store = (*SQLiteStore)(nil)
