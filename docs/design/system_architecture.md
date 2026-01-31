# System Architecture

## Overview
Chexum is a human-first hashing utility designed for performance, reliability, and ease of use. This document outlines the high-level architecture of the system and how its various components interact.

## Component Map

### 1. CLI Entry Point (`cmd/chexum`)
The primary interface for users. It handles:
- Flag parsing via `internal/config`.
- Environment setup and signal handling.
- Execution flow control (standard mode vs. match mode).

### 2. Configuration System (`internal/config`)
Manages application state based on CLI flags, environment variables, and configuration files. It ensures that inputs are validated and conflicts are resolved before the core logic executes.

### 3. Hashing Engine (`internal/hash`)
The core of the application. It provides:
- File discovery and filtering.
- Concurrent hashing using a worker pool.
- Support for multiple algorithms (SHA-256, SHA-512, BLAKE2b, etc.).
- Performance optimizations like buffered I/O and streaming.

### 4. Analysis Engine (`internal/checkpoint`)
An internal quality assurance tool that:
- Analyzes codebase for standards compliance.
- Monitors test coverage.
- Catalogs and validates CLI flags.
- Generates remediation plans and project health dashboards.

### 5. Output Management (`internal/console`, `internal/output`)
Handles terminal I/O, colorization, and progress indicators. It supports multiple output formats including Plain, JSON, JSONL, and CSV.

## Data Flow
1. **Input:** User provides files/directories and flags.
2. **Parsing:** `config` package parses and validates inputs.
3. **Discovery:** `hash` package identifies files matching inclusion/exclusion criteria.
4. **Processing:** Worker pool hashes files concurrently.
5. **Collection:** Results are aggregated and sorted.
6. **Output:** Results are formatted and displayed via `output` and `console`.

## Concurrency Model
Chexum uses a worker pool pattern for hashing. A dispatcher scans the file system and sends tasks to a channel. A fixed number of workers (proportional to CPU cores or user-defined) consume these tasks and perform the hashing. This ensures high throughput without overwhelming system resources.
