# Using floppy from an agent

Exact commands, flags, output shapes, and exit codes for anything that
shells out to `floppy`: an agent, an MCP tool wrapper, a script. For the
conceptual overview, see the [top-level README](../../README.md).

## Invocation

```
floppy [--vault <path>] <command> [flags]
```

- `--vault <path>` is a persistent flag, valid on every subcommand. If
  omitted, floppy resolves the vault root from `FLOPPY_VAULT`, then
  `~/.floppy`.
- Exit code is `0` on success, `1` on any error. On error, a single line is
  written to stderr in the form `floppy: <message>`. Nothing is written to
  stdout.

```
$ floppy create task --title "x"
floppy: required flag(s) "due" not set
$ echo $?
1
```

## Output format

Commands that print results (`search`, `list`) emit JSON on stdout, one
object per invocation:

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

Guarantees:

- `items` is always an array, never `null`. No results means `{"items": [],
  "count": 0}` and exit `0`. An empty result is not an error.
- Every field above is always present. The nullable ones (`due`, `status`,
  `remind_at`) are `null` when absent rather than omitted, so you can read
  them without an existence check. `due` is only set on tasks, `remind_at`
  only on reminders, `status` on tasks and reminders.
- `path` is absolute and points at the Markdown file backing the item.
- The response is an object, not a bare array, so future fields can be added
  alongside `items` without breaking your parser. Ignore keys you don't know.

Errors go to stderr (prefixed `floppy:`) with a non-zero exit code, never to
stdout. Stdout holds either valid JSON or nothing.

The `.md` files remain the source of truth: this JSON is derived from the
SQLite index. If you need the body of an item, read `path`; `search`/`list`
return metadata only.

## Commands

### Create an item

```
floppy create note    --title <string> [--body <string>] [--tags a,b] [--related id1,id2] [--quiet]
floppy create task     --title <string> --due <YYYY-MM-DD> [--body ...] [--tags ...] [--related ...] [--quiet]
floppy create reminder --title <string> --remind-at <YYYY-MM-DD[ HH:MM]> [--body ...] [--tags ...] [--related ...] [--quiet]
```

- `--title` is required for all three; it's slugified into the file's `id`.
- On success, exit `0`. The file is written to
  `<vault>/<notes|tasks|reminders>/<YYYY>/<MM>/<id>.md`, and the item is
  indexed immediately, so it shows up in `search`/`list` right away. No
  extra step needed.
- The created item is printed to stdout as a single JSON object, with the
  same fields as a `list` entry and no `items`/`count` wrapper. Read the
  generated `id` from there instead of guessing the slug or re-querying.
  Pass `--quiet` to suppress it when you only care about the exit code.
- `id` generation reads the target day's folder to compute the next
  sequence number, then writes the file. It is not atomic. Two `create`
  calls fired at effectively the same instant, for the same type and day,
  can race. Fine for interactive/sequential use; don't fire concurrent
  `create` calls from a batch job without serializing them.
- Files whose `id` does not follow `<YYYY-MM-DD>-<NNNN>-<slug>` are ignored
  when computing that sequence, so a hand-named file in the folder does not
  block `create`.
- A title with no letters or digits (`"!!!"`) slugifies to `untitled`, and
  the `title` in the frontmatter keeps whatever you passed.

### Update an item

```
floppy update note     --id <id> [--title <string>] [--body <string>] [--tags a,b] [--related id1,id2] [--replace] [--quiet]
floppy update task     --id <id> [--title ...] [--body ...] [--tags ...] [--related ...] [--due <YYYY-MM-DD>] [--status open|done|archived] [--replace] [--quiet]
floppy update reminder --id <id> [--title ...] [--body ...] [--tags ...] [--related ...] [--remind-at <YYYY-MM-DD[ HH:MM]>] [--status pending|fired|dismissed] [--replace] [--quiet]
```

- `--id` is required for all three; it must match an existing item of that
  type, or the command errors.
- By default, `--body`, `--tags`, and `--related` are appended to the
  existing values. Pass `--replace` to overwrite them instead. `--title`
  always overwrites when set.
- `--status` only exists on `update task` and `update reminder`. It is
  validated against the type's allowed values (see [Data
  model](#data-model)); an invalid value is rejected with a non-zero exit
  and nothing is written. The check runs before the file is touched, so a
  bad flag does not leave an item half-updated.
- On success, the file is rewritten in place and reindexed immediately, the
  same guarantee as `create`, and the updated item is printed to stdout as
  a single JSON object (`--quiet` suppresses it).

### Rebuild the index

```
floppy index --rebuild
```

Full rebuild only; there is no incremental "just this file" mode. `create`
and `update` already index as they write, so you only need this after
editing a `.md` file by hand outside of `floppy`, or to recover a deleted
or corrupted index. Safe to call anytime: it's a full walk-and-reindex
inside one transaction (a mid-rebuild failure leaves the previous index
untouched).

### Search

```
floppy search <term> [<term> ...] [--limit N]   # default limit: 20
```

Full-text (FTS5) over `title` + body, ranked by relevance. Requires at
least one term. There is no "list everything" mode here; use `list` for
that.

Terms are passed to FTS5 as written, so FTS5 operators (`AND`, `OR`,
`NEAR`) work, and so do its syntax errors: an unbalanced quote or a bare
`*` exits non-zero with `invalid search query`. Quote your terms if you
want them treated literally.

### List

```
floppy list [--type note|task|reminder] [--status <value>] [--tag <value>]
            [--since <date>] [--until <date>] [--limit N]   # default limit: 10
```

Metadata filtering, with no ranking involved. All flags are optional and
combine with AND.

- `--status` accepts either a task status (`open`/`done`/`archived`) or a
  reminder status (`pending`/`fired`/`dismissed`). It does not cross-check
  against `--type`, so `--type note --status open` is accepted but returns
  nothing (notes never have a status).
- `--since`/`--until` accept `YYYY-MM-DD`, `YYYY-MM-DD HH:MM`, or
  `YYYY-MM-DDTHH:MM`, and filter on `created`, not on `due` or
  `remind_at`, even when `--type task`/`--type reminder` is set.
- `--tag` takes exactly one tag, not a list.

## Data model

An item is one `.md` file: YAML frontmatter between `---` lines, then a
Markdown body.

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

Body content, plain Markdown.
```

| Field | Type | Applies to | Constraints |
| --- | --- | --- | --- |
| `id` | string | all | `<YYYY-MM-DD>-<NNNN>-<title-slug>`; unique per vault; don't hand-construct one, let `create` generate it |
| `title` | string | all | required, non-empty |
| `type` | string | all | exactly `note`, `task`, or `reminder` |
| `created` | string | all | RFC3339 (`2006-01-02T15:04:05-07:00`), local time, not UTC |
| `tags` | string list | all | optional |
| `related` | string list | all | optional; ids of other items, not validated against the vault (broken references are allowed) |
| `due` | string | `task` only | `YYYY-MM-DD`; must be absent on `note`/`reminder` |
| `status` | string | `task`, `reminder` | task: `open`/`done`/`archived`; reminder: `pending`/`fired`/`dismissed`; must be absent on `note` |
| `remind_at` | string | `reminder` only | RFC3339, local time; must be absent on `note`/`task` |

If you're generating or editing a `.md` file directly instead of going
through `create`, match this exactly. The same validation that `create`
runs on write also runs on every file during `index --rebuild`, and a
malformed one is skipped (with a warning on stderr) rather than crashing the
rebuild.

## Recommended workflow for an agent

1. `create` (or `update`) the item(s) you need. Both index as they write,
   so the item is queryable right after the command returns `0`.
2. `search`/`list` to confirm, or read the file back directly if you need
   the exact frontmatter you just wrote.
3. Only run `index --rebuild` if you edited `.md` files by hand outside of
   `floppy`, or suspect the index is out of sync with the vault.
