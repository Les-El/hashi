# ADR-002: JSON & JSONL Output Formats

## Status
**ACCEPTED** - Implementation in progress

## Context & Problem Statement
`chexum` requires structured output formats to move beyond simple text-parsing and support modern CI/CD pipelines. We need to support both single-operation reports and high-volume data streaming while maintaining an audit trail of how hashes were generated.

## Decision

### A. Dual Output Format Strategy
We will implement two complementary JSON-based output formats:

#### `--json` (The Final Report)
- **Purpose:** A single, valid JSON document representing a complete "session"
- **Structure:** An "Envelope" object containing a `chexum` metadata block and a `results` array
- **Use Case:** Small-to-medium batches where the user needs a verifiable record of the command and version used
- **Metadata:** Includes normalized command string (expanded flags), version, and start/end timestamps

#### `--jsonl` (The Data Stream)
- **Purpose:** Line-delimited JSON for high-performance streaming
- **Structure:** One independent JSON object per line
- **Use Case:** Large-scale file discovery (millions of files), piping into `jq`, or database ingestion
- **Metadata:** By default, "pure" data (one file per line). A "Header" record may be requested via flag but is excluded by default to avoid breaking `jq` filters

### B. Unified Schema Design
Both formats will share a "Flat Ergonomic" schema for individual items to ensure predictability.

#### Core Fields
| Field | Type | Description |
|-------|------|-------------|
| `type` | string | The entity type: `file`, `directory`, `string_literal`, or `hex_literal` |
| `name` | string | The normalized path (using `/`) or input label |
| `hash` | string | The calculated or found hash string |
| `status` | string | The outcome: `success`, `failure`, `match`, or `mismatch` |
| `reference` | object | (Optional) An object describing what the hash was compared against |
| `meta` | object | File system metadata (size_bytes, mtime, permissions) |
| `timestamp` | string | ISO-8601 timestamp of when the hashing occurred |

### C. Stream Separation & Integrity
- **STDOUT is the API:** Only JSON/JSONL artifacts are emitted to STDOUT
- **STDERR is for Humans:** Progress bars, colored warnings, and "Done" messages are strictly relegated to STDERR
- **Error Resilience:** If a file cannot be read, it still generates a JSON object with `status: "failure"` and an error message, ensuring the stream count matches the file discovery count

### D. Command Normalization
To support reproducibility, `chexum` will implement a normalization engine:
- Shorthand flags (e.g., `-r -j`) are expanded to long-form (e.g., `--recursive --json`) in the recorded metadata
- This ensures that the "Command" field in a JSON report is a canonical representation of the operation

## Examples

### JSON Format Example
```json
{
  "chexum": {
    "version": "1.2.0",
    "command": "chexum --recursive --json --algorithm=sha256 /path/to/files",
    "start_time": "2024-01-15T10:30:00Z",
    "end_time": "2024-01-15T10:30:05Z"
  },
  "results": [
    {
      "type": "file",
      "name": "docs/README.md",
      "hash": "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
      "status": "success",
      "meta": {
        "size_bytes": 1024,
        "mtime": "2024-01-15T09:15:00Z",
        "permissions": "644"
      },
      "timestamp": "2024-01-15T10:30:01Z"
    }
  ]
}
```

### JSONL Format Example
```jsonl
{"type":"file","name":"docs/README.md","hash":"e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855","status":"success","meta":{"size_bytes":1024,"mtime":"2024-01-15T09:15:00Z","permissions":"644"},"timestamp":"2024-01-15T10:30:01Z"}
{"type":"file","name":"src/main.go","hash":"a1b2c3d4e5f6789012345678901234567890123456789012345678901234567890","status":"success","meta":{"size_bytes":2048,"mtime":"2024-01-15T09:20:00Z","permissions":"644"},"timestamp":"2024-01-15T10:30:02Z"}
```

## Rationale

### Why Two Formats?
1. **Different Use Cases:** Single reports vs. streaming data have fundamentally different requirements
2. **Performance:** JSONL allows constant memory usage for large datasets
3. **Tooling Compatibility:** JSON works with standard parsers, JSONL works with line-based tools like `jq`, `grep`, and `awk`

### Why Unified Schema?
1. **Predictability:** Users can write tools that work with both formats
2. **Consistency:** Same field names and types across all outputs
3. **Migration Path:** Easy to convert between formats

### Why Stream Separation?
1. **API Clarity:** STDOUT becomes a clean data interface
2. **Pipeline Compatibility:** Works seamlessly with Unix pipes and redirection
3. **Error Handling:** Human-readable errors don't corrupt machine-readable output

## Consequences

### Positive
- High interoperability with `jq`, `grep`, and other CLI tools
- Constant memory footprint for `--jsonl` mode
- Robust audit trails for security compliance
- Clear separation between human and machine interfaces
- Reproducible operations through command normalization

### Negative
- Slightly more complex `jq` queries for the `--json` format due to the `results` wrapper
- Requires careful handling of "Header" records in JSONL streams if implemented
- Additional complexity in output handling logic

### Neutral
- Existing text-based output formats remain unchanged
- New flags (`--json`, `--jsonl`) are additive, not breaking changes

## Implementation Notes

### Flag Integration
- `--json` and `--jsonl` are mutually exclusive
- Both flags are incompatible with existing format flags (`--plain`, `--verbose`)
- Error handling ensures graceful degradation when files cannot be processed

### Performance Considerations
- JSONL streaming should not buffer entire result sets in memory
- JSON format can buffer for small-to-medium datasets but should warn on large operations
- Progress indicators remain on STDERR for both formats

### Testing Strategy
- Unit tests for schema validation
- Integration tests for pipeline compatibility
- Performance tests for large file sets
- Compatibility tests with common JSON tools (`jq`, `python -m json.tool`)

## Related Decisions
- This ADR supersedes any previous informal decisions about output formats
- Complements ADR-001 regarding flag simplification philosophy
- Aligns with the project's "script-friendly" design goals from the README

## References
- [JSON Lines specification](http://jsonlines.org/)
- [RFC 7159: The JavaScript Object Notation (JSON) Data Interchange Format](https://tools.ietf.org/html/rfc7159)
- [jq Manual](https://stedolan.github.io/jq/manual/)