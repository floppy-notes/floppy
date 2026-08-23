# floppy

Local-first CLI for notes, tasks, and reminders in Markdown with YAML
frontmatter.

## Vault

Every command accepts `--vault <path>`, setting the root where `notes/`,
`tasks/`, `reminders/`, and the index (`floppy.db`, directly in that same
root, no subfolder) live.

Without `--vault`:
1. uses the `FLOPPY_VAULT` env var, if set;
2. otherwise, `~/.floppy`.

## Commands

### `floppy index --rebuild`

Walks the vault and rebuilds the SQLite index from scratch (atomic — if it
fails midway, the old index is preserved). Only full rebuild exists today;
running without `--rebuild` returns an error.

```
floppy --vault ./my-vault index --rebuild
```

### `floppy search <terms...>`

Full-text search (FTS5) over the body of already-indexed items.

- `--limit, -l` — max results (default: `20`)

```
floppy search team meeting --limit 5
```

### `floppy create note|task|reminder`

Builds a new item from the flags. **Doesn't write to the vault yet** for
now it only validates and prints what would be created.

Common flags across all three:
- `--tags, -t` tags for the item
- `--related, -r` ids of related items
- `--body, -b` item body
- `--quiet, -q` suppress output

`create task` requires `--due, -d` (`YYYY-MM-DD`).
`create reminder` requires `--remind-at` (`YYYY-MM-DD` or `YYYY-MM-DD HH:MM`).

```
floppy create task -t work -b "review PR" --due 2026-08-25
```

## Not implemented yet

- `list` list items by metadata/date
- `create` actually writing the file to the vault
