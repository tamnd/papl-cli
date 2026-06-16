package papl

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

// DB wraps a SQLite database holding PAPL chapter data.
type DB struct {
	db   *sql.DB
	path string
}

// OpenDB opens (or creates) the database at path.
func OpenDB(path string) (*DB, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("create db dir: %w", err)
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite %s: %w", path, err)
	}
	db.SetMaxOpenConns(1)
	d := &DB{db: db, path: path}
	if err := d.initSchema(); err != nil {
		db.Close()
		return nil, err
	}
	return d, nil
}

// Close closes the database.
func (d *DB) Close() error { return d.db.Close() }

func (d *DB) initSchema() error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS editions (
			id            TEXT PRIMARY KEY,
			year          INTEGER DEFAULT 0,
			base_url      TEXT DEFAULT '',
			chapter_count INTEGER DEFAULT 0,
			fetched_at    DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS chapters (
			id               TEXT PRIMARY KEY,
			edition_id       TEXT DEFAULT '',
			slug             TEXT DEFAULT '',
			number           INTEGER DEFAULT 0,
			title            TEXT DEFAULT '',
			url              TEXT DEFAULT '',
			part             TEXT DEFAULT '',
			description      TEXT DEFAULT '',
			topics           TEXT DEFAULT '[]',
			exercise_count   INTEGER DEFAULT 0,
			code_block_count INTEGER DEFAULT 0,
			has_do_now       INTEGER DEFAULT 0,
			do_now_count     INTEGER DEFAULT 0,
			body_md          TEXT DEFAULT '',
			fetched_at       DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS exercises (
			id         TEXT PRIMARY KEY,
			chapter_id TEXT DEFAULT '',
			number     INTEGER DEFAULT 0,
			body       TEXT DEFAULT '',
			has_code   INTEGER DEFAULT 0,
			concepts   TEXT DEFAULT '[]',
			fetched_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
	}
	for _, stmt := range stmts {
		if _, err := d.db.Exec(stmt); err != nil {
			return fmt.Errorf("init schema: %w", err)
		}
	}
	return nil
}

// UpsertEdition inserts or updates an edition record.
func (d *DB) UpsertEdition(e Edition) error {
	_, err := d.db.Exec(`INSERT INTO editions (id, year, base_url, chapter_count, fetched_at)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT (id) DO UPDATE SET
			year=excluded.year,
			base_url=CASE WHEN excluded.base_url != '' THEN excluded.base_url ELSE editions.base_url END,
			chapter_count=excluded.chapter_count,
			fetched_at=excluded.fetched_at`,
		e.ID, e.Year, e.BaseURL, e.ChapterCount,
		e.FetchedAt.UTC().Format(time.RFC3339))
	return err
}

// UpsertChapter inserts or updates a chapter record.
func (d *DB) UpsertChapter(c Chapter) error {
	_, err := d.db.Exec(`INSERT INTO chapters
		(id, edition_id, slug, number, title, url, part, description, topics,
		 exercise_count, code_block_count, has_do_now, do_now_count, body_md, fetched_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT (id) DO UPDATE SET
			edition_id=excluded.edition_id,
			slug=CASE WHEN excluded.slug != '' THEN excluded.slug ELSE chapters.slug END,
			number=excluded.number,
			title=CASE WHEN excluded.title != '' THEN excluded.title ELSE chapters.title END,
			url=CASE WHEN excluded.url != '' THEN excluded.url ELSE chapters.url END,
			part=CASE WHEN excluded.part != '' THEN excluded.part ELSE chapters.part END,
			description=CASE WHEN excluded.description != '' THEN excluded.description ELSE chapters.description END,
			topics=CASE WHEN excluded.topics != '[]' THEN excluded.topics ELSE chapters.topics END,
			exercise_count=excluded.exercise_count,
			code_block_count=excluded.code_block_count,
			has_do_now=excluded.has_do_now,
			do_now_count=excluded.do_now_count,
			body_md=CASE WHEN excluded.body_md != '' THEN excluded.body_md ELSE chapters.body_md END,
			fetched_at=excluded.fetched_at`,
		c.ID, c.EditionID, c.Slug, c.Number, c.Title, c.URL, c.Part,
		c.Description, c.Topics, c.ExerciseCount, c.CodeBlockCount,
		boolToInt(c.HasDoNow), c.DoNowCount, c.BodyMD,
		c.FetchedAt.UTC().Format(time.RFC3339))
	return err
}

// ListChapters returns all chapters ordered by number, id.
func (d *DB) ListChapters() ([]Chapter, error) {
	rows, err := d.db.Query(`SELECT id, COALESCE(edition_id,''), COALESCE(slug,''), number,
		COALESCE(title,''), COALESCE(url,''), COALESCE(part,''), COALESCE(description,''),
		COALESCE(topics,'[]'), COALESCE(exercise_count,0), COALESCE(code_block_count,0),
		COALESCE(has_do_now,0), COALESCE(do_now_count,0), COALESCE(body_md,''), fetched_at
		FROM chapters ORDER BY number, id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Chapter
	for rows.Next() {
		var c Chapter
		var hasDN int
		var fetchedAt string
		if err := rows.Scan(&c.ID, &c.EditionID, &c.Slug, &c.Number,
			&c.Title, &c.URL, &c.Part, &c.Description,
			&c.Topics, &c.ExerciseCount, &c.CodeBlockCount,
			&hasDN, &c.DoNowCount, &c.BodyMD, &fetchedAt); err != nil {
			return nil, err
		}
		c.HasDoNow = hasDN != 0
		c.FetchedAt, _ = time.Parse(time.RFC3339, fetchedAt)
		out = append(out, c)
	}
	return out, rows.Err()
}

// Stats returns aggregate statistics about the database.
func (d *DB) Stats() (DBStats, error) {
	var s DBStats
	_ = d.db.QueryRow(`SELECT COUNT(*) FROM editions`).Scan(&s.Editions)
	_ = d.db.QueryRow(`SELECT COUNT(*) FROM chapters`).Scan(&s.Chapters)
	_ = d.db.QueryRow(`SELECT COUNT(*) FROM exercises`).Scan(&s.Exercises)
	if fi, err := os.Stat(d.path); err == nil {
		s.DBSize = fi.Size()
	}
	return s, nil
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
