package papl

import (
	"context"
	"fmt"
	"strconv"
	"time"
)

// SeedTask fetches the TOC page and enqueues all chapter URLs.
type SeedTask struct {
	Config  Config
	DB      *DB
	StateDB *State
	Client  *Client
}

// Run executes the seed task.
func (t *SeedTask) Run(ctx context.Context, emit func(*SeedState)) (SeedMetric, error) {
	start := time.Now()
	state := &SeedState{}
	emit(state)

	baseURL := t.Config.BaseURL
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}

	body, err := t.Client.FetchPage(ctx, baseURL)
	if err != nil {
		return SeedMetric{}, fmt.Errorf("fetch TOC: %w", err)
	}
	state.Pages++
	emit(state)

	toc, err := ParseTOC(baseURL, body)
	if err != nil {
		return SeedMetric{}, fmt.Errorf("parse TOC: %w", err)
	}

	// Derive edition ID from URL.
	editionID := EditionFromURL(baseURL)
	year, _ := strconv.Atoi(editionID)

	// Insert/update edition row.
	edition := Edition{
		ID:           editionID,
		Year:         year,
		BaseURL:      baseURL,
		ChapterCount: len(toc.Chapters),
		FetchedAt:    time.Now(),
	}
	_ = t.DB.UpsertEdition(edition)

	for _, ch := range toc.Chapters {
		if ctx.Err() != nil {
			break
		}
		if !t.StateDB.IsVisited(ch.URL) {
			_ = t.StateDB.Enqueue(ch.URL, EntityChapter, PriorityChapter, ch.Slug)
			state.Enqueued++
		}
		state.Chapters++
		emit(state)
	}

	return SeedMetric{
		Pages:    state.Pages,
		Chapters: state.Chapters,
		Enqueued: state.Enqueued,
		Duration: time.Since(start),
	}, nil
}
