package papl_test

import (
	"testing"

	"github.com/tamnd/papl-cli/papl"
)

// tocHTML is a minimal mock of the PAPL TOC page.
const tocHTML = `<!DOCTYPE HTML>
<html>
<body id="scribble-racket-lang-org">
<div class="tocset">
<div class="tocview">
<div class="tocviewlisttopspace">
<div class="tocviewtitle">
<table><tr><td></td><td>
<a href="" class="tocviewselflink" data-pltdoc="x">Programming and Programming Languages</a>
</td></tr></table>
</div>
<div class="tocviewsublistonly" style="display: block;" id="tocview_0">
<table>
<tr><td align="right">1&nbsp;</td><td>
<a href="Introduction.html" class="tocviewlink" data-pltdoc="x">Introduction</a>
</td></tr>
<tr><td align="right">2&nbsp;</td><td>
<a href="getting-started.html" class="tocviewlink" data-pltdoc="x">Getting Started</a>
</td></tr>
<tr><td align="right">3&nbsp;</td><td>
<a href="Naming_Values.html" class="tocviewlink" data-pltdoc="x">Naming Values</a>
</td></tr>
</table>
</div>
</div>
</div>
<div class="maincolumn">
<p>Welcome to PAPL.</p>
</div>
</body>
</html>`

func TestParseTOC(t *testing.T) {
	baseURL := "https://papl.cs.brown.edu/2020/"
	toc, err := papl.ParseTOC(baseURL, []byte(tocHTML))
	if err != nil {
		t.Fatalf("ParseTOC: %v", err)
	}
	if len(toc.Chapters) == 0 {
		t.Fatal("ParseTOC returned 0 chapters")
	}

	// Check a known chapter.
	foundGS := false
	for _, ch := range toc.Chapters {
		if ch.Slug == "getting-started" {
			foundGS = true
			if ch.Title == "" {
				t.Errorf("getting-started has empty title")
			}
			if ch.URL == "" {
				t.Errorf("getting-started has empty URL")
			}
		}
	}
	if !foundGS {
		t.Errorf("getting-started not found; chapters: %+v", toc.Chapters)
	}
}

// chapterHTML is a minimal mock of a PAPL chapter page.
const chapterHTML = `<!DOCTYPE HTML>
<html>
<head><title>Getting Started - Programming and Programming Languages</title></head>
<body id="scribble-racket-lang-org">
<div class="maincolumn">
<div class="main">
<h2>Getting Started</h2>
<p>This chapter introduces the Pyret programming language and shows you how
to get started writing programs.</p>
<div class="Exercise">
<p>Exercise 1.1: Write a function that adds two numbers.</p>
<pre><code>fun add(a, b): a + b end</code></pre>
</div>
<div class="ExerciseBI">
<p>Do Now! Try running the code above.</p>
</div>
<pre>fun hello(): "hello" end</pre>
</div>
</div>
</body>
</html>`

func TestParseChapter(t *testing.T) {
	ch, err := papl.ParseChapter("2020", "getting-started",
		"https://papl.cs.brown.edu/2020/getting-started.html",
		[]byte(chapterHTML))
	if err != nil {
		t.Fatalf("ParseChapter: %v", err)
	}

	if ch.ID != "2020/getting-started" {
		t.Errorf("ID = %q, want %q", ch.ID, "2020/getting-started")
	}
	if ch.Title == "" {
		t.Errorf("Title is empty")
	}
	if ch.BodyMD == "" {
		t.Errorf("BodyMD is empty")
	}
	if ch.ExerciseCount == 0 {
		t.Errorf("ExerciseCount = 0, want > 0")
	}
	if ch.CodeBlockCount == 0 {
		t.Errorf("CodeBlockCount = 0, want > 0")
	}
	if !ch.HasDoNow {
		t.Errorf("HasDoNow = false, want true")
	}
}

func TestEditionFromURL(t *testing.T) {
	tests := []struct {
		url  string
		want string
	}{
		{"https://papl.cs.brown.edu/2020/", "2020"},
		{"https://papl.cs.brown.edu/2018/", "2018"},
		{"https://papl.cs.brown.edu/2020", "2020"},
	}
	for _, tt := range tests {
		got := papl.EditionFromURL(tt.url)
		if got != tt.want {
			t.Errorf("EditionFromURL(%q) = %q, want %q", tt.url, got, tt.want)
		}
	}
}
