package papl

import (
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"

	htmltomd "github.com/JohannesKaufmann/html-to-markdown/v2/converter"
	"github.com/JohannesKaufmann/html-to-markdown/v2/plugin/base"
	"github.com/JohannesKaufmann/html-to-markdown/v2/plugin/commonmark"
	"github.com/JohannesKaufmann/html-to-markdown/v2/plugin/table"
	"github.com/PuerkitoBio/goquery"
)

// mdPool is a pool of HTML-to-Markdown converters.
var mdPool = sync.Pool{New: func() any {
	return htmltomd.NewConverter(htmltomd.WithPlugins(
		base.NewBasePlugin(),
		commonmark.NewCommonmarkPlugin(),
		table.NewTablePlugin(),
	))
}}

// HTMLToMarkdown converts an HTML string to Markdown.
func HTMLToMarkdown(html string) string {
	if strings.TrimSpace(html) == "" {
		return ""
	}
	conv := mdPool.Get().(*htmltomd.Converter)
	defer mdPool.Put(conv)
	md, err := conv.ConvertString(html)
	if err != nil {
		return html
	}
	return strings.TrimSpace(md)
}

// TOCChapter is one entry from the table of contents.
type TOCChapter struct {
	Slug   string
	URL    string
	Title  string
	Number int
}

// TOCResult holds chapter URLs found on the root page.
type TOCResult struct {
	Chapters []TOCChapter
}

// ParseTOC parses the PAPL root page HTML and extracts chapter links.
func ParseTOC(baseURL string, body []byte) (*TOCResult, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
	if err != nil {
		return nil, fmt.Errorf("parse TOC HTML: %w", err)
	}

	base := strings.TrimRight(baseURL, "/") + "/"
	seen := make(map[string]bool)
	var chapters []TOCChapter
	num := 0

	// Try multiple selectors for the TOC links.
	selectors := []string{
		"a.tocviewlink", "a.tocviewselflink", ".toc a", "nav a",
		"#toc a", ".tocview a", "ol li a", "ul li a",
	}
	for _, sel := range selectors {
		doc.Find(sel).Each(func(_ int, s *goquery.Selection) {
			href, exists := s.Attr("href")
			if !exists || href == "" {
				return
			}
			// Skip anchors and absolute external links.
			if strings.HasPrefix(href, "#") {
				return
			}
			if strings.HasPrefix(href, "http") && !strings.Contains(href, "papl.cs.brown.edu") {
				return
			}

			var chapterURL string
			if strings.HasPrefix(href, "http") {
				chapterURL = href
			} else {
				chapterURL = base + strings.TrimPrefix(href, "/")
			}
			// Remove fragment.
			if idx := strings.Index(chapterURL, "#"); idx >= 0 {
				chapterURL = chapterURL[:idx]
			}
			chapterURL = strings.TrimRight(chapterURL, "/")

			if seen[chapterURL] {
				return
			}
			slug := slugFromURL(chapterURL)
			if slug == "" {
				return
			}
			seen[chapterURL] = true
			num++
			chapters = append(chapters, TOCChapter{
				Slug:   slug,
				URL:    chapterURL,
				Title:  strings.TrimSpace(s.Text()),
				Number: num,
			})
		})
		if len(chapters) > 0 {
			break
		}
	}

	return &TOCResult{Chapters: chapters}, nil
}

// ParseChapter parses a chapter HTML page and extracts its content.
func ParseChapter(editionID, slug, chapterURL string, body []byte) (*Chapter, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
	if err != nil {
		return nil, fmt.Errorf("parse chapter HTML: %w", err)
	}

	// Extract title.
	title := strings.TrimSpace(doc.Find(".maincolumn h2, .main h2, .main h1, h1").First().Text())
	if title == "" {
		title = strings.TrimSpace(doc.Find("title").Text())
		// Strip site suffix.
		if idx := strings.Index(title, " - "); idx > 0 {
			title = strings.TrimSpace(title[:idx])
		}
	}

	// Extract content area.
	var bodyHTML string
	var contentSel *goquery.Selection
	for _, sel := range []string{".maincolumn .main", ".main", "article", "body"} {
		if s := doc.Find(sel); s.Length() > 0 {
			contentSel = s.First()
			break
		}
	}
	if contentSel != nil {
		// Remove inner TOC.
		contentSel.Find(".tocset, .toc, nav").Remove()
		h, _ := contentSel.Html()
		bodyHTML = strings.TrimSpace(h)
	}
	bodyMD := HTMLToMarkdown(bodyHTML)

	// Extract description: first paragraph with > 40 chars.
	description := ""
	doc.Find("p").Each(func(_ int, s *goquery.Selection) {
		if description != "" {
			return
		}
		txt := strings.TrimSpace(s.Text())
		if len(txt) > 40 {
			description = txt
		}
	})

	// Count exercises and "Do Now!" prompts.
	exerciseCount := doc.Find("div.Exercise, div.ExerciseBI, div.SubExercise, .exercise, .problem").Length()
	bodyText := doc.Text()
	doNowCount := strings.Count(bodyText, "Do Now")
	if doNowCount == 0 {
		doNowCount = strings.Count(bodyText, "Do now")
	}
	hasDoNow := doNowCount > 0
	if hasDoNow {
		exerciseCount += doNowCount
	}

	// Count code blocks.
	codeBlockCount := doc.Find("pre, code").Length()

	chapterID := editionID + "/" + slug

	return &Chapter{
		ID:             chapterID,
		EditionID:      editionID,
		Slug:           slug,
		Title:          title,
		URL:            chapterURL,
		Description:    description,
		Topics:         "[]",
		ExerciseCount:  exerciseCount,
		CodeBlockCount: codeBlockCount,
		HasDoNow:       hasDoNow,
		DoNowCount:     doNowCount,
		BodyMD:         bodyMD,
		FetchedAt:      time.Now(),
	}, nil
}

// EditionFromURL extracts the edition year from a base URL.
// "https://papl.cs.brown.edu/2020/" -> "2020"
func EditionFromURL(baseURL string) string {
	u, err := url.Parse(baseURL)
	if err != nil {
		return "2020"
	}
	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	for _, p := range parts {
		if len(p) == 4 {
			return p
		}
	}
	return "2020"
}

// slugFromURL extracts the chapter slug from a full URL.
// "https://papl.cs.brown.edu/2020/getting-started.html" -> "getting-started"
func slugFromURL(rawURL string) string {
	rawURL = strings.TrimRight(rawURL, "/")
	idx := strings.LastIndex(rawURL, "/")
	if idx < 0 {
		return ""
	}
	slug := rawURL[idx+1:]
	slug = strings.TrimSuffix(slug, ".html")
	return slug
}
