# mdu - Metadata Updater

A command-line tool for reading and updating EPUB metadata while preserving the original OPF file structure and all XML namespaces.

Perfect for organizing your ebook library with media management tools like [Kavita](https://www.kavitareader.com/).

## Features

- 📖 **Read metadata** from single files or entire directories
- ✏️ **Update metadata** with command-line flags or input files (JSON/YAML)
- 🔍 **Validate changes** by comparing original and modified files
- 🛡️ **Preserves structure** - keeps all existing metadata and XML namespaces intact
- 📦 **Batch operations** - process multiple files at once
- 💾 **Automatic backups** - creates `.backup` files by default
- ✅ **EPUB validation** - checks file structure and required metadata

## Installation

```bash
# Clone the repository
git clone https://github.com/adamfitz/mdu.git
cd mdu

# Install dependencies
go get gopkg.in/yaml.v3

# Build
go build -o mdu

# Or run directly
go run . [command]
```

## Quick Start

```bash
# Read metadata from an EPUB file
mdu read --file book.epub

# Update metadata with command-line flags
mdu update --file book.epub --author "Author Name" --series "Series Name" --series-index "1"

# Generate a metadata template
mdu generate -o metadata.yaml

# Update using an input file
mdu update --file book.epub --input metadata.yaml

# Validate changes
mdu validate --file book.epub

# Check EPUB validity
mdu check --dir ./books
```

## Commands

### `read` - Read Metadata

Read and display metadata from EPUB files.

```bash
# Read single file
mdu read --file book.epub

# Read all metadata fields (not just supported ones)
mdu read --file book.epub --all

# Read all EPUBs in a directory
mdu read --dir ./books

# Save output to file
mdu read --dir ./books -o metadata.txt

# List supported metadata fields
mdu read --list-fields
```

**Supported Fields:**
- `author` - Book author/creator
- `title` - Book title
- `summary` - Book description
- `isbn` - ISBN identifier
- `calibre:series` - Series name
- `calibre:series_index` - Series position

### `update` - Update Metadata

Update metadata in EPUB files while preserving the original OPF structure.

**Using command-line flags:**
```bash
mdu update --file book.epub \
  --author "Author Name" \
  --series "Series Name" \
  --series-index "1" \
  --summary "Book description" \
  --isbn "978-1234567890"

# Batch update directory
mdu update --dir ./books --series "The Expanse"

# Update without creating backups
mdu update --file book.epub --author "New Author" --backup=false
```

**Using input files (recommended for batch operations):**
```bash
# Generate a template
mdu generate -o metadata.yaml

# Edit the template with your metadata
# Then apply it
mdu update --file book.epub --input metadata.yaml

# Apply to multiple files
mdu update --dir ./books --input shared-metadata.yaml
```

**Flags:**
- `--file` - Target EPUB file
- `--dir` - Target directory (batch operation)
- `--input/-i` - Input file with metadata (JSON or YAML)
- `--author` - Set author/creator
- `--summary` - Set book description
- `--isbn` - Set ISBN identifier
- `--series` - Set series name (for Kavita)
- `--series-index` - Set series position (for Kavita)
- `--backup` - Create backup files (default: true)
- `--output/-o` - Write results to file

**Note:** Input files take precedence over command-line flags.

### `generate` - Generate Template Files

Create example input files for batch metadata updates.

```bash
# Generate JSON template
mdu generate -o template.json

# Generate YAML template
mdu generate -o template.yaml
```

**JSON Format:**
```json
{
  "author": "Author Name",
  "summary": "Book description or summary",
  "isbn": "978-1234567890",
  "series": "Series Name",
  "series_index": "1",
  "additional": {
    "publisher": "Publisher Name",
    "language": "en-US"
  }
}
```

**YAML Format:**
```yaml
# EPUB Metadata Update Template
author: "Author Name"
summary: "Book description or summary"
isbn: "978-1234567890"
series: "Series Name"
series_index: "1"

additional:
  publisher: "Publisher Name"
  language: "en-US"
```

### `validate` - Validate Changes

Compare an EPUB file with its backup to validate changes.

```bash
# After updating with --backup (default)
mdu update --file book.epub --author "New Author"
mdu validate --file book.epub

# Save validation report
mdu validate --file book.epub -o validation.txt
```

**Output Example:**
```
=== METADATA COMPARISON ===

~ CHANGED: author
  Old: Original Author
  New: New Author

=== SUMMARY ===
Unchanged: 15
Changed:   1
Added:     0
Removed:   0
```

### `compare` - Compare Two Files

Compare metadata between two EPUB files.

```bash
mdu compare original.epub modified.epub

# Save comparison report
mdu compare original.epub modified.epub -o diff.txt
```

### `check` - Validate EPUB Structure

Check if EPUB files have valid structure and required metadata.

```bash
# Check single file
mdu check --file book.epub

# Check entire directory
mdu check --dir ./books

# Save check results
mdu check --dir ./books -o validation-report.txt
```

**Validates:**
- ✓ `META-INF/container.xml` exists
- ✓ OPF file exists at location specified in container.xml
- ✓ OPF file is readable
- ✓ Required metadata present (title, identifier, language)

## Using Input Files

Input files are ideal for batch operations and complex metadata management.

### Basic Workflow

```bash
# 1. Generate a template
mdu generate -o metadata.yaml

# 2. Edit the template with your metadata
# (use your favorite text editor)

# 3. Apply to your files
mdu update --file book.epub --input metadata.yaml

# 4. Validate changes
mdu validate --file book.epub
```

### Use Case: Shared Series Metadata

Create one file with common series information, customize per book:

**shared-series.yaml:**
```yaml
author: "James S.A. Corey"
series: "The Expanse"
additional:
  publisher: "Orbit Books"
  language: "en"
```

**Apply to each book:**
```bash
mdu update --file "Leviathan Wakes.epub" \
  --input shared-series.yaml \
  --series-index "1" \
  --summary "Humanity has colonized the solar system..."

mdu update --file "Caliban's War.epub" \
  --input shared-series.yaml \
  --series-index "2" \
  --summary "On Ganymede, breadbasket of the outer planets..."
```

### Use Case: Complete Per-Book Metadata

Create individual metadata files for each book:

**vol1-metadata.json:**
```json
{
  "author": "Shimesaba",
  "summary": "Volume 1 description",
  "series": "Higehiro",
  "series_index": "1",
  "isbn": "9781975344207"
}
```

**Apply:**
```bash
mdu update --file vol1.epub --input vol1-metadata.json
mdu update --file vol2.epub --input vol2-metadata.json
mdu update --file vol3.epub --input vol3-metadata.json
```

## Kavita Integration

Kavita uses specific metadata fields to organize your library:

| Field | Kavita Uses For | How to Set |
|-------|----------------|------------|
| `calibre:series` | Series grouping | `--series` flag or `series` in input file |
| `calibre:series_index` | Series order | `--series-index` flag or `series_index` in input file |
| `dc:creator` | Author | `--author` flag or `author` in input file |
| `dc:description` | Book description | `--summary` flag or `summary` in input file |
| `dc:title` | Book title | Read-only (use calibre or other tools to change) |

### Complete Kavita Workflow

```bash
# 1. Check your EPUBs are valid
mdu check --dir ./my-books

# 2. Create metadata template for the series
mdu generate -o expanse-metadata.yaml

# 3. Edit with your series info
cat > expanse-metadata.yaml << EOF
author: "James S.A. Corey"
series: "The Expanse"
additional:
  publisher: "Orbit Books"
  language: "en"
EOF

# 4. Apply to each book with unique series_index and summary
mdu update \
  --file "Leviathan Wakes.epub" \
  --input expanse-metadata.yaml \
  --series-index "1" \
  --summary "Book 1 summary..."

mdu update \
  --file "Caliban's War.epub" \
  --input expanse-metadata.yaml \
  --series-index "2" \
  --summary "Book 2 summary..."

# 5. Validate all changes
mdu check --dir . -o validation.txt

# 6. Import into Kavita - series will be properly organized!
```

## How It Works

### EPUB Structure

EPUB files are ZIP archives with a specific structure:

```
book.epub
├── META-INF/
│   └── container.xml          # Points to OPF file location
├── OEBPS/                     # Common content directory
│   ├── package.opf            # Metadata, manifest, spine
│   ├── chapter1.xhtml
│   └── ...
└── mimetype
```

### Metadata Location

1. `META-INF/container.xml` specifies the OPF file location:
```xml
<container>
  <rootfiles>
    <rootfile full-path="OEBPS/package.opf" .../>
  </rootfiles>
</container>
```

2. The OPF file (e.g., `OEBPS/package.opf`) contains all metadata:
```xml
<package>
  <metadata>
    <dc:title>Book Title</dc:title>
    <dc:creator>Author Name</dc:creator>
    <dc:identifier opf:scheme="ISBN">978-1234567890</dc:identifier>
    <meta name="calibre:series" content="Series Name"/>
    <meta name="calibre:series_index" content="1"/>
  </metadata>
  <manifest>...</manifest>
  <spine>...</spine>
</package>
```

### Preservation Strategy

**Why not standard XML parsing?**

Traditional XML unmarshaling/marshaling can lose:
- XML namespace declarations (`xmlns:dc`, `xmlns:opf`, `xmlns:calibre`)
- Formatting and whitespace
- Comment nodes
- Attribute order
- Unknown elements

**mdu's approach:**

1. Reads the entire OPF file as text
2. Uses targeted string manipulation to update specific elements
3. Writes the modified content back, preserving everything else
4. Only the fields you explicitly update are changed

This ensures:
- ✅ All XML namespaces preserved
- ✅ Original formatting maintained
- ✅ All attributes kept intact
- ✅ Unknown elements remain untouched
- ✅ EPUB stays valid

## EPUB Specification

### Required Metadata (EPUB 3 Spec)

Every valid EPUB must have:
- **`dc:title`** - Book title
- **`dc:identifier`** - Unique identifier (ISBN, UUID, etc.)
- **`dc:language`** - Language code (e.g., "en-US")

mdu validates these requirements and will warn if they're missing.

### Common OPF Locations

While the EPUB spec doesn't mandate a specific location, common conventions include:
- `OEBPS/content.opf`
- `OEBPS/package.opf`
- `content.opf` (root level)
- `OPS/package.opf`

The actual location is specified in `META-INF/container.xml` and mdu automatically finds it.

## Error Handling

### Invalid EPUB
```
Error: failed to locate OPF file: container.xml not found in META-INF
```
**Fix:** File is not a valid EPUB. Re-download or use a different file.

### Missing Required Metadata
```
Warning: missing required EPUB metadata fields: dc:title, dc:language
```
**Fix:** Add missing fields manually with an EPUB editor, or the EPUB may have issues with some readers.

### OPF File Not Found
```
Error: failed to read OPF file at 'OEBPS/content.opf': file not found
```
**Fix:** EPUB structure is broken - `container.xml` points to non-existent OPF file.

## Best Practices

1. **Always keep backups** - The tool creates `.backup` files by default, but keep original files elsewhere too
2. **Test on one file first** - Before batch operations, test on a single file and validate
3. **Use `--all` flag** - See all available metadata fields before updating
4. **Run `check` before updating** - Validate EPUB structure before making changes
5. **Use `validate` after updates** - Verify changes are correct
6. **Use input files for series** - Easier to manage consistent metadata across multiple books
7. **Version control your metadata** - Keep input files in git for reproducibility

## Examples

### Example 1: Simple Update
```bash
# Update a single book
mdu update --file book.epub \
  --author "Brandon Sanderson" \
  --series "Mistborn" \
  --series-index "1"

# Validate the changes
mdu validate --file book.epub
```

### Example 2: Batch Series Update
```bash
# Create metadata template
cat > mistborn.yaml << EOF
author: "Brandon Sanderson"
series: "Mistborn"
additional:
  publisher: "Tor Books"
  language: "en"
EOF

# Update each book
mdu update --file "The Final Empire.epub" --input mistborn.yaml --series-index "1"
mdu update --file "The Well of Ascension.epub" --input mistborn.yaml --series-index "2"
mdu update --file "The Hero of Ages.epub" --input mistborn.yaml --series-index "3"

# Check all files
mdu check --dir . -o validation.txt
```

### Example 3: Fix Missing Metadata
```bash
# Check what's missing
mdu check --dir ./books

# Create fix template
cat > fix-metadata.yaml << EOF
additional:
  language: "en-US"
  publisher: "Publisher Name"
EOF

# Apply fixes
mdu update --dir ./books --input fix-metadata.yaml

# Verify fixes
mdu check --dir ./books -o after-fix.txt
```

### Example 4: Light Novel Series
```bash
# Generate template
mdu generate -o higehiro.yaml

# Edit template
cat > higehiro.yaml << EOF
author: "Shimesaba"
series: "Higehiro: After Being Rejected, I Shaved and Took in a High School Runaway"
additional:
  publisher: "Yen Press"
  language: "en-US"
EOF

# Update each volume
for i in {1..5}; do
  mdu update \
    --file "Higehiro v$(printf '%02d' $i).epub" \
    --input higehiro.yaml \
    --series-index "$i" \
    --summary "Volume $i description"
done

# Validate all
mdu check --dir .
```

## Technical Details

### Dependencies

- Go 1.16 or higher
- `github.com/spf13/cobra` - CLI framework
- `gopkg.in/yaml.v3` - YAML parsing

### Supported File Formats

- **Input files:** `.json`, `.yaml`, `.yml`
- **EPUB files:** Standard EPUB 2 and EPUB 3 formats

### Thread Safety

- Safe for concurrent reads (multiple `read` commands)
- Updates are atomic (create new file, then replace)
- No file locking (don't run multiple updates on same file simultaneously)

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

[Add your license here]

## Acknowledgments

Built for use with [Kavita](https://www.kavitareader.com/) media management.

## Support

If you encounter any issues or have questions:
1. Check the [error handling](#error-handling) section
2. Run `mdu check` on your EPUB files to validate structure
3. Open an issue on GitHub with the error message and EPUB details

---

**Note:** This tool modifies EPUB files. Always keep backups of your original files!