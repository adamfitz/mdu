# mdu

`mdu` is a simple CLI tool for reading and updating EPUB file metadata. It supports common metadata fields such as author, title, series, series index, summary, ISBN, and more. You can also render metadata as a formatted table and optionally write it to a file.

---

## Installation

Make sure you have Go installed (>= 1.20), then build the tool:

```bash
git clone https://github.com/adamfitz/mdu
cd mdu
go build -o mdu
```

Or run directly:

```bash
go run main.go [command] [flags]
```

---

## Usage

```bash
mdu [command] [flags]
```

---

## Commands

### `read`

Read metadata from an EPUB file.

```bash
mdu read --file <EPUB_FILE> [flags]
```

#### Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--file` | - | Target EPUB file (required) |
| `--all` | `-a` | Return all metadata fields, not just the known ones |
| `--list-fields` | - | List all supported metadata fields |
| `--output` | `-o` | Write the formatted metadata output to a file (also prints to stdout) |

#### Examples

Read known metadata fields and print to console:

```bash
mdu read --file ./books/example.epub
```

Read all metadata fields and write to a file:

```bash
mdu read --file ./books/example.epub --all --output metadata.txt
```

List supported metadata fields:

```bash
mdu read --list-fields
```

---

### `update`

Update metadata fields in an EPUB file. The file is updated **in place**.

```bash
mdu update --file <EPUB_FILE> [flags]
```

#### Flags

| Flag | Description |
|------|-------------|
| `--file` | Target EPUB file (required) |
| `--series` | Series name |
| `--series-index` | Series index |
| `--summary` | Book summary |
| `--isbn` | ISBN identifier |
| `--author` | Author/creator name |

#### Examples

Update author and series index:

```bash
mdu update --file ./books/example.epub --author "John Doe" --series-index "2"
```

Update multiple fields at once:

```bash
mdu update --file ./books/example.epub --author "Jane Doe" --series "My Saga" --summary "A great book" --isbn "9781234567890"
```

---

## Metadata Fields

The tool supports the following metadata fields by default:

- `title` – Book title  
- `author` – Author / creator  
- `summary` – Book summary  
- `isbn` – ISBN identifier  
- `calibre:series` – Series name  
- `calibre:series_index` – Series index  

You can also return **all metadata fields** using the `--all` flag with the `read` command.

---

## Output

Metadata is printed as a formatted **two-column table**, left-aligned:

```
Metadata field      Metadata value
---------------     --------------
author              Jane Doe
calibre:series      My Saga
isbn                9781234567890
summary             This is a sample book
title               Sample Book
```

If `--output` is provided, the same table is written to the specified file **and printed to stdout**.

---

## Notes

- The `update` command updates the EPUB file **in place**. Always keep a backup if needed.  
- Flags not provided in the `update` command will leave existing metadata unchanged.  
- The tool works with standard EPUB 2 and EPUB 3 files.

---

## License

MIT License
