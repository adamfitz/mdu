# mdu - Metadata Updater

A command-line tool for managing metadata in EPUB and CBZ files with integrated MangaDex support.

Perfect for organizing your ebook and manga library with media management tools like [Kavita](https://www.kavitareader.com/), [Komga](https://komga.org/), and other comic/manga readers.

## Features

### EPUB Support
* 📖 **Read metadata** from single files or entire directories
* ✏️ **Update metadata** with command-line flags or input files (JSON/YAML)
* 🔍 **Validate changes** by comparing original and modified files
* 🛡️ **Preserves structure** - keeps all existing metadata and XML namespaces intact
* 📦 **Batch operations** - process multiple files at once
* 💾 **Automatic backups** - creates `.backup` files by default
* ✅ **EPUB validation** - checks file structure and required metadata

### CBZ/ComicInfo.xml Support
* 📚 **MangaDex Integration** - Fetch manga metadata directly from MangaDex API
* 🔍 **Smart Search** - Find manga titles with fuzzy matching
* 🏷️ **ComicInfo.xml** - Create, update, and validate ComicInfo.xml metadata files
* 🔄 **Update Existing Files** - Add or update ComicInfo.xml in existing CBZ archives
* ✓ **Metadata Validation** - Verify ComicInfo.xml structure and content
* 🎯 **Smart Chapter Extraction** - Automatic chapter number detection from filenames
* 🛡️ **Integrity Validation** - Internal SHA256 verification during repackaging (with automatic retry)

## Installation

### Download Pre-built Binaries (Recommended)

Download the latest release for your platform from the [releases page](https://github.com/adamfitz/mdu/releases):

#### Linux (.deb package)
```bash
# Download the .deb file
wget https://github.com/adamfitz/mdu/releases/latest/download/mdu_amd64.deb

# Install
sudo dpkg -i mdu_amd64.deb

# Run from anywhere
mdu --version
```

#### Linux (.rpm package)
```bash
# Download the .rpm file
wget https://github.com/adamfitz/mdu/releases/latest/download/mdu_x86_64.rpm

# Install on Fedora/RHEL/CentOS
sudo dnf install mdu_x86_64.rpm

# Or on older systems
sudo rpm -i mdu_x86_64.rpm

# Run from anywhere
mdu --version
```

### Build from Source

```bash
# Clone the repository
git clone https://github.com/adamfitz/mdu.git
cd mdu

# Install dependencies
go mod download

# Build
go build -o mdu

# Or install to GOPATH
go install
```

## Quick Start

### EPUB Files
```bash
# Read metadata from an EPUB file
mdu epub read --file book.epub

# Update metadata
mdu epub update --file book.epub --author "Author Name" --series "Series Name" --series-index "1"
```

### CBZ Files
```bash
# Search for a manga on MangaDex
mdu comicinfo search "One Piece"

# Generate ComicInfo.xml from MangaDex for your CBZ files
mdu comicinfo generate --mangadex-id "a1c7c817-4e59-43b7-9365-09675a149a6f" --dir ./manga

# Update existing CBZ with metadata
mdu comicinfo update --file manga.cbz --series "One Piece" --number "1"

# Verify ComicInfo.xml in a CBZ file
mdu comicinfo validate manga.cbz
```

## Commands

### EPUB Commands

#### `epub read` - Read EPUB Metadata

```bash
# Read single file
mdu epub read --file book.epub

# Read all metadata fields
mdu epub read --file book.epub --all

# Read all EPUBs in a directory
mdu epub read --dir ./books

# Save output to file
mdu epub read --dir ./books -o metadata.txt

# List supported metadata fields
mdu epub read --list-fields
```

**Supported Fields:**
- `author` - Book author/creator
- `summary` - Book description
- `isbn` - ISBN identifier
- `calibre:series` - Series name
- `calibre:series_index` - Series position

#### `epub update` - Update EPUB Metadata

```bash
# Update using command-line flags
mdu epub update --file book.epub \
  --author "Author Name" \
  --series "Series Name" \
  --series-index "1" \
  --summary "Book description"

# Batch update directory
mdu epub update --dir ./books --series "The Expanse"

# Update without creating backups
mdu epub update --file book.epub --author "New Author" --backup=false

# Using input files
mdu epub update --file book.epub --input metadata.yaml
```

#### `epub validate` - Validate EPUB Changes

```bash
# Validate changes (compares with .backup file)
mdu epub validate --file book.epub

# Save validation report
mdu epub validate --file book.epub -o validation.txt
```

#### `epub check` - Check EPUB Structure

```bash
# Check single file
mdu epub check --file book.epub

# Check entire directory
mdu epub check --dir ./books

# Save check results
mdu epub check --dir ./books -o validation-report.txt
```

#### `epub compare` - Compare Two EPUB Files

```bash
# Compare two EPUB files
mdu epub compare original.epub modified.epub

# Save comparison report
mdu epub compare original.epub modified.epub -o diff.txt
```

---

### ComicInfo Commands

#### `comicinfo search` - Search MangaDex

Search MangaDex for a manga title using fuzzy matching to find the best match.

```bash
# Search for a manga (returns MangaDex ID)
mdu comicinfo search "One Piece"

# Multi-word searches work without quotes
mdu comicinfo search Attack on Titan

# Search for exact matches
mdu comicinfo search "Berserk"
```

**Output Example:**
```
$ go run . comicinfo search  One Piece

🔍 Searching MangaDex for: One Piece

Best Match (Score: 1.00):
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Name                                     Alt Name                                 Mangadex ID                             
---------------------------------------- ---------------------------------------- ----------------------------------------
One Piece Party                          ワンピースパーティー                     a0b49136-9f4d-46d4-b5dd-d5393e015009    

Other Top Matches:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Name                                     Alt Name                                 Mangadex ID                             
---------------------------------------- ---------------------------------------- ----------------------------------------
One Piece (Official Colored)             Ван Пис                                  a2c1d849-af05-4bbc-b2a7-866ebb10331f    
ONE PIECE STRONG WORLD                   One Piece: Strong World                  fcb5111f-be88-4f5a-b456-22135f3eda49    
Dragon Ball X One Piece - Cross Epoch    Época Cruzada                            6c6236b0-0fd7-4926-8a1a-cc346a6627e3    
One Piece: Ace’s Story—The Manga         One Piece Episode Ace                    470ac1ec-c3dd-4dc2-9e5d-0104600ed1a5    
ONE PIECE                                ワンピース                               a1c7c817-4e59-43b7-9365-09675a149a6f    
One Piece Academy                        ONE PIECE学園                            b70113a5-32a3-44e8-a28f-0e88392808ba    
One Piece (Fan Colored)                  -                                        c2e76d62-702b-4a8e-a4d0-c7cecd45b8ea    
One Piece Romance Dawn V.1               ROMANCE DAWN - ロマンス ドーン           fc4eccf5-5fe6-49de-b12c-4124fcea4e16    
One Piece - Vivi no Bouken               ビビの冒険                               7bd37661-c0b1-4166-87c1-7e633adb55ac    

💡 Use this ID with: mdu comicinfo generate --mangadex-id a0b49136-9f4d-46d4-b5dd-d5393e015009 --dir <path>
```

**How Search Works:**
- Uses token-based fuzzy matching with Levenshtein distance
- Searches English titles only
- Combines Jaccard similarity (70%) and edit distance (30%) for scoring
- Returns the best match with confidence score

#### `comicinfo generate` - Generate ComicInfo.xml from MangaDex

Create ComicInfo.xml files in your CBZ archives by fetching metadata from MangaDex.

```bash
# Generate for a single file
mdu comicinfo generate --mangadex-id "abc123-def456" chapter01.cbz

# Process entire directory
mdu comicinfo generate --mangadex-id "abc123-def456" --dir ./chapters

# Positional argument also works
mdu comicinfo generate --mangadex-id "abc123" ./manga/
```

**What This Command Does:**
1. Fetches manga metadata from MangaDex (series title, summary, genres, etc.)
2. Fetches author and artist names from MangaDex API
3. Extracts chapter numbers from CBZ filenames
4. Creates ComicInfo.xml for each chapter with appropriate metadata
5. Repackages CBZ files with the new ComicInfo.xml
6. Validates integrity with SHA256 checksums (internal, automatic)
7. Uses retry logic (up to 3 attempts) if repackaging fails
8. Replaces original files only after successful validation

**Chapter Number Extraction:**

The command automatically extracts chapter numbers from filenames using these patterns:
- `ch001`, `ch-01`, `ch_1`, `ch 1`
- `chapter 1`, `chapter-01`, `chapter_001`
- `c1`, `c-01`, `c_001`
- Files like `One Piece - ch100.cbz` → extracts "100"
- Files like `series_name_chapter_025.cbz` → extracts "25"

**Example Output:**
```
🔍 Fetching metadata from MangaDex (ID: abc123)...
✓ Metadata fetched successfully
✓ Metadata converted to ComicInfo format

📚 Processing 5 file(s)...

[1/5] Processing: One Piece ch001.cbz
  📄 Extracted chapter number: 1
  📁 Created temp directory: /tmp/mdu_One_Piece_ch001_1234567890
  📦 Extracting CBZ contents...
  📝 Writing ComicInfo.xml...
  🗜️  Creating new CBZ file...
  🔒 Validating CBZ integrity...
  ✓ Integrity check passed
  🔄 Replacing original file...
✓ Successfully processed: One Piece ch001.cbz

[2/5] Processing: One Piece ch002.cbz
  ...

────────────────────────────────────────
✓ Successfully processed: 5
Total files: 5

File - One Piece ch001.cbz:

  Series: One Piece
  Number: 1
  Title: Chapter 1
  Summary: Monkey D. Luffy sets sail...
  Writer: Eiichiro Oda
  Penciller: Eiichiro Oda
  Genre: Action, Adventure, Fantasy
  ...
```

#### `comicinfo read` - Read ComicInfo.xml

```bash
# Read single file
mdu comicinfo read comic.cbz

# Read with positional argument
mdu comicinfo read comic.cbz

# Read all files in directory
mdu comicinfo read --dir ./comics

# Save to file
mdu comicinfo read --dir ./comics -o metadata.txt

# List supported fields
mdu comicinfo read --list-fields
```

#### `comicinfo update` - Update ComicInfo.xml

Update existing ComicInfo.xml metadata in CBZ files.

```bash
# Update specific fields
mdu comicinfo update comic.cbz --series "One Piece" --number "100"

# Update with flags
mdu comicinfo update --file manga.cbz --writer "Author Name" --publisher "Publisher"

# Update multiple files
mdu comicinfo update --dir ./manga --series "My Series"

# Update from input file
mdu comicinfo update --file manga.cbz --input metadata.yaml

# Positional argument works
mdu comicinfo update ./manga/ --series "New Series"
```

**Supported Flags:**
- `--series` - Series name
- `--number` - Issue/chapter number
- `--volume` - Volume number
- `--summary` - Issue summary/description
- `--writer` - Writer name
- `--publisher` - Publisher name

**Input File Support:**

You can use JSON or YAML files for batch updates:

```yaml
# metadata.yaml
author: "Eiichiro Oda"
series: "One Piece"
summary: "Chapter description"
additional:
  publisher: "Shueisha"
```

```bash
mdu comicinfo update --dir ./chapters --input metadata.yaml
```

#### `comicinfo validate` - Validate CBZ Structure

```bash
# Validate single file
mdu comicinfo validate comic.cbz

# With positional argument
mdu comicinfo validate comic.cbz

# Save validation report
mdu comicinfo validate --file comic.cbz -o validation.txt
```

**Validation Checks:**
- CBZ file can be opened as ZIP archive
- ComicInfo.xml exists in archive
- ComicInfo.xml has valid XML structure
- All files in archive are readable

#### `comicinfo check` - Check CBZ Files

Check multiple CBZ files for validity and ComicInfo.xml presence.

```bash
# Check single file
mdu comicinfo check comic.cbz

# Check directory
mdu comicinfo check --dir ./comics

# Save report
mdu comicinfo check --dir ./comics -o report.txt
```

**Output Example:**
```
=== Checking: manga_ch001.cbz ===
✓ CBZ structure valid
✓ ComicInfo.xml found and readable
  Series: One Piece
  Number: 1
  Writer: Eiichiro Oda

=== Checking: manga_ch002.cbz ===
⚠  CBZ structure valid but ComicInfo.xml not found

=== Summary ===
Valid (with ComicInfo.xml): 1
Valid (missing ComicInfo.xml): 1
Invalid: 0
```

#### `comicinfo compare` - Compare Two CBZ Files

```bash
# Compare metadata between files
mdu comicinfo compare original.cbz modified.cbz

# Save comparison
mdu comicinfo compare original.cbz modified.cbz -o diff.txt
```

---

## ComicInfo.xml Format

ComicInfo.xml is the standard metadata format for comic book archives, supported by many readers including Kavita, Komga, and others.

### Standard Fields

| Field | Type | Description | Example |
|-------|------|-------------|---------|
| `Series` | String | Series name | "One Piece" |
| `Number` | String | Issue/chapter number | "1", "100.5" |
| `Volume` | String | Volume number | "1" |
| `Title` | String | Chapter/issue title | "Romance Dawn" |
| `Summary` | String | Chapter description | "The adventure begins..." |
| `Writer` | String | Author(s) | "Eiichiro Oda" |
| `Penciller` | String | Artist(s) | "Eiichiro Oda" |
| `Publisher` | String | Publisher name | "Shueisha" |
| `Genre` | String | Genre (comma-separated) | "Action, Adventure, Fantasy" |
| `Tags` | String | Tags (comma-separated) | "Shounen, Pirates" |
| `LanguageISO` | String | Language code | "en", "ja" |
| `Manga` | String | Reading direction | "Yes", "YesAndRightToLeft" |
| `PageCount` | Integer | Number of pages | 20 |
| `Year` | Integer | Publication year | 1997 |
| `Month` | Integer | Publication month (1-12) | 7 |
| `Day` | Integer | Publication day (1-31) | 22 |
| `Web` | String | Source URL | "https://mangadex.org/title/..." |
| `AgeRating` | String | Content rating | "Everyone", "Teen", "Mature 17+" |
| `BlackAndWhite` | String | Color information | "Yes", "No" |
| `Notes` | String | Additional notes | "Status: ongoing" |

### Example ComicInfo.xml

```xml
<?xml version="1.0"?>
<ComicInfo xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" 
           xmlns:xsd="http://www.w3.org/2001/XMLSchema">
  <Series>One Piece</Series>
  <Number>1</Number>
  <Volume>1</Volume>
  <Title>Chapter 1</Title>
  <Summary>Monkey D. Luffy sets sail to become the Pirate King...</Summary>
  <Writer>Eiichiro Oda</Writer>
  <Penciller>Eiichiro Oda</Penciller>
  <Publisher>Shueisha</Publisher>
  <Genre>Action, Adventure, Fantasy</Genre>
  <Tags>Shounen, Pirates</Tags>
  <Manga>Yes</Manga>
  <LanguageISO>ja</LanguageISO>
  <PageCount>20</PageCount>
  <Year>1997</Year>
  <Month>7</Month>
  <Web>https://mangadex.org/title/a1c7c817-4e59-43b7-9365-09675a149a6f</Web>
  <Notes>Status: ongoing</Notes>
  <AgeRating>Teen</AgeRating>
  <BlackAndWhite>Yes</BlackAndWhite>
</ComicInfo>
```

## MangaDex Integration

mdu integrates directly with the MangaDex API to fetch accurate manga metadata.

### Finding MangaDex IDs

**Method 1: Use the search command (Recommended)**
```bash
# Search by title
mdu comicinfo search "Attack on Titan"

# Output shows the MangaDex ID
💡 Use this ID with: mdu comicinfo generate --mangadex-id 304ceac3-8cdb-4fe7-acf7-2b6ff7a60613 --dir <path>
```

**Method 2: From Browser**
- Visit the manga on MangaDex.org
- Copy the ID from the URL: `https://mangadex.org/title/{manga-id}/manga-name`

### Fetched Metadata

When you provide a MangaDex ID, mdu automatically retrieves:

- **Series Information**: Title (English preferred), description
- **Chapter Details**: Volume/chapter numbers, title
- **Creator Credits**: Authors and artists (fetched from API by ID)
- **Publication Info**: Release dates, publisher
- **Language & Region**: Original language
- **Reading Direction**: Set to "Yes" for manga
- **Tags & Genres**: Categories separated into Genre and Tags fields
- **Content Rating**: Mapped to appropriate age ratings
- **Status**: Ongoing, completed, etc.

### Complete Workflow Example

```bash
# 1. Search for the manga
mdu comicinfo search "One Piece"
# Output: MangaDex ID: a1c7c817-4e59-43b7-9365-09675a149a6f

# 2. Generate ComicInfo.xml for all chapters
mdu comicinfo generate \
  --mangadex-id "a1c7c817-4e59-43b7-9365-09675a149a6f" \
  --dir ./One_Piece_Chapters/

# 3. Verify the results
mdu comicinfo check --dir ./One_Piece_Chapters/ -o verification.txt
```

## Input Files for Batch Operations

For complex metadata or batch operations, use JSON or YAML input files.

### Generate Template

```bash
# Generate template for EPUB
mdu generate --output epub-template.yaml --format epub

# Generate template for ComicInfo
mdu generate --output comic-template.json --format comic
```

### JSON Format

```json
{
  "author": "Eiichiro Oda",
  "summary": "Chapter description",
  "series": "One Piece",
  "series_index": "1",
  "additional": {
    "publisher": "Shueisha",
    "language": "en"
  }
}
```

### YAML Format

```yaml
# Metadata Template
author: "Eiichiro Oda"
summary: "Chapter description"
series: "One Piece"
series_index: "1"

additional:
  publisher: "Shueisha"
  language: "en"
```

### Using Input Files

```bash
# Create shared series metadata
cat > series-metadata.yaml << EOF
author: "Kohei Horikoshi"
series: "My Hero Academia"
additional:
  publisher: "Shueisha"
  language: "en"
EOF

# Apply to multiple files
mdu comicinfo update --file "MHA Ch001.cbz" --input series-metadata.yaml --number "1"
mdu comicinfo update --file "MHA Ch002.cbz" --input series-metadata.yaml --number "2"
mdu comicinfo update --file "MHA Ch003.cbz" --input series-metadata.yaml --number "3"
```

## Chapter Number Detection

mdu automatically extracts chapter numbers from filenames using these patterns:

### Supported Patterns

```
ch001, ch-01, ch_1, ch 1           → extracts: 1, 1, 1, 1
chapter 1, chapter-01, chapter_001  → extracts: 1, 1, 1
c1, c-01, c_001                     → extracts: 1, 1, 1
something-ch01                      → extracts: 1
series_name_chapter_025             → extracts: 25
```

### Pattern Matching

The patterns are case-insensitive and work with:
- Leading zeros are removed (ch001 → 1, ch100 → 100)
- Various separators: `-`, `_`, space
- Prefixes: `ch`, `chapter`, `c`
- Embedded chapter indicators in longer filenames

### Examples

```bash
# These filenames will have chapters auto-detected:
One Piece - ch001.cbz          → Chapter: 1
one_piece_chapter_100.cbz      → Chapter: 100
[Group] Series c025.cbz        → Chapter: 25
attack-on-titan-ch-050.cbz     → Chapter: 50

# Files without detectable patterns will be skipped
random_file.cbz                → ⚠️  Skipped (no chapter number found)
```

## Integrity Validation

mdu uses SHA256 checksums internally to ensure CBZ files are correctly repackaged.

### How It Works

1. **During Repackaging**:
   - Extracts CBZ to temporary directory
   - Adds/updates ComicInfo.xml
   - Creates new CBZ file
   - Validates all files in new CBZ are readable
   - Verifies SHA256 checksums of all extracted files match archived files
   - Replaces original only if validation passes

2. **Automatic Retry Logic**:
   - If validation fails, retries up to 3 times
   - Cleans up failed attempts automatically
   - Only replaces original after successful validation

3. **Safety Features**:
   - Temporary directories used (cleaned up automatically)
   - Original preserved until new file validates
   - ZipSlip vulnerability protection
   - All files verified readable before replacement

**Note:** This is an internal integrity check, not a user-facing checksum feature for tracking file modifications over time.

## Media Server Integration

### Kavita

Kavita uses ComicInfo.xml to organize manga and comics.

**Key Fields for Kavita:**
- `Series`: Groups files into series
- `Number`: Orders chapters/issues
- `Volume`: Groups chapters into volumes
- `LocalizedSeries`: Display name (alternative series name)

```bash
# Prepare manga for Kavita
mdu comicinfo generate \
  --mangadex-id "abc123" \
  --dir ./manga

# Verify before importing
mdu comicinfo check --dir ./manga -o validation.txt
```

### Komga

Komga also reads ComicInfo.xml with similar fields.

**Additional Komga Features:**
- Reads `Publisher` for filtering
- Uses `Genre` and `Tags`
- Displays `Summary` in reader
- Supports `Web` links

```bash
# Add comprehensive metadata for Komga
mdu comicinfo update --file manga.cbz \
  --series "One Piece" \
  --publisher "Shueisha" \
  --summary "Full description..."
```

### Other Supported Readers

ComicInfo.xml is supported by:
- **Panels** (iOS)
- **Perfect Viewer** (Android)
- **CDisplayEx** (Windows)
- **Chunky** (iOS)
- **Tachiyomi** (Android)
- **YACReader** (Cross-platform)

## Best Practices

### For EPUB Files
1. **Always keep backups** - Use default backup behavior or keep originals elsewhere
2. **Test on one file first** - Before batch operations
3. **Use `check` before updating** - Validate structure first
4. **Use input files for series** - Easier to manage consistent metadata

### For CBZ Files
1. **Use MangaDex search** - Find correct manga IDs with `comicinfo search`
2. **Organize files first** - Use consistent chapter naming for auto-detection
3. **Verify after generation** - Run `comicinfo check` to ensure success
4. **Test with one file** - Always test new operations on a single file first
5. **Trust the retry logic** - If generation fails, it will automatically retry up to 3 times

### File Organization

```
manga-library/
├── One Piece/
│   ├── One Piece - ch001.cbz  ← Include series name and chapter
│   ├── One Piece - ch002.cbz
│   └── One Piece - ch003.cbz
├── Attack on Titan/
│   ├── aot_chapter_001.cbz    ← Various formats work
│   └── aot_chapter_002.cbz
└── metadata/
    ├── one-piece-metadata.yaml  ← Store templates
    └── aot-metadata.yaml
```

## Examples

### Example 1: Complete MangaDex Workflow

```bash
# Step 1: Search for manga
mdu comicinfo search "Chainsaw Man"
# Output shows: ID = a1c7c817-4e59-43b7-9365-09675a149a6f

# Step 2: Generate metadata for all chapters
mdu comicinfo generate \
  --mangadex-id "a1c7c817-4e59-43b7-9365-09675a149a6f" \
  --dir ~/manga/chainsaw_man/

# Step 3: Verify results
mdu comicinfo check --dir ~/manga/chainsaw_man/ -o check.txt

# Step 4: Import into Kavita or Komga
```

### Example 2: Update Existing Collection

```bash
# Check current state
mdu comicinfo read --dir ./manga -o before.txt

# Create shared metadata
cat > metadata.yaml << EOF
series: "Bleach"
author: "Tite Kubo"
additional:
  publisher: "Shueisha"
  language: "en"
EOF

# Update all files
mdu comicinfo update --dir ./manga --input metadata.yaml

# Verify changes
mdu comicinfo check --dir ./manga -o after.txt
```

### Example 3: Mixed Manual and MangaDex Metadata

```bash
# Some files have MangaDex data
mdu comicinfo generate \
  --mangadex-id "abc123" \
  --dir ./series1/

# Others need manual updates
mdu comicinfo update ./series2/ch001.cbz \
  --series "Custom Series" \
  --number "1" \
  --writer "Author Name"

# Check everything
mdu comicinfo check --dir . -o full_report.txt
```

### Example 4: Batch Process Multiple Series

```bash
# Create a script for multiple series
#!/bin/bash

# Series 1
mdu comicinfo generate \
  --mangadex-id "id-for-series-1" \
  --dir ./Series_1/

# Series 2
mdu comicinfo generate \
  --mangadex-id "id-for-series-2" \
  --dir ./Series_2/

# Series 3
mdu comicinfo generate \
  --mangadex-id "id-for-series-3" \
  --dir ./Series_3/

# Verify all
for dir in ./*/; do
  echo "Checking $dir"
  mdu comicinfo check --dir "$dir"
done
```

## Troubleshooting

### Common Issues

**"Could not extract chapter number from filename"**
```bash
# Files are skipped if chapter number can't be detected
# Solution: Rename files to include chapter numbers
mv "random_name.cbz" "series_ch001.cbz"
```

**"No ComicInfo.xml found in archive"**
```bash
# File doesn't have ComicInfo.xml
# Solution: Generate or update to add it
mdu comicinfo update --file manga.cbz --series "Series" --number "1"
```

**"MangaDex API error" or "failed to fetch MangaDex metadata"**
```bash
# Wrong MangaDex ID or API issues
# Solution: Use search to find correct ID
mdu comicinfo search "Series Name"
```

**"Failed after 3 attempts: integrity validation failed"**
```bash
# CBZ repackaging failed multiple times
# This is rare - possible causes:
# - Corrupted source CBZ file
# - Disk space issues
# - File permissions problems

# Solution: Check the source file is valid
mdu comicinfo validate original.cbz

# Try with a different file to isolate the issue
```

**"Cannot open CBZ" or "failed to open CBZ"**
```bash
# File is not a valid ZIP archive
# Solution: Re-download or recreate the CBZ file
unzip -t file.cbz  # Test ZIP validity
```

## Technical Details

### CBZ File Structure

CBZ files are ZIP archives containing images and optional metadata:

```
manga.cbz (ZIP archive)
├── 001.jpg
├── 002.jpg
├── 003.jpg
├── ...
└── ComicInfo.xml  ← Metadata file
```

### Dependencies

- Go 1.20 or higher
- `github.com/spf13/cobra` - CLI framework
- `gopkg.in/yaml.v3` - YAML parsing
- `github.com/agnivade/levenshtein` - Fuzzy string matching
- `github.com/mattn/go-runewidth` - Unicode-aware text width
- Standard library: archive/zip, crypto/sha256, encoding/xml, encoding/json

### Supported File Formats

- **EPUB**: `.epub` files (EPUB 2 and EPUB 3)
- **CBZ**: `.cbz` (ZIP-based comic archives)
- **Input**: `.json`, `.yaml`, `.yml` for metadata templates

## Contributing

Contributions are welcome! Please:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Add tests if applicable
5. Commit your changes (`git commit -m 'Add amazing feature'`)
6. Push to the branch (`git push origin feature/amazing-feature`)
7. Open a Pull Request

## License

[MIT License](LICENSE)

## Acknowledgments

- Built for use with [Kavita](https://www.kavitareader.com/) and [Komga](https://komga.org/)
- MangaDex API integration for metadata
- ComicInfo.xml standard from [Anansi Project](https://github.com/anansi-project/comicinfo)
- Fuzzy search powered by Levenshtein distance algorithm

## Support

If you encounter issues:

1. Check this README and examples
2. Search [existing issues](https://github.com/adamfitz/mdu/issues)
3. Open a new issue with:
   - Command you ran (exact command line)
   - Error message (full output)
   - File details (filename pattern, how many files)
   - Operating system and version
   - mdu version (`mdu --version`)

---

**Note:** This tool modifies EPUB and CBZ files. Always keep backups of your original files!