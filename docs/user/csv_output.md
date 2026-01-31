# CSV Output

Chexum provides a machine-readable CSV (Comma-Separated Values) output format designed for easy integration with spreadsheets, databases, and custom scripts.

## Usage

To enable CSV output, use the `--format csv` flag or the `--csv` shorthand:

```bash
chexum --format csv [files...]
chexum --csv [files...]
```

## Format Specification

The CSV output follows a strict four-column structure:

| Column | Description |
| :--- | :--- |
| **Type** | The category of the entry: `FILE`, `REFERENCE`, or `INVALID`. |
| **Path** | The file path or the raw input string (for invalid entries). |
| **Hash** | The computed or provided cryptographic hash. |
| **Algorithm** | The algorithm used (e.g., `sha256`, `md5`). |

### Row Types

- **FILE**: Represents a file that was successfully processed.
- **REFERENCE**: Represents a hash provided as an argument for comparison (e.g., when checking a file against a known hash). The "Path" column will contain `-`.
- **INVALID**: Represents an input that could not be processed (e.g., a file that doesn't exist or an invalid hash string).

---

## Use Cases and Examples

### 1. Basic File Hashing
When hashing one or more files, Chexum lists each file with its computed hash.

**Command:**
```bash
chexum --format csv file1.txt file2.txt
```

**Output:**
```csv
FILE,file1.txt,e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855,sha256
FILE,file2.txt,2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824,sha256
```

### 2. Comparison Mode
When comparing files against a reference hash, the reference hash is included in the output.

**Command:**
```bash
chexum --format csv e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855 file.txt
```

**Output:**
```csv
FILE,file.txt,e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855,sha256
REFERENCE,-,e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855,sha256
```

### 3. Handling Invalid Inputs
If a file is missing or a string is not a valid hash, it is marked as `INVALID`.

**Command:**
```bash
chexum --format csv missing.txt "not-a-hash"
```

**Output:**
```csv
INVALID,missing.txt,-,-
INVALID,not-a-hash,-,-
```

### 4. Recursive Processing
When combined with the recursive flag, CSV output is ideal for generating a manifest of an entire directory.

**Command:**
```bash
chexum -r --format csv ./src
```

**Output:**
```csv
FILE,src/main.go,a1b2c3d4...,sha256
FILE,src/utils.go,e5f6g7h8...,sha256
FILE,src/config.go,i9j0k1l2...,sha256
```

---

## Integration Tips

### Filtering with `awk`
You can easily filter the output using standard Unix tools. For example, to list only the paths of successfully hashed files:

```bash
chexum --format csv * | awk -F, '$1=="FILE" {print $2}'
```

### Importing to Spreadsheets
The output is fully compatible with Excel, Google Sheets, and LibreOffice Calc. You can redirect the output to a file and open it directly:

```bash
chexum -r --format csv . > manifest.csv
```
