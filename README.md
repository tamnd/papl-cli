# papl

Programming and Programming Languages (PAPL) textbook archiver

`papl` is a single pure-Go binary that fetches and locally archives the
"Programming and Programming Languages" textbook by Shriram Krishnamurthi
(Brown University) from `papl.cs.brown.edu`. Every chapter is stored as
structured Markdown in a local SQLite database.

No API key required.

## Install

```bash
go install github.com/tamnd/papl-cli/cmd/papl@latest
```

Or grab a prebuilt binary from the [releases](https://github.com/tamnd/papl-cli/releases).

## Quick start

```bash
# Step 1: parse table of contents and enqueue chapter URLs
papl seed

# Step 2: fetch each chapter and store content in DB
papl crawl

# Step 3: write all chapters to markdown files
papl export

# Check progress at any time
papl info
```

## Commands

### `papl seed`

Fetch the TOC page, discover all chapter URLs, and enqueue unseen ones.

```
papl seed [--db PATH] [--state PATH] [--base-url URL] [--delay MS]
```

### `papl crawl`

Fetch each enqueued chapter, convert HTML to Markdown, store in DB.

```
papl crawl [--db PATH] [--state PATH] [--base-url URL] [--delay MS]
           [--timeout S] [--workers N]
```

### `papl export`

Write each chapter to a `.md` file named by slug.

```
papl export [--db PATH] [--export-dir PATH]
```

Output: `$HOME/data/papl/export/{slug}.md`

Each file has front-matter:
```yaml
---
title: "Getting Started"
edition: 2020
url: https://papl.cs.brown.edu/2020/getting-started.html
exercise_count: 3
code_block_count: 12
has_do_now: true
do_now_count: 2
fetched_at: 2026-06-14T10:00:00Z
---
```

### `papl info`

Show DB stats and queue depth.

```
papl info [--db PATH] [--state PATH]
```

### `papl queue`

List queue items by status.

```
papl queue [--status pending|done|failed|in_progress] [--limit N]
```

### `papl reset-failed`

Reset all failed items to pending for retry.

```
papl reset-failed [--state PATH]
```

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--db` | `$HOME/data/papl/papl.db` | SQLite database |
| `--state` | `$HOME/data/papl/state.db` | Crawl-queue database |
| `--export-dir` | `$HOME/data/papl/export` | Markdown output directory |
| `--base-url` | `https://papl.cs.brown.edu/2020/` | PAPL edition base URL |
| `--delay` | `500` | Delay between requests (ms) |
| `--timeout` | `30` | HTTP timeout (seconds) |
| `--workers` | `4` | Parallel chapter fetch workers |

## Multiple editions

Use `--base-url` to archive older editions:

```bash
papl seed --base-url https://papl.cs.brown.edu/2018/
papl crawl
```

## Development

```
cmd/papl/    main entry point
cli/         cobra command tree
papl/        library: client, parser, DB, state, tasks
```

```bash
make build      # ./bin/papl
make test       # go test ./...
make vet        # go vet ./...
```

## License

Apache-2.0. See [LICENSE](LICENSE).
