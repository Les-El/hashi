// Package hash provides hash computation logic for hashi.
//
// It supports multiple hash algorithms (SHA-256, SHA-512, SHA-1, MD5, BLAKE2b)
// and uses streaming file processing for memory-efficient handling
// of large files.
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
	Original  string    // Original argument (file path or hash string)
	Hash      string    // Computed or provided hash value
	IsFile    bool      // True if this entry represents a file
	Error     error     // Processing error, if any
	Size      int64     // File size in bytes
	ModTime   time.Time // File modification time
	Algorithm string    // Hash algorithm used
}

// MatchGroup represents a group of entries with matching hashes.
type MatchGroup struct {
	Hash    string  // The common hash value
	Entries []Entry // All entries with this hash
	Count   int     // Number of entries in the group
}

// Result holds the complete results of a hash processing operation.
type Result struct {
	Entries        []Entry      // All processed entries
	Matches        []MatchGroup // Groups of matching hashes
	Unmatched      []Entry      // Entries with unique hashes
	Errors         []error      // All errors encountered
	Duration       time.Duration // Total processing time
	FilesProcessed int          // Number of files processed
	BytesProcessed int64        // Total bytes processed
}

// Computer handles hash computation for files.
type Computer struct {
	algorithm string
}

// NewComputer creates a new hash computer with the specified algorithm.
func NewComputer(algorithm string) (*Computer, error) {
	// Validate algorithm
	switch algorithm {
	case AlgorithmSHA256, AlgorithmMD5, AlgorithmSHA1, AlgorithmSHA512, AlgorithmBLAKE2b:
		return &Computer{algorithm: algorithm}, nil
	default:
		return nil, fmt.Errorf("unsupported algorithm: %s", algorithm)
	}
}

// newHasher returns a new hash.Hash for the configured algorithm.
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
		h, _ := blake2b.New512(nil)
		return h
	default:
		return sha256.New()
	}
}

// ComputeFile computes the hash of a file.
func (c *Computer) ComputeFile(path string) (*Entry, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Get file info for size and mod time
	info, err := file.Stat()
	if err != nil {
		return nil, err
	}

	// Compute hash using streaming
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
func (c *Computer) ComputeReader(r io.Reader) (string, error) {
	hasher := c.newHasher()
	if _, err := io.Copy(hasher, r); err != nil {
		return "", err
	}
	return hex.EncodeToString(hasher.Sum(nil)), nil
}

// ComputeBytes computes the hash of a byte slice.
func (c *Computer) ComputeBytes(data []byte) string {
	hasher := c.newHasher()
	hasher.Write(data)
	return hex.EncodeToString(hasher.Sum(nil))
}

// Algorithm returns the algorithm used by this computer.
func (c *Computer) Algorithm() string {
	return c.algorithm
}

// DetectHashAlgorithm returns possible algorithms for a hash string.
// Returns empty slice if not a valid hash format.
// Returns multiple algorithms if ambiguous (e.g., SHA-512 and BLAKE2b-512 both = 128 chars).
func DetectHashAlgorithm(hashStr string) []string {
	// First validate that string contains only hex characters
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
		// Ambiguous - could be SHA-512 or BLAKE2b-512
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

	// Check if all characters are valid hex
	return isValidHexString(hash)
}
