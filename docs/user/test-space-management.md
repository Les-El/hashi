# Resource and Space Management

## Problem

Hashing operations and extensive test suites can stress system resources:
- **Disk Space**: Accumulated temporary files in `/tmp` (or `%TEMP%`) can cause disk exhaustion.
- **Memory**: Hashing multi-gigabyte files can lead to Out-of-Memory (OOM) errors if the tool attempts to buffer the entire file.

## Solutions

The chexum project implements several layers of resource management:

### 1. Memory Management (Streaming)

Chexum follows a **Streaming over Buffering** design principle. It uses `io.Copy` to pipe data directly from file handles into cryptographic hashers. 
- **Constant Footprint**: Memory usage remains small and stable (typically < 50MB) regardless of whether you are hashing a 1KB text file or a 100GB disk image.
- **Zero Buffering**: Files are never loaded into RAM in their entirety.

### 2. Automatic Runtime Cleanup

The `chexum` tool manages its own temporary artifacts:
- **Proactive Checks**: Before starting, chexum checks storage usage. If the partition hosting `/tmp` is > 85% full, it automatically triggers a cleanup of its own stale temporary files.
- **Graceful Exit**: On completion (or cancellation via Ctrl+C), chexum automatically removes its active workspace and temporary files.

### 3. Automated Cleanup in Tests

Each test package (`cmd/chexum`, `cmd/checkpoint`, `cmd/cleanup`, `internal/checkpoint`) includes logic to remove test-created artifacts after completion. 

### 4. Standalone Cleanup Tool

For manual management or CI/CD pipelines, use the `cleanup` utility:

```bash
./cleanup --dry-run          # Preview cleanup
./cleanup --force            # Force cleanup even if storage is healthy
./cleanup --threshold 70     # Set a custom usage threshold for cleanup
```

### 5. Pre-Test Cleanup Script

Before running extensive test suites, you can use the provided shell script:

```bash
bash scripts/cleanup-before-tests.sh
go test ./...
```

## Best Practices

1. **In CI/CD**: Add cleanup to your pipeline before test execution:
   ```yaml
   - name: Clean tmp before tests
     run: bash scripts/cleanup-before-tests.sh
   ```

2. **Monitoring**: Check /tmp usage with:
   ```bash
   df -h /tmp
   du -sh /tmp
   ```

3. **Safe Patterns**: The cleanup system only removes patterns known to be safe:
   - `/tmp/go-build*` - Go compiler artifacts
   - `/tmp/chexum-*` - Project-specific temp files
   - `/tmp/checkpoint-*` - Checkpoint artifacts
   - `/tmp/test-*` - Test temporary files
   - `/tmp/*.tmp` - Generic temporary files

## See Also

- [CleanupManager API Documentation](../internal/checkpoint/cleanup.go)
- [Cleanup Tool Source](../cmd/cleanup/main.go)
- [Performance Guidelines](./performance.md)
