package hash

import (
	"math/rand"
	"testing"
	"testing/quick"
)

// Note: Removed unused bytes and crypto/hash imports to resolve build failures.
// These should be re-added when their respective hashing functions are
// actively used in the tests ("bytes", "crypto/md5", "crypto/sha1", "crypto/sha256", "crypto/sha512")

func TestProperty_HashDeterminism(t *testing.T) {
	f := func(input []byte) bool {
		if len(input) == 0 {
			return true
		}

		h1, _ := NewComputer("sha256")
		hash1 := h1.ComputeBytes(input)

		h2, _ := NewComputer("sha256")
		hash2 := h2.ComputeBytes(input)

		return hash1 == hash2
	}

	if err := quick.Check(f, nil); err != nil {
		t.Error(err)
	}
}

func TestProperty_HashLength(t *testing.T) {
	tests := []struct {
		algo   string
		length int
	}{
		{"sha256", 64},
		{"md5", 32},
		{"sha1", 40},
		{"sha512", 128},
		{"blake2b", 128},
	}

	for _, tt := range tests {
		f := func(input []byte) bool {
			h, _ := NewComputer(tt.algo)
			hash := h.ComputeBytes(input)
			return len(hash) == tt.length
		}
		if err := quick.Check(f, nil); err != nil {
			t.Errorf("%s: %v", tt.algo, err)
		}
	}
}

func TestProperty_DetectHashAlgorithm(t *testing.T) {
	t.Run("ValidHexLengths", func(t *testing.T) {
		tests := []struct {
			algo   string
			length int
		}{
			{AlgorithmMD5, 32},
			{AlgorithmSHA1, 40},
			{AlgorithmSHA256, 64},
			{AlgorithmSHA512, 128},
		}

		for _, tt := range tests {
			f := func() bool {
				hashStr := generateRandomHex(tt.length)
				detected := DetectHashAlgorithm(hashStr)
				for _, d := range detected {
					if d == tt.algo {
						return true
					}
				}
				return false
			}
			if err := quick.Check(f, nil); err != nil {
				t.Errorf("%s: %v", tt.algo, err)
			}
		}
	})

	t.Run("NonHexStrings", func(t *testing.T) {
		fNonHex := func(s string) bool {
			if !hasNonHexChars(s) {
				return true
			}
			return len(DetectHashAlgorithm(s)) == 0
		}
		if err := quick.Check(fNonHex, nil); err != nil {
			t.Error(err)
		}
	})
}

func generateRandomHex(length int) string {
	const hexBytes = "0123456789abcdef"
	b := make([]byte, length)
	for i := range b {
		b[i] = hexBytes[rand.Intn(len(hexBytes))]
	}
	return string(b)
}

func hasNonHexChars(s string) bool {
	for _, c := range s {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return true
		}
	}
	return false
}

	

func TestProperty_ComputeBatch_LargeLists(t *testing.T) {
	f := func(numFiles int) bool {
		if numFiles < 0 {
			numFiles = -numFiles
		}
		numFiles = numFiles % 500 // Up to 500 files per test run

		files := make([]string, numFiles)
		for i := 0; i < numFiles; i++ {
			files[i] = "non_existent_file_path_for_property_test"
		}

		computer, _ := NewComputer(AlgorithmSHA256)
		// Use auto workers (0)
		resultChan := computer.ComputeBatch(files, 0)

		count := 0
		for range resultChan {
			count++
		}

		return count == numFiles
	}

	// MaxCount 20 is sufficient to test varying sizes without being too slow
	if err := quick.Check(f, &quick.Config{MaxCount: 20}); err != nil {
		t.Error(err)
	}
}