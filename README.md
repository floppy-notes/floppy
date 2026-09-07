# floppy

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
  - [`floppy index --rebuild`](#floppy-index---rebuild)
  - [`floppy search`](#floppy-search-terms)
  - [`floppy list`](#floppy-list)
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

## Quick start

```sh
floppy create note --title "Sprint retro notes" --tags team,retro \
  --body "Went well: deploy cadence. To improve: on-call handoff."

floppy create task --title "Fix flaky CI job" --due 2026-09-15 --tags ci

floppy index --rebuild

floppy search flaky
floppy list --type task --status open
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

The filename mirrors the frontmatter `id` on purpose. It's there for human
browsing, not for queries. Queries always go through the frontmatter (or,
for search/filtering, the index built from it).

## Commands

### `floppy create note|task|reminder`

Creates a new item: generates its `id`, validates the frontmatter, writes
the `.md` file to the vault. It does **not** touch the index. Run
`floppy index --rebuild` afterward to make the new item searchable/listable.

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

### `floppy index --rebuild`

Walks the vault, parses every `.md` file, and rebuilds the SQLite index from
scratch inside a single transaction (if it fails midway, the previous index
is left untouched). Only a full rebuild exists today. There is no
incremental mode yet, so running the command without `--rebuild` returns an
error instead of silently doing nothing.

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

Lists items by metadata, without ranking. It's the counterpart to `search`
for "browse/filter" instead of "find by relevance."

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

- Incremental indexing (today, any change requires `index --rebuild`)
- Structured (JSON) output, for scripting and agent consumption
- Backlinks/graph queries over `related`
