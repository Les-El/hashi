# Scripting with hashi

## The -b Flag (Boolean Mode)

The `-b` / `--bool` flag is the easiest way to use hashi in scripts. It outputs just `true` or `false`:

```bash
# Simple comparison
$ hashi -b file1.txt file2.txt
true

# Use in conditions
$ hashi -b file1.txt file2.txt && echo "match" || echo "different"
match

# Capture result
MATCH=$(hashi -b file1.txt file2.txt)
```

See [BOOL_MODE.md](BOOL_MODE.md) for complete documentation.

## Exit Codes

hashi uses meaningful exit codes for scripting:

```bash
0   - Success (all files processed, or matches found with --match-required)
1   - No matches found (with --match-required)
2   - Some files failed to process
3   - Invalid arguments
4   - File not found
5   - Permission denied
130 - Interrupted (Ctrl-C)
```

## Basic Patterns

### Check if Two Files Match

```bash
# Using -b flag (recommended)
if hashi -b file1.txt file2.txt; then
    echo "Files match"
else
    echo "Files differ"
fi

# Or capture the boolean
MATCH=$(hashi -b file1.txt file2.txt)
[ "$MATCH" = "true" ] && echo "match" || echo "differ"

# Using exit code with --quiet --match-required
hashi --quiet --match-required file1.txt file2.txt && echo "match" || echo "differ"
```

### Get Just the Hash

```bash
# Capture hash output
HASH=$(hashi --quiet file.txt)

# Use in comparison
if [ "$(hashi --quiet file1.txt)" = "$(hashi --quiet file2.txt)" ]; then
    echo "Files match"
fi
```

### Process Multiple Files

```bash
# Find duplicates
hashi --quiet *.txt | sort | uniq -d

# Count unique files
hashi --quiet *.txt | sort -u | wc -l
```

## Advanced Patterns

### Conditional Processing

```bash
# Only process if files don't match (using -b)
if ! hashi -b old.txt new.txt; then
    echo "Files changed, processing..."
    process_file new.txt
fi

# Alternative with exit code
if ! hashi --quiet --match-required old.txt new.txt; then
    echo "Files changed, processing..."
    process_file new.txt
fi
```

### Parallel Processing

```bash
# Hash files in parallel
find . -type f | parallel hashi --quiet {}
```

### Error Handling

```bash
# Capture exit code
hashi --quiet file.txt
EXIT_CODE=$?

case $EXIT_CODE in
    0) echo "Success" ;;
    4) echo "File not found" >&2 ;;
    5) echo "Permission denied" >&2 ;;
    *) echo "Unknown error: $EXIT_CODE" >&2 ;;
esac
```

## The -b Flag vs --quiet

### When to Use -b
- You want a true/false answer
- You're using the result in conditions
- You need to capture the boolean value
- Maximum simplicity and clarity

```bash
# Perfect for -b
hashi -b file1 file2 && echo "match"
RESULT=$(hashi -b file1 file2)
```

### When to Use --quiet
- You want the actual hash values (not just boolean)
- You're piping output to other tools
- You need exit codes but also want output
- You're using other flags that affect output format

```bash
# Better with --quiet
HASH=$(hashi --quiet file.txt)
hashi --quiet *.txt | sort | uniq
```

## The --quiet Flag

`--quiet` is essential for scripting. It:
- Suppresses progress indicators
- Suppresses informational messages
- Suppresses warnings (including flag conflicts)
- Keeps only the essential output

**Important:** Errors are NEVER suppressed, even with `--quiet`.

```bash
# Good for scripting
hashi --quiet file.txt

# Bad for scripting (too much noise)
hashi file.txt
```

## Output Formats

### Default Format
Human-readable, includes headers and formatting:
```
Comparing 2 files...
✓ file1.txt: a1b2c3d4...
✓ file2.txt: a1b2c3d4...
```

### Quiet Mode
Minimal, machine-readable:
```
a1b2c3d4... file1.txt
a1b2c3d4... file2.txt
```

### JSON Format
Structured data (use with `--json`, not `--quiet`):
```json
{
  "files": [
    {"path": "file1.txt", "hash": "a1b2c3d4..."},
    {"path": "file2.txt", "hash": "a1b2c3d4..."}
  ]
}
```

### Plain Format
Just the hashes:
```
a1b2c3d4...
a1b2c3d4...
```

## Future: Boolean Output

A potential `--bool` flag for even simpler scripting:

```bash
# Output just "true" or "false"
hashi --bool file1.txt file2.txt
# Output: true

# Use in conditions
if [ "$(hashi --bool file1.txt file2.txt)" = "true" ]; then
    echo "Match"
fi
```

However, this might be redundant since exit codes already provide this:

```bash
# Equivalent using exit codes
hashi --quiet --match-required file1.txt file2.txt && echo "true" || echo "false"
```

## Best Practices

1. **Use -b for boolean checks** - it's the simplest and clearest
2. **Use --quiet for hash extraction** - when you need the actual values
3. **Check exit codes** rather than parsing output when possible
4. **Use --match-required** only if not using -b
5. **Redirect stderr** if you want to suppress errors (but usually don't)
6. **Use --json** if you need structured data for complex processing

## Examples

### Backup Verification Script

```bash
#!/bin/bash
set -e

BACKUP_FILE="backup.tar.gz"
ORIGINAL_FILE="original.tar.gz"

# Simple boolean check
if hashi -b "$ORIGINAL_FILE" "$BACKUP_FILE"; then
    echo "Backup verified successfully"
    exit 0
else
    echo "ERROR: Backup verification failed!" >&2
    exit 1
fi
```

### Duplicate Finder

```bash
#!/bin/bash

# Find all duplicate files in a directory
find . -type f -print0 | \
    xargs -0 hashi --quiet | \
    sort | \
    uniq -d -w 64 | \
    cut -d' ' -f2-
```

### Change Detection

```bash
#!/bin/bash

CONFIG_FILE="/etc/myapp/config.yaml"
BACKUP_FILE="/etc/myapp/config.yaml.bak"

# Check if changed using -b
if ! hashi -b "$CONFIG_FILE" "$BACKUP_FILE"; then
    echo "Config changed, reloading..."
    cp "$CONFIG_FILE" "$BACKUP_FILE"
    systemctl reload myapp
else
    echo "Config unchanged"
fi
```
