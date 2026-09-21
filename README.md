<p align="center">
  <img src="docs/assets/logo.png" alt="floppy" width="120">
</p>

<p align="center">
  <a href="https://github.com/floppy-notes/floppy/actions/workflows/ci.yml"><img src="https://github.com/floppy-notes/floppy/actions/workflows/ci.yml/badge.svg" alt="CI"></a>
  <a href="https://codecov.io/gh/floppy-notes/floppy"><img src="https://codecov.io/gh/floppy-notes/floppy/branch/main/graph/badge.svg" alt="Coverage"></a>
  <a href="https://pkg.go.dev/github.com/floppy-notes/floppy"><img src="https://pkg.go.dev/badge/github.com/floppy-notes/floppy.svg" alt="Go Reference"></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/license-MIT-blue.svg" alt="License: MIT"></a>
</p>

Local-first CLI for notes, tasks, and reminders, stored as plain Markdown
files with YAML frontmatter.

- **Local-first**: every item is a `.md` file on disk. No server, no
  account, no lock-in.
- **Fast to query**: a disposable SQLite (FTS5) index sits on top of the
  vault for full-text search and metadata filtering.
- **Your files stay yours**: the index is derived data. Delete it any time
  and rebuild it from the Markdown files, which remain the single source of
  truth.

## Table of contents

- [Installation](#installation)
- [Quick start](#quick-start)
- [The vault](#the-vault)
- [Commands](#commands)
  - [`floppy create`](#floppy-create-notetaskreminder)
  - [`floppy update`](#floppy-update-notetaskreminder)
  - [`floppy index --rebuild`](#floppy-index---rebuild)
  - [`floppy search`](#floppy-search-terms)
  - [`floppy list`](#floppy-list)
  - [Output](#output)
- [Data model](#data-model)
- [AI / agent integration](#ai--agent-integration)
- [Roadmap](#roadmap)

## Installation

Requires Go 1.26+.

```sh
go install github.com/floppy-notes/floppy/cmd/floppy@latest
```

Or build from source:

```sh
git clone https://github.com/floppy-notes/floppy.git
cd floppy
go build -o floppy ./cmd/floppy
```

`floppy --version` reports the build. A source build picks up the commit from
git on its own; release builds stamp the tag with
`-ldflags "-X github.com/floppy-notes/floppy/internal/cli.version=v0.1.0"`.

## Quick start

```sh
floppy create note --title "Sprint retro notes" --tags team,retro \
  --body "Went well: deploy cadence. To improve: on-call handoff."

floppy create task --title "Fix flaky CI job" --due 2026-09-15 --tags ci

floppy search flaky
floppy list --type task --status open

floppy update task --id <id> --status done
```

## The vault

A vault is a directory holding `notes/`, `tasks/`, `reminders/`, and the
SQLite index (`floppy.db`), all directly under its root, with no hidden
subfolder in between.

Every command accepts `--vault <path>` to point at one. Without it, floppy
resolves the root in this order:

1. `--vault <path>` flag
2. `FLOPPY_VAULT` environment variable
3. `~/.floppy` (default)

Items are stored under `<type>/<YYYY>/<MM>/<id>.md`, e.g.:

```
my-vault/
├── floppy.db
├── notes/2026/09/2026-09-07-0001-sprint-retro-notes.md
├── tasks/2026/09/2026-09-07-0001-fix-flaky-ci-job.md
└── reminders/2026/09/2026-09-08-0001-call-the-dentist.md
```

The filename mirrors the frontmatter `id`, which keeps the vault readable
when you browse it by hand. Queries always go through the frontmatter, or
through the index built from it.

## Commands

### `floppy create note|task|reminder`

Creates a new item: generates its `id`, validates the frontmatter, writes
the `.md` file to the vault, and indexes it immediately. It's searchable
and listable as soon as the command returns.

Common flags:

| Flag | Shorthand | Required | Description |
| --- | --- | --- | --- |
| `--title` | `-t` | yes | item title (used to derive the `id`) |
| `--body` | `-b` | no | item body (Markdown, freeform) |
| `--tags` | | no | comma-separated tags |
| `--related` | `-r` | no | comma-separated ids of related items |
| `--quiet` | `-q` | no | suppress output |

Type-specific flags:

| Command | Flag | Required | Format |
| --- | --- | --- | --- |
| `create task` | `--due, -d` | yes | `YYYY-MM-DD` |
| `create reminder` | `--remind-at` | yes | `YYYY-MM-DD` or `YYYY-MM-DD HH:MM` |

```sh
floppy create task --title "Review onboarding PR" -d 2026-09-15 --tags work
```

### `floppy update note|task|reminder`

Updates an existing item by `id`: rereads its `.md` file, applies the given
changes, rewrites the file, and reindexes it immediately.

Common flags:

| Flag | Shorthand | Required | Description |
| --- | --- | --- | --- |
| `--id` | `-i` | yes | id of the item to update |
| `--title` | `-t` | no | new title |
| `--body` | `-b` | no | new body content (appended by default; see `--replace`) |
| `--tags` | | no | tags to add (or replace, with `--replace`) |
| `--related` | | no | related ids to add (or replace, with `--replace`) |
| `--replace` | `-r` | no | replace `body`/`tags`/`related` instead of appending to them |
| `--quiet` | `-q` | no | suppress output |

Type-specific flags:

| Command | Flag | Format |
| --- | --- | --- |
| `update task` | `--due, -d` | `YYYY-MM-DD` |
| `update task` | `--status, -s` | `open`, `done`, or `archived` |
| `update reminder` | `--remind-at` | `YYYY-MM-DD` or `YYYY-MM-DD HH:MM` |
| `update reminder` | `--status, -s` | `pending`, `fired`, or `dismissed` |

```sh
floppy update task --id 2026-09-15-0001-review-onboarding-pr --status done
```

### `floppy index --rebuild`

Walks the vault, parses every `.md` file, and rebuilds the SQLite index from
scratch inside a single transaction (if it fails midway, the previous index
is left untouched). Only a full rebuild exists today; there is no
incremental "just this file" mode, so running the command without
`--rebuild` returns an error instead of silently doing nothing. `create` and
`update` already index as they go, so a full rebuild is only needed after
editing `.md` files by hand or to recover a deleted/corrupted index.

```sh
floppy index --rebuild
```

### `floppy search <terms...>`

Full-text search (SQLite FTS5) over item titles and bodies, ranked by
relevance.

| Flag | Shorthand | Default | Description |
| --- | --- | --- | --- |
| `--limit` | `-l` | `20` | max results |

```sh
floppy search onboarding review --limit 5
```

### `floppy list`

Lists items by metadata, without ranking. Use `search` when you want
results ordered by relevance.

| Flag | Shorthand | Description |
| --- | --- | --- |
| `--type` | `-t` | `note`, `task`, or `reminder` |
| `--status` | `-s` | task status (`open`, `done`, `archived`) or reminder status (`pending`, `fired`, `dismissed`) |
| `--tag` | | filter by a single tag |
| `--since` | | only items created on/after this date |
| `--until` | | only items created on/before this date |
| `--limit` | `-l` | max results (default `10`) |

`--since`/`--until` accept `YYYY-MM-DD`, `YYYY-MM-DD HH:MM`, or
`YYYY-MM-DDTHH:MM`, and are matched against `created` (not `due`/`remind_at`).

```sh
floppy list --type task --status open
floppy list --tag golang --since 2026-09-01
```

### Output

`search` and `list` print JSON on stdout. The object carries `items` and
`count`, and holds metadata only (read `path` for the body):

```json
{
  "items": [
    {
      "id": "2026-09-16-0001-fix-flaky-ci-job",
      "title": "Fix flaky CI job",
      "type": "task",
      "created": "2026-09-16T13:57:53-03:00",
      "due": "2026-09-30",
      "status": "open",
      "remind_at": null,
      "path": "/home/me/.floppy/tasks/2026/09/2026-09-16-0001-fix-flaky-ci-job.md"
    }
  ],
  "count": 1
}
```

`items` is always an array, so no results is `{"items": [], "count": 0}` with
exit `0`. Errors go to stderr, so stdout stays parseable. Pipe it into `jq`:

```sh
floppy list --type task --status open | jq -r '.items[].title'
```

`create` and `update` print the resulting item as a single JSON object,
which is where you read the generated `id` from. Use `--quiet` to suppress
it.

See [`docs/ai/`](docs/ai/) for the full contract.

## Data model

Every item is a Markdown file with YAML frontmatter:

```yaml
---
id: 2026-09-15-0001-review-onboarding-pr
title: Review onboarding PR
type: task
created: 2026-09-07T14:32:07-03:00
due: 2026-09-15
status: open
tags: [work]
related: []
---

Body content goes here, as plain Markdown.
```

| Field | Applies to | Notes |
| --- | --- | --- |
| `id` | all | `<created-date>-<NNNN>-<title-slug>`, sequence resets per type per day |
| `title` | all | required |
| `type` | all | `note`, `task`, or `reminder` |
| `created` | all | RFC3339, in local time |
| `tags` | all | optional |
| `related` | all | optional, ids of other items |
| `due` | `task` only | `YYYY-MM-DD` |
| `status` | `task`, `reminder` | task: `open`/`done`/`archived`; reminder: `pending`/`fired`/`dismissed` |
| `remind_at` | `reminder` only | RFC3339, in local time |

A note may not carry `due`, `status`, or `remind_at`. A task may not carry
`remind_at`. A reminder may not carry `due`. Validation is enforced both when
writing a new item and when the vault is walked for indexing.

## AI / agent integration

If you're building an agent or tool that shells out to `floppy`, see
[`docs/ai/`](docs/ai/) for a reference written specifically for that: exact
command/flag contracts, output shapes, and current limitations to code
around.

## Roadmap

- Incremental indexing for hand-edited files (`create`/`update` already
  index as they go; only manual `.md` edits still require `index --rebuild`)
- Backlinks/graph queries over `related`
