# Error Handling and Troubleshooting

This guide explains chexum's error messages and how to troubleshoot common issues.

## Understanding Error Messages

Chexum uses different types of error messages depending on the situation:

### Validation Errors (Always Specific)

These errors provide immediate, actionable feedback:

```bash
# Invalid file extension
$ chexum --output script.py file.txt
Error: output file: output files must have extension: .txt, .json, .csv (got .py)

# Invalid algorithm
$ chexum --algo invalid file.txt
Error: invalid algorithm "invalid": must be one of sha256, md5, sha1, sha512, blake2b

# Invalid output format
$ chexum --format invalid file.txt
Error: invalid output format "invalid": must be one of default, verbose, json, plain
```

**What to do:** Follow the error message guidance to fix the command.

### Generic Errors (Security and System Issues)

Some errors use a generic message for security and system reliability:

```bash
$ chexum --output config.json file.txt
Error: output file: Unknown write/append error

$ chexum --output /protected/file.txt file.txt
Error: output file: Unknown write/append error
```

**What to do:** Use the `--verbose` flag for more details:

```bash
$ chexum --verbose --output config.json file.txt
Error: output file: cannot write to configuration file: config.json

$ chexum --verbose --output /protected/file.txt file.txt
Error: output file: permission denied writing to /protected/file.txt
```

## Common Issues and Solutions

### "Unknown write/append error"

This generic error can occur for several reasons:

1. **Configuration file protection** - You're trying to write to a protected file
2. **Permission issues** - You don't have write access to the location
3. **Disk space** - Not enough space to write the file
4. **Network issues** - Writing to a network location that's unavailable
5. **Path issues** - The path is too long or contains invalid characters

**Solution:** Use `--verbose` to get specific details:

```bash
$ chexum --verbose --output problematic-path.txt file.txt
```

### Output File Restrictions

Chexum protects against accidentally overwriting important files:

**Allowed extensions:** `.txt`, `.json`, `.csv`
```bash
✓ chexum --output results.txt file.txt
✓ chexum --output data.json file.txt
✓ chexum --output report.csv file.txt
```

**Protected files and directories:**
- Configuration files (`.chexum.toml`, `config.json`, etc.)
- Configuration directories (`.chexum/`, `.config/chexum/`)

```bash
✗ chexum --output .chexum.toml file.txt
✗ chexum --output .chexum/output.txt file.txt
```

### Permission Issues

If you encounter permission errors:

1. **Check file permissions:**
   ```bash
   ls -la /path/to/output/directory/
   ```

2. **Ensure directory exists:**
   ```bash
   mkdir -p /path/to/output/directory/
   ```

3. **Use a different location:**
   ```bash
   chexum --output ~/results.txt file.txt
   ```

### Disk Space Issues

If you're running out of disk space:

1. **Check available space:**
   ```bash
   df -h
   ```

2. **Use a different location:**
   ```bash
   chexum --output /tmp/results.txt file.txt
   ```

3. **Clean up temporary files:**
   ```bash
   rm -f /tmp/chexum_*
   ```

## Verbose Mode

The `--verbose` flag provides detailed error information and is your best tool for troubleshooting:

```bash
# Generic error
$ chexum --output problematic.txt file.txt
Error: output file: Unknown write/append error

# Detailed error with --verbose
$ chexum --verbose --output problematic.txt file.txt
Error: output file: permission denied writing to problematic.txt
```

**When to use verbose mode:**
- When you encounter "Unknown write/append error"
- When troubleshooting file access issues
- When reporting bugs or asking for help
- When you need to understand what chexum is doing

## Getting Help

If you're still having trouble:

1. **Try verbose mode first:**
   ```bash
   chexum --verbose [your command]
   ```

2. **Check the examples:**
   ```bash
   chexum --help
   ```

3. **Verify your command syntax:**
   ```bash
   # Basic usage
   chexum file.txt
   
   # With output file
   chexum --output results.txt file.txt
   
   # With specific algorithm
   chexum --algo sha256 file.txt
   ```

4. **Test with a simple case:**
   ```bash
   # Create a test file
   echo "test" > test.txt
   
   # Try basic hashing
   chexum test.txt
   
   # Try with output
   chexum --output results.txt test.txt
   ```

## Security Note

Chexum uses generic error messages for certain operations to prevent information disclosure. This is a security feature, not a bug. The `--verbose` flag provides the details you need while maintaining security for automated tools and scripts.