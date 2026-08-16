package papl

import "time"

const (
	// EntityChapter is the entity type for chapter queue items.
	EntityChapter = "chapter"

	// PriorityChapter is the queue priority for chapters.
	PriorityChapter = 50
)

// Edition represents one textbook edition.
type Edition struct {
	ID           string
	Year         int
	BaseURL      string
	ChapterCount int
	FetchedAt    time.Time
}

// Chapter is one row in the chapters table.
type Chapter struct {
	ID             string
	EditionID      string
	Slug           string
	Number         int
	Title          string
	URL            string
	Part           string
	Description    string
	Topics         string // JSON array
	ExerciseCount  int
	CodeBlockCount int
	HasDoNow       bool
	DoNowCount     int
	BodyMD         string
	FetchedAt      time.Time
}

// Exercise is one row in the exercises table.
type Exercise struct {
	ID        string
	ChapterID string
	Number    int
	Body      string
	HasCode   bool
	Concepts  string // JSON array
	FetchedAt time.Time
}

// QueueItem is one item in the crawl queue.
type QueueItem struct {
	URL        string
	EntityType string
	Priority   int
	Meta       string // slug
}

// DBStats summarises the DB for the info command.
type DBStats struct {
	Editions  int64
	Chapters  int64
	Exercises int64
	DBSize    int64
}

// SeedState is the live progress struct emitted during seed.
type SeedState struct {
	Pages    int
	Chapters int
	Enqueued int
}

// SeedMetric is returned by SeedTask.Run.
type SeedMetric struct {
	Pages    int
	Chapters int
	Enqueued int
	Duration time.Duration
}

// CrawlState is the live progress struct emitted during crawl.
type CrawlState struct {
	Done    int64
	Pending int64
	Failed  int64
	RPS     float64
}

// CrawlMetric is returned by CrawlTask.Run.
type CrawlMetric struct {
	Done     int64
	Failed   int64
	Duration time.Duration
}

// ExportState is the live progress struct emitted during export.
type ExportState struct {
	Written int
	Current string
}

// ExportMetric is returned by ExportTask.Run.
type ExportMetric struct {
	Files    int
	Duration time.Duration
}
