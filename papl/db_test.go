package papl_test

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/tamnd/papl-cli/papl"
)

func TestDBRoundtrip(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	db, err := papl.OpenDB(dbPath)
	if err != nil {
		t.Fatalf("OpenDB: %v", err)
	}
	defer db.Close()

	edition := papl.Edition{
		ID:           "2020",
		Year:         2020,
		BaseURL:      "https://papl.cs.brown.edu/2020/",
		ChapterCount: 1,
		FetchedAt:    time.Now(),
	}
	if err := db.UpsertEdition(edition); err != nil {
		t.Fatalf("UpsertEdition: %v", err)
	}

	chapter := papl.Chapter{
		ID:             "2020/getting-started",
		EditionID:      "2020",
		Slug:           "getting-started",
		Number:         1,
		Title:          "Getting Started",
		URL:            "https://papl.cs.brown.edu/2020/getting-started.html",
		Description:    "First steps with Pyret",
		Topics:         "[]",
		ExerciseCount:  3,
		CodeBlockCount: 5,
		HasDoNow:       true,
		DoNowCount:     2,
		BodyMD:         "# Getting Started\n\nSome content here.",
		FetchedAt:      time.Now(),
	}
	if err := db.UpsertChapter(chapter); err != nil {
		t.Fatalf("UpsertChapter: %v", err)
	}

	chapters, err := db.ListChapters()
	if err != nil {
		t.Fatalf("ListChapters: %v", err)
	}
	if len(chapters) != 1 {
		t.Fatalf("ListChapters returned %d chapters, want 1", len(chapters))
	}
	if chapters[0].Title != "Getting Started" {
		t.Errorf("Title = %q, want %q", chapters[0].Title, "Getting Started")
	}
	if !chapters[0].HasDoNow {
		t.Errorf("HasDoNow = false, want true")
	}

	stats, err := db.Stats()
	if err != nil {
		t.Fatalf("Stats: %v", err)
	}
	if stats.Editions != 1 {
		t.Errorf("Stats.Editions = %d, want 1", stats.Editions)
	}
	if stats.Chapters != 1 {
		t.Errorf("Stats.Chapters = %d, want 1", stats.Chapters)
	}
}
