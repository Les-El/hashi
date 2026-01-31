# Hashing Engine Design

## Purpose
The Hashing Engine is responsible for the efficient calculation of cryptographic hashes for one or more files. It is designed to be highly concurrent and memory-efficient.

## Core Concepts

### File Discovery
The engine uses a recursive traversal (when requested) to identify target files. It respects glob patterns for inclusion and exclusion, allowing users to precisely target specific file types or skip directories like `vendor/` or `.git/`.

### The Worker Pool
To maximize performance, the engine employs a concurrency pool:
- **Task Dispatcher:** Walks the file system and pushes file paths into a task channel.
- **Workers:** A pool of goroutines that pull paths from the channel and compute hashes.
- **Result Collector:** Gathers the computed hashes and any errors into a result slice.

### Memory Management
The engine uses `io.Copy` with a small buffer to stream file content into the hashers. This ensures that even very large files can be hashed without loading them entirely into memory.

## Supported Algorithms
- **SHA-256:** Default, balance of security and speed.
- **SHA-512:** High security.
- **BLAKE2b:** High performance.
- **MD5/SHA-1:** Provided for legacy compatibility.

## Error Handling
The engine treats file access errors as non-fatal to the entire operation. If a file cannot be read, an error is recorded for that specific file, but the engine continues processing other files in the queue.
