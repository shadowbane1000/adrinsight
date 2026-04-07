package store

import (
	"context"
	"database/sql"
	"fmt"

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
