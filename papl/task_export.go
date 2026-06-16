package papl

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ExportTask writes chapter records to markdown files.
type ExportTask struct {
	Config Config
	DB     *DB
}

// Run executes the export task.
func (t *ExportTask) Run(ctx context.Context, emit func(*ExportState)) (ExportMetric, error) {
	start := time.Now()
	state := &ExportState{}

	base := t.Config.ExportDir
	if err := os.MkdirAll(base, 0o755); err != nil {
		return ExportMetric{}, fmt.Errorf("mkdir export: %w", err)
	}

	chapters, err := t.DB.ListChapters()
	if err != nil {
		return ExportMetric{}, fmt.Errorf("list chapters: %w", err)
	}

	written := 0
	for _, c := range chapters {
		if ctx.Err() != nil {
			break
		}

		fname := c.Slug
		if fname == "" {
			fname = strings.ReplaceAll(c.ID, "/", "_")
		}
		fullPath := filepath.Join(base, fname+".md")

		state.Current = fname + ".md"
		emit(state)

		content := renderChapterMarkdown(c)
		if err := writeFile(fullPath, content); err == nil {
			written++
			state.Written = written
		}
	}

	return ExportMetric{Files: written, Duration: time.Since(start)}, nil
}

func renderChapterMarkdown(c Chapter) string {
	var sb strings.Builder
	sb.WriteString("---\n")
	fmt.Fprintf(&sb, "title: %q\n", c.Title)
	fmt.Fprintf(&sb, "edition: %s\n", c.EditionID)
	if c.Part != "" {
		fmt.Fprintf(&sb, "part: %s\n", c.Part)
	}
	fmt.Fprintf(&sb, "url: %s\n", c.URL)
	fmt.Fprintf(&sb, "exercise_count: %d\n", c.ExerciseCount)
	fmt.Fprintf(&sb, "code_block_count: %d\n", c.CodeBlockCount)
	fmt.Fprintf(&sb, "has_do_now: %v\n", c.HasDoNow)
	fmt.Fprintf(&sb, "do_now_count: %d\n", c.DoNowCount)
	fmt.Fprintf(&sb, "fetched_at: %s\n", c.FetchedAt.UTC().Format(time.RFC3339))
	sb.WriteString("---\n\n")
	if c.Title != "" {
		fmt.Fprintf(&sb, "# %s\n\n", c.Title)
	}
	if c.Description != "" {
		fmt.Fprintf(&sb, "%s\n\n", c.Description)
	}
	if c.BodyMD != "" {
		sb.WriteString(c.BodyMD)
		sb.WriteString("\n")
	}
	return sb.String()
}

func writeFile(path, content string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(content), 0o644)
}
