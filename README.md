# papl

Programming and Programming Languages (PAPL) textbook archiver

`papl` is a single pure-Go binary. It reads public papl data
over plain HTTPS, shapes it into clean records, and prints output that pipes
into the rest of your tools. No API key, nothing to run alongside it.

The same package is also a [resource-URI driver](#use-it-as-a-resource-uri-driver),
so a host program like [ant](https://github.com/tamnd/ant) can address
papl as `papl://` URIs.

## Install

```bash
go install github.com/tamnd/papl-cli/cmd/papl@latest
```

Or grab a prebuilt binary from the [releases](https://github.com/tamnd/papl-cli/releases), or run
the container image:

```bash
docker run --rm ghcr.io/tamnd/papl:latest --help
```

## Usage

```bash
papl page <path>                      # fetch one page as a record
papl page <path> -o json              # as JSON, ready for jq
papl page <path> --template '{{.Body}}'  # just the readable body text
papl links <path>                     # the pages it links to, one per line
papl --help                           # the whole command tree
```

Every command shares one output contract: `-o table|json|jsonl|csv|tsv|url|raw`,
`--fields` to pick columns, `--template` for a custom line, and `-n` to limit.
The default adapts to where output goes (a table on a terminal, JSONL in a
pipe), so the same command reads well by hand and parses cleanly downstream.

This is a fresh scaffold. It ships one example resource type, `page`, wired end
to end. Model the real papl records in `papl/` and declare their
operations in `papl/domain.go`; each one becomes a command, an HTTP
route, and an MCP tool at once.

## Serve it

The same operations are available over HTTP and as an MCP tool set for agents,
with no extra code:

```bash
papl serve --addr :7777    # GET /v1/page/<path>  returns NDJSON
papl mcp                   # speak MCP over stdio
```

## Use it as a resource-URI driver

`papl` registers a `papl` domain the way a program registers a
database driver with `database/sql`. A host enables it with one blank import:

```go
import _ "github.com/tamnd/papl-cli/papl"
```

Then [ant](https://github.com/tamnd/ant) (or any program that links the package)
dereferences `papl://` URIs without knowing anything about papl:

```bash
ant get papl://page/<path>   # fetch the record
ant cat papl://page/<path>   # just the body text
ant ls  papl://page/<path>   # the pages it links to, each addressable
ant url papl://page/<path>   # the live https URL
```

## Development

```
cmd/papl/   thin main: hands cli.NewApp to kit.Run
cli/                 assembles the kit App from the papl domain
papl/                the library: HTTP client, data models, and domain.go (the driver)
docs/                tago documentation site
```

```bash
make build      # ./bin/papl
make test       # go test ./...
make vet        # go vet ./...
```

## Releasing

Push a version tag and GitHub Actions runs GoReleaser, which builds the
archives, Linux packages, the multi-arch GHCR image, checksums, SBOMs, and a
cosign signature:

```bash
git tag v0.1.0
git push --tags
```

The Homebrew and Scoop steps self-disable until their tokens exist, so the first
release works with no extra secrets.

## License

Apache-2.0. See [LICENSE](LICENSE).
