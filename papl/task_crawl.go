package papl

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

const (
	crawlFlushEvery    = 50
	crawlFlushInterval = 15 * time.Second
)

// CrawlTask fetches each chapter page and stores content in the DB.
type CrawlTask struct {
	Config  Config
	Client  *Client
	DB      *DB
	StateDB *State
}

type workResult struct {
	item    QueueItem
	code    int
	err     error
	chapter *Chapter
}

type writeBuffer struct {
	chapters  []Chapter
	done      []struct{ url, typ string; code int }
	failed    []struct{ url, msg string }
	size      int
	lastFlush time.Time
}

func (wb *writeBuffer) add(r workResult) {
	if r.chapter != nil {
		wb.chapters = append(wb.chapters, *r.chapter)
	}
	if r.err != nil {
		wb.failed = append(wb.failed, struct{ url, msg string }{r.item.URL, r.err.Error()})
	} else {
		wb.done = append(wb.done, struct{ url, typ string; code int }{r.item.URL, r.item.EntityType, r.code})
	}
	wb.size++
}

func (wb *writeBuffer) ready() bool {
	return wb.size >= crawlFlushEvery || (wb.size > 0 && time.Since(wb.lastFlush) >= crawlFlushInterval)
}

func (wb *writeBuffer) flush(db *DB, state *State) error {
	if wb.size == 0 {
		return nil
	}
	for _, c := range wb.chapters {
		if err := db.UpsertChapter(c); err != nil {
			return fmt.Errorf("upsert chapter: %w", err)
		}
	}
	for _, d := range wb.done {
		_ = state.Done(d.url, d.code, d.typ)
	}
	for _, f := range wb.failed {
		_ = state.Fail(f.url, f.msg)
	}
	*wb = writeBuffer{lastFlush: time.Now()}
	return nil
}

// Run executes the crawl task.
func (t *CrawlTask) Run(ctx context.Context, emit func(*CrawlState)) (CrawlMetric, error) {
	start := time.Now()
	var done atomic.Int64
	var failed atomic.Int64
	var lastPending atomic.Int64
	var inFlight atomic.Int64

	workers := t.Config.Workers
	if workers <= 0 {
		workers = DefaultWorkers
	}

	// Derive edition ID from config BaseURL.
	editionID := EditionFromURL(t.Config.BaseURL)

	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				elapsed := time.Since(start).Seconds()
				rps := 0.0
				if elapsed > 0 {
					rps = float64(done.Load()) / elapsed
				}
				emit(&CrawlState{
					Done:    done.Load(),
					Pending: lastPending.Load(),
					Failed:  failed.Load(),
					RPS:     rps,
				})
			}
		}
	}()

	results := make(chan workResult, workers*8)
	writerDone := make(chan struct{})

	go func() {
		defer close(writerDone)
		buf := &writeBuffer{lastFlush: time.Now()}
		doFlush := func() {
			if err := buf.flush(t.DB, t.StateDB); err != nil {
				fmt.Printf("\nflush error: %v\n", err)
			}
		}
		for r := range results {
			if r.err != nil {
				failed.Add(1)
			} else {
				done.Add(1)
			}
			buf.add(r)
			if buf.ready() {
				doFlush()
			}
			inFlight.Add(-1)
		}
		doFlush()
	}()

	sem := make(chan struct{}, workers)
	var wg sync.WaitGroup
	popBatch := workers * 8
	if popBatch < 8 {
		popBatch = 8
	}

	for {
		if ctx.Err() != nil {
			break
		}
		items, _ := t.StateDB.Pop(popBatch)
		if len(items) == 0 {
			if inFlight.Load() == 0 {
				break
			}
			time.Sleep(50 * time.Millisecond)
			continue
		}
		inFlight.Add(int64(len(items)))
		pending, _, _, _ := t.StateDB.QueueStats()
		lastPending.Store(pending)

		for _, it := range items {
			sem <- struct{}{}
			wg.Add(1)
			go func(item QueueItem) {
				defer wg.Done()
				defer func() { <-sem }()
				results <- t.process(ctx, editionID, item)
			}(it)
		}
	}

	wg.Wait()
	close(results)
	<-writerDone

	return CrawlMetric{
		Done:     done.Load(),
		Failed:   failed.Load(),
		Duration: time.Since(start),
	}, nil
}

func (t *CrawlTask) process(ctx context.Context, editionID string, item QueueItem) workResult {
	res := workResult{item: item, code: 200}

	body, err := t.Client.FetchPage(ctx, item.URL)
	if err != nil {
		res.err = err
		return res
	}

	slug := item.Meta
	if slug == "" {
		// Extract slug from URL as fallback.
		parts := item.URL
		for len(parts) > 0 && parts[len(parts)-1] == '/' {
			parts = parts[:len(parts)-1]
		}
		for i := len(parts) - 1; i >= 0; i-- {
			if parts[i] == '/' {
				slug = parts[i+1:]
				break
			}
		}
		// Strip .html.
		if len(slug) > 5 && slug[len(slug)-5:] == ".html" {
			slug = slug[:len(slug)-5]
		}
	}

	ch, err := ParseChapter(editionID, slug, item.URL, body)
	if err != nil {
		res.err = fmt.Errorf("parse chapter: %w", err)
		return res
	}
	res.chapter = ch
	return res
}
