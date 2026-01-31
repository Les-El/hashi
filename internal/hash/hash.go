// Package hash provides hash computation logic for chexum.
//
// DESIGN PRINCIPLE: Streaming over Buffering
// ------------------------------------------
// Hashing large files (multi-gigabyte ISOs, video files, etc.) can easily
// exhaust available system memory if the entire file is loaded at once.
//
// This package prioritizes memory efficiency by using "Streaming". Instead of
// ioutil.ReadFile (which buffers everything), we use io.Copy to pipe data from
// an open file handle directly into the cryptographic hasher. This maintains a
// constant, small memory footprint regardless of file size.
package hash

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"os"
	"runtime"
	"sync"
	"time"

	"golang.org/x/crypto/blake2b"
)

// Supported hash algorithms
const (
	AlgorithmSHA256  = "sha256"
	AlgorithmMD5     = "md5"
	AlgorithmSHA1    = "sha1"
	AlgorithmSHA512  = "sha512"
	AlgorithmBLAKE2b = "blake2b"
)

// Entry represents a hash computation result for a single file or input.
type Entry struct {
	Original    string    // Original argument (file path or hash string)
	Hash        string    // Computed or provided hash value
	IsFile      bool      // True if this entry represents a file
	IsReference bool      // True if this entry represents a user-provided reference hash
	Error       error     // Processing error, if any
	Size        int64     // File size in bytes
	ModTime     time.Time // File modification time
	Algorithm   string    // Hash algorithm used
}

// MatchGroup represents a group of entries with matching hashes.
// This structure is key to Chexum's "Human-First" grouping logic.
type MatchGroup struct {
	Hash    string  // The common hash value
	Entries []Entry // All entries with this hash
	Count   int     // Number of entries in the group
}

// PoolMatch represents a match between a file and a provided hash string.
type PoolMatch struct {
	FilePath     string
	ComputedHash string
	ProvidedHash string
	Algorithm    string
}

// Result holds the complete results of a hash processing operation.
type Result struct {
	Entries        []Entry       // All processed entries
	Matches        []MatchGroup  // Groups of matching hashes
	Unmatched      []Entry       // Entries with unique hashes
	PoolMatches    []PoolMatch   // Matches between files and provided hash strings
	Unknowns       []string      // Invalid arguments identified as neither files nor hashes
	RefOrphans     []Entry       // Reference hashes that didn't match any processed files
	Errors         []error       // All errors encountered
	Duration       time.Duration // Total processing time
	FilesProcessed int           // Number of files processed
	BytesProcessed int64         // Total bytes processed
}

// Computer handles hash computation for files.
// It abstracts away the specific algorithm implementation from the caller.
type Computer struct {
	algorithm string
}

// NewComputer creates a new hash computer with the specified algorithm.
func NewComputer(algorithm string) (*Computer, error) {
	// We perform validation here to ensure the Computer is always in a valid state
	// when used in subsequent hashing operations.
	switch algorithm {
	case AlgorithmSHA256, AlgorithmMD5, AlgorithmSHA1, AlgorithmSHA512, AlgorithmBLAKE2b:
		return &Computer{algorithm: algorithm}, nil
	default:
		return nil, fmt.Errorf("unsupported algorithm: %s", algorithm)
	}
}

// newHasher returns a new hash.Hash for the configured algorithm.
// Note: This uses the standard library hash.Hash interface, allowing us
// to handle different algorithms polymorphically.
func (c *Computer) newHasher() hash.Hash {
	switch c.algorithm {
	case AlgorithmMD5:
		return md5.New()
	case AlgorithmSHA1:
		return sha1.New()
	case AlgorithmSHA512:
		return sha512.New()
	case AlgorithmBLAKE2b:
		// BLAKE2b-512 (64 bytes output)
		// BLAKE2b is often faster than SHA-2 on modern CPUs.
		h, _ := blake2b.New512(nil)
		return h
	default:
		// We default to SHA-256 as it is the current industry standard
		// for cryptographic security and compatibility.
		return sha256.New()
	}
}

// ComputeFile computes the hash of a file using a streaming approach.
//
// STEP-BY-STEP PROCESS:
// 1. Open the file handle (read-only).
// 2. Fetch file metadata (size, modtime) for the result entry.
// 3. Initialize the cryptographic hasher.
// 4. Use io.Copy to stream data in chunks from the file to the hasher.
// 5. Finalize the hash (Sum) and convert the binary digest to a hex string.
func (c *Computer) ComputeFile(path string) (*Entry, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	// defer ensures the file handle is closed even if io.Copy fails,
	// preventing file descriptor leaks.
	defer file.Close()

	// Get file info for size and mod time before we start hashing.
	info, err := file.Stat()
	if err != nil {
		return nil, err
	}

	// io.Copy handles the heavy lifting of reading from the file and writing
	// to the hasher. By default, it uses a 32KB buffer, which is a good
	// balance between memory usage and performance.
	hasher := c.newHasher()
	size, err := io.Copy(hasher, file)
	if err != nil {
		return nil, err
	}

	return &Entry{
		Original:  path,
		Hash:      hex.EncodeToString(hasher.Sum(nil)),
		IsFile:    true,
		Size:      size,
		ModTime:   info.ModTime(),
		Algorithm: c.algorithm,
	}, nil
}

// ComputeReader computes the hash of data from an io.Reader.
// This allows hashing data from stdin or network streams.
func (c *Computer) ComputeReader(r io.Reader) (string, error) {
	hasher := c.newHasher()
	if _, err := io.Copy(hasher, r); err != nil {
		return "", err
	}
	return hex.EncodeToString(hasher.Sum(nil)), nil
}

// ComputeBytes computes the hash of a byte slice.
// Used primarily for testing or very small metadata strings.
func (c *Computer) ComputeBytes(data []byte) string {
	hasher := c.newHasher()
	hasher.Write(data)
	// hasher.Sum(nil) appends the current hash to nil, returning the digest.
	return hex.EncodeToString(hasher.Sum(nil))
}

// ComputeBatch performs hashing of multiple files using a worker pool.
// It returns a channel that will receive the results as they are computed.
//
// Reviewed: NESTED-LOOP - Standard worker pool pattern for concurrent processing.
func (c *Computer) ComputeBatch(files []string, workers int) <-chan Entry {
	if workers <= 0 {
		workers = runtime.NumCPU()
	}

	jobs := make(chan string, workers)
	// Buffer results slightly to prevent tight coupling, though the consumer should be fast
	results := make(chan Entry, workers)

	// Feeder
	go func() {
		for _, f := range files {
			jobs <- f
		}
		close(jobs)
	}()

	// Workers
	var wg sync.WaitGroup
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			for path := range jobs {
				entry, err := c.ComputeFile(path)
				if err != nil {
					results <- Entry{Original: path, Error: err}
				} else {
					results <- *entry
				}
			}
		}()
	}

	// Closer
	go func() {
		wg.Wait()
		close(results)
	}()

	return results
}

// Algorithm returns the algorithm used by this computer.
func (c *Computer) Algorithm() string {
	return c.algorithm
}

// DetectHashAlgorithm returns possible algorithms for a hash string.
//
// RATIONALE:
// Since different hash algorithms produce digests of specific lengths
// (e.g. MD5 is always 32 chars), we can guess the algorithm by its
// representation. This enables Chexum's "Smart Detection" where users
// don't have to specify --algo if the hash string is unambiguous.
func DetectHashAlgorithm(hashStr string) []string {
	// First validate that string contains only hex characters.
	// This prevents misidentifying random words as hashes.
	if !isValidHexString(hashStr) {
		return []string{}
	}

	// Map lengths to algorithms
	switch len(hashStr) {
	case 32:
		return []string{AlgorithmMD5}
	case 40:
		return []string{AlgorithmSHA1}
	case 64:
		return []string{AlgorithmSHA256}
	case 128:
		// Ambiguous - could be SHA-512 or BLAKE2b-512 as both output 512 bits.
		return []string{AlgorithmSHA512, AlgorithmBLAKE2b}
	default:
		return []string{}
	}
}

// isValidHexString checks if a string contains only valid hexadecimal characters.
func isValidHexString(s string) bool {
	if len(s) == 0 {
		return false
	}

	for _, c := range s {
		// A-F and 0-9 are the only valid digits in hexadecimal.
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}

// IsValidHash checks if a string is a valid hash for the given algorithm.
func IsValidHash(hash, algorithm string) bool {
	expectedLen := 0
	switch algorithm {
	case AlgorithmMD5:
		expectedLen = 32
	case AlgorithmSHA1:
		expectedLen = 40
	case AlgorithmSHA256:
		expectedLen = 64
	case AlgorithmSHA512:
		expectedLen = 128
	case AlgorithmBLAKE2b:
		expectedLen = 128 // BLAKE2b-512 produces 64 bytes = 128 hex chars
	default:
		return false
	}

	if len(hash) != expectedLen {
		return false
	}

	return isValidHexString(hash)
}
