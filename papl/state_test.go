package papl_test

import (
	"path/filepath"
	"testing"

	"github.com/tamnd/papl-cli/papl"
)

func TestStateLifecycle(t *testing.T) {
	dir := t.TempDir()
	statePath := filepath.Join(dir, "state.db")

	state, err := papl.OpenState(statePath)
	if err != nil {
		t.Fatalf("OpenState: %v", err)
	}
	defer state.Close()

	// Initially empty.
	p, ip, d, f := state.QueueStats()
	if p != 0 || ip != 0 || d != 0 || f != 0 {
		t.Errorf("initial stats = (%d,%d,%d,%d), want all 0", p, ip, d, f)
	}

	url1 := "https://papl.cs.brown.edu/2020/getting-started.html"
	url2 := "https://papl.cs.brown.edu/2020/Introduction.html"

	_ = state.Enqueue(url1, papl.EntityChapter, papl.PriorityChapter, "getting-started")
	_ = state.Enqueue(url2, papl.EntityChapter, papl.PriorityChapter, "Introduction")
	// Duplicate enqueue is a no-op.
	_ = state.Enqueue(url1, papl.EntityChapter, papl.PriorityChapter, "getting-started")

	p, _, _, _ = state.QueueStats()
	if p != 2 {
		t.Errorf("pending = %d, want 2", p)
	}

	// Pop 1.
	items, err := state.Pop(1)
	if err != nil {
		t.Fatalf("Pop: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("Pop returned %d items, want 1", len(items))
	}

	// Mark done.
	if err := state.Done(items[0].URL, 200, papl.EntityChapter); err != nil {
		t.Fatalf("Done: %v", err)
	}

	_, _, d, _ = state.QueueStats()
	if d != 1 {
		t.Errorf("done = %d, want 1", d)
	}

	if !state.IsVisited(items[0].URL) {
		t.Errorf("IsVisited(%q) = false, want true", items[0].URL)
	}

	// Pop and fail.
	items2, _ := state.Pop(1)
	if len(items2) != 1 {
		t.Fatalf("second Pop returned %d items, want 1", len(items2))
	}
	if err := state.Fail(items2[0].URL, "network error"); err != nil {
		t.Fatalf("Fail: %v", err)
	}
	// After first failure, back to pending.
	p2, _, _, _ := state.QueueStats()
	if p2 != 1 {
		t.Errorf("after Fail, pending = %d, want 1", p2)
	}

	// ListQueue.
	qItems, err := state.ListQueue("pending", 10)
	if err != nil {
		t.Fatalf("ListQueue: %v", err)
	}
	if len(qItems) != 1 {
		t.Errorf("ListQueue pending = %d, want 1", len(qItems))
	}

	// ResetFailed: no failed items yet (failed only after 3 attempts).
	n, err := state.ResetFailed()
	if err != nil {
		t.Fatalf("ResetFailed: %v", err)
	}
	_ = n // 0 since nothing is actually failed
}
