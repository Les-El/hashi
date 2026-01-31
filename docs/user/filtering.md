# File Filtering in chexum

chexum provides powerful filtering options to help you focus on the files that matter. Filters can be combined to create complex selection rules.

## Pattern Matching

You can include or exclude files based on glob patterns.

### Including Files (`--include`, `-i`)
Only files matching the pattern(s) will be processed.

```bash
# Only process .go files
chexum --include "*.go"

# Only process .jpg and .png files
chexum -i "*.jpg" -i "*.png"
```

### Excluding Files (`--exclude`, `-e`)
Files matching the pattern(s) will be skipped. Exclude patterns take precedence over include patterns.

```bash
# Skip all log files
chexum --exclude "*.log"

# Process all files except those in node_modules (if using -r)
chexum -r --exclude "node_modules/*"
```

## Size Filtering

Filter files based on their size on disk.

### Minimum Size (`--min-size`)
Only process files larger than or equal to the specified size.

```bash
# Only process files at least 1MB
chexum --min-size 1MB

# Units supported: B, KB, MB, GB, TB
chexum --min-size 500KB
```

### Maximum Size (`--max-size`)
Only process files smaller than or equal to the specified size.

```bash
# Skip files larger than 1GB
chexum --max-size 1GB
```

## Date Filtering

Filter files based on their last modification time. Dates should be in `YYYY-MM-DD` format.

### Modified After (`--modified-after`)
Only process files modified on or after the given date.

```bash
# Only process files modified in 2024
chexum --modified-after 2024-01-01
```

### Modified Before (`--modified-before`)
Only process files modified on or before the given date.

```bash
# Only process old files
chexum --modified-before 2023-12-31
```

## Combining Filters

When multiple filters are provided, they are combined using **AND** logic. A file must pass **ALL** active filters to be processed.

```bash
# Process .pdf files larger than 10MB modified this year
chexum --include "*.pdf" --min-size 10MB --modified-after 2024-01-01
```
