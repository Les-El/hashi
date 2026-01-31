# Previewing Operations with Dry Run

chexum allows you to preview which files will be processed before you start any expensive hashing operations. This is especially useful when testing complex filters.

## Using Dry Run (`--dry-run`)

When you use the `--dry-run` flag, chexum will:
1. Discover files based on your arguments and recursion settings.
2. Apply any active filters (`--include`, `--min-size`, etc.).
3. Display a list of files that passed the filters.
4. Show a summary with the total file count, aggregate size, and estimated hashing time.

```bash
# Preview hashing all .go files recursively
chexum -r --include "*.go" --dry-run
```

## Example Output

```text
Dry Run: Previewing files that would be processed

cmd/chexum/main.go    (estimated size: 12.5 KB)
internal/hash/hash.go    (estimated size: 4.2 KB)
internal/config/config.go    (estimated size: 22.1 KB)

Summary:
  Files to process: 3
  Aggregate size:   38.8 KB
  Estimated time:   < 1s
```

## Why Use Dry Run?

- **Verify Filters**: Make sure your `--exclude` patterns aren't accidentally skipping files you need.
- **Estimate Time**: Get a rough idea of how long a large batch will take to process.
- **Safety**: Confirm you're hashing the intended directory structure before generating a large JSON report or manifest.

> **Note**: Estimated time is based on average hashing speeds and may vary significantly depending on your hardware and current system load.
