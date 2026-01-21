// Package hash tests for hash computation logic.
package hash

import (
	"bytes"
	"math/rand"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"testing/quick"
)

// TestNewComputer tests computer creation with various algorithms.
func TestNewComputer(t *testing.T) {
	tests := []struct {
		name      string
		algorithm string
		wantErr   bool
	}{
		{"sha256", AlgorithmSHA256, false},
		{"md5", AlgorithmMD5, false},
		{"sha1", AlgorithmSHA1, false},
		{"sha512", AlgorithmSHA512, false},
		{"blake2b", AlgorithmBLAKE2b, false},
		{"invalid", "invalid", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := NewComputer(tt.algorithm)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewComputer() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && c.Algorithm() != tt.algorithm {
				t.Errorf("Algorithm() = %v, want %v", c.Algorithm(), tt.algorithm)
			}
		})
	}
}

// TestComputeBytes tests hash computation for byte slices.
func TestComputeBytes(t *testing.T) {
	tests := []struct {
		name      string
		algorithm string
		data      []byte
		want      string
	}{
		{
			"sha256 empty",
			AlgorithmSHA256,
			[]byte{},
			"e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		},
		{
			"sha256 hello",
			AlgorithmSHA256,
			[]byte("hello"),
			"2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824",
		},
		{
			"md5 empty",
			AlgorithmMD5,
			[]byte{},
			"d41d8cd98f00b204e9800998ecf8427e",
		},
		{
			"md5 hello",
			AlgorithmMD5,
			[]byte("hello"),
			"5d41402abc4b2a76b9719d911017c592",
		},
		{
			"blake2b empty",
			AlgorithmBLAKE2b,
			[]byte{},
			"786a02f742015903c6c6fd852552d272912f4740e15847618a86e217f71f5419d25e1031afee585313896444934eb04b903a685b1448b755d56f701afe9be2ce",
		},
		{
			"blake2b hello",
			AlgorithmBLAKE2b,
			[]byte("hello"),
			"e4cfa39a3d37be31c59609e807970799caa68a19bfaa15135f165085e01d41a65ba1e1b146aeb6bd0092b49eac214c103ccfa3a365954bbbe52f74a2b3620c94",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := NewComputer(tt.algorithm)
			if err != nil {
				t.Fatalf("NewComputer() error = %v", err)
			}

			got := c.ComputeBytes(tt.data)
			if got != tt.want {
				t.Errorf("ComputeBytes() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestComputeReader tests hash computation from a reader.
func TestComputeReader(t *testing.T) {
	c, err := NewComputer(AlgorithmSHA256)
	if err != nil {
		t.Fatalf("NewComputer() error = %v", err)
	}

	data := []byte("test data for reader")
	reader := bytes.NewReader(data)

	hash, err := c.ComputeReader(reader)
	if err != nil {
		t.Fatalf("ComputeReader() error = %v", err)
	}

	// Verify against ComputeBytes
	expected := c.ComputeBytes(data)
	if hash != expected {
		t.Errorf("ComputeReader() = %v, want %v", hash, expected)
	}
}

// TestComputeFile tests hash computation for files.
func TestComputeFile(t *testing.T) {
	// Create a temporary file
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.txt")
	content := []byte("test file content")

	if err := os.WriteFile(tmpFile, content, 0644); err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	c, err := NewComputer(AlgorithmSHA256)
	if err != nil {
		t.Fatalf("NewComputer() error = %v", err)
	}

	entry, err := c.ComputeFile(tmpFile)
	if err != nil {
		t.Fatalf("ComputeFile() error = %v", err)
	}

	// Verify the entry
	if entry.Original != tmpFile {
		t.Errorf("Original = %v, want %v", entry.Original, tmpFile)
	}
	if !entry.IsFile {
		t.Error("IsFile = false, want true")
	}
	if entry.Size != int64(len(content)) {
		t.Errorf("Size = %v, want %v", entry.Size, len(content))
	}
	if entry.Algorithm != AlgorithmSHA256 {
		t.Errorf("Algorithm = %v, want %v", entry.Algorithm, AlgorithmSHA256)
	}

	// Verify hash matches ComputeBytes
	expected := c.ComputeBytes(content)
	if entry.Hash != expected {
		t.Errorf("Hash = %v, want %v", entry.Hash, expected)
	}
}

// TestComputeFile_NotFound tests error handling for missing files.
func TestComputeFile_NotFound(t *testing.T) {
	c, err := NewComputer(AlgorithmSHA256)
	if err != nil {
		t.Fatalf("NewComputer() error = %v", err)
	}

	_, err = c.ComputeFile("/nonexistent/file.txt")
	if err == nil {
		t.Error("ComputeFile() expected error for nonexistent file")
	}
}

// TestIsValidHash tests hash validation.
func TestIsValidHash(t *testing.T) {
	tests := []struct {
		name      string
		hash      string
		algorithm string
		want      bool
	}{
		{"valid sha256", "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855", AlgorithmSHA256, true},
		{"valid md5", "d41d8cd98f00b204e9800998ecf8427e", AlgorithmMD5, true},
		{"valid sha1", "da39a3ee5e6b4b0d3255bfef95601890afd80709", AlgorithmSHA1, true},
		{"valid blake2b", "786a02f742015903c6c6fd852552d272912f4740e15847618a86e217f71f5419d25e1031afee585313896444934eb04b903a685b1448b755d56f701afe9be2ce", AlgorithmBLAKE2b, true},
		{"too short", "abc123", AlgorithmSHA256, false},
		{"too long", "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855aa", AlgorithmSHA256, false},
		{"invalid chars", "g3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855", AlgorithmSHA256, false},
		{"uppercase valid", "E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855", AlgorithmSHA256, true},
		{"invalid algorithm", "abc123", "invalid", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidHash(tt.hash, tt.algorithm)
			if got != tt.want {
				t.Errorf("IsValidHash() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Property-based test: ComputeBytes should be deterministic
// For any input, computing the hash twice should give the same result
func TestComputeBytes_Deterministic(t *testing.T) {
	c, err := NewComputer(AlgorithmSHA256)
	if err != nil {
		t.Fatalf("NewComputer() error = %v", err)
	}

	f := func(data []byte) bool {
		hash1 := c.ComputeBytes(data)
		hash2 := c.ComputeBytes(data)
		return hash1 == hash2
	}

	if err := quick.Check(f, nil); err != nil {
		t.Error(err)
	}
}

// Property-based test: Different inputs should (almost always) produce different hashes
// This is a probabilistic test - collisions are theoretically possible but extremely unlikely
func TestComputeBytes_DifferentInputs(t *testing.T) {
	c, err := NewComputer(AlgorithmSHA256)
	if err != nil {
		t.Fatalf("NewComputer() error = %v", err)
	}

	f := func(data1, data2 []byte) bool {
		// Skip if inputs are equal
		if bytes.Equal(data1, data2) {
			return true
		}

		hash1 := c.ComputeBytes(data1)
		hash2 := c.ComputeBytes(data2)

		// Different inputs should produce different hashes
		return hash1 != hash2
	}

	if err := quick.Check(f, nil); err != nil {
		t.Error(err)
	}
}

// Property-based test: Hash length should be consistent for an algorithm
func TestComputeBytes_ConsistentLength(t *testing.T) {
	algorithms := []struct {
		name   string
		length int
	}{
		{AlgorithmSHA256, 64},
		{AlgorithmMD5, 32},
		{AlgorithmSHA1, 40},
		{AlgorithmSHA512, 128},
		{AlgorithmBLAKE2b, 128},
	}

	for _, alg := range algorithms {
		t.Run(alg.name, func(t *testing.T) {
			c, err := NewComputer(alg.name)
			if err != nil {
				t.Fatalf("NewComputer() error = %v", err)
			}

			f := func(data []byte) bool {
				hash := c.ComputeBytes(data)
				return len(hash) == alg.length
			}

			if err := quick.Check(f, nil); err != nil {
				t.Error(err)
			}
		})
	}
}

// Property 27: Hash algorithm detection validates hex characters
// **Validates: Requirements 21.1**
func TestDetectHashAlgorithm_HexValidation(t *testing.T) {
	f := func(s string) bool {
		algorithms := DetectHashAlgorithm(s)
		
		// If algorithms were detected, the string must be valid hex
		if len(algorithms) > 0 {
			// Check that string contains only hex characters
			for _, c := range s {
				if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
					return false // Invalid hex character found but algorithms detected
				}
			}
			// Also check that length matches expected algorithm lengths
			validLengths := []int{32, 40, 64, 128}
			validLength := false
			for _, length := range validLengths {
				if len(s) == length {
					validLength = true
					break
				}
			}
			if !validLength {
				return false // Invalid length but algorithms detected
			}
		}
		
		// If no algorithms detected, either invalid hex or invalid length
		if len(algorithms) == 0 {
			// Check if it's invalid hex
			hasInvalidHex := false
			for _, c := range s {
				if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
					hasInvalidHex = true
					break
				}
			}
			
			// If it has valid hex but no algorithms, it must be wrong length
			if !hasInvalidHex {
				validLengths := []int{32, 40, 64, 128}
				for _, length := range validLengths {
					if len(s) == length {
						return false // Valid hex and valid length but no algorithms detected
					}
				}
			}
		}
		
		return true
	}

	if err := quick.Check(f, nil); err != nil {
		t.Error(err)
	}
}

// Property 28: Hash algorithm detection identifies correct algorithms by length
// **Validates: Requirements 21.2**
func TestDetectHashAlgorithm_AlgorithmIdentification(t *testing.T) {
	// Generate valid hex strings of specific lengths and verify correct algorithms are returned
	f := func(length int, hexChars []byte) bool {
		// Only test valid algorithm lengths
		validLengths := map[int][]string{
			32:  {AlgorithmMD5},
			40:  {AlgorithmSHA1},
			64:  {AlgorithmSHA256},
			128: {AlgorithmSHA512, AlgorithmBLAKE2b},
		}
		
		expectedAlgorithms, isValidLength := validLengths[length]
		if !isValidLength {
			return true // Skip invalid lengths
		}
		
		// Create a hex string of the specified length
		hexString := ""
		for i := 0; i < length; i++ {
			// Use modulo to cycle through hex characters
			if len(hexChars) == 0 {
				hexString += "0" // Default to '0' if no hex chars provided
			} else {
				char := hexChars[i%len(hexChars)]
				// Ensure it's a valid hex character
				if (char >= '0' && char <= '9') || (char >= 'a' && char <= 'f') || (char >= 'A' && char <= 'F') {
					hexString += string(char)
				} else {
					hexString += "0" // Default to '0' for invalid chars
				}
			}
		}
		
		algorithms := DetectHashAlgorithm(hexString)
		
		// Check that the returned algorithms match expected
		if len(algorithms) != len(expectedAlgorithms) {
			return false
		}
		
		// Check that all expected algorithms are present
		for _, expected := range expectedAlgorithms {
			found := false
			for _, actual := range algorithms {
				if actual == expected {
					found = true
					break
				}
			}
			if !found {
				return false
			}
		}
		
		return true
	}

	// Custom generator for valid lengths and hex characters
	config := &quick.Config{
		Values: func(values []reflect.Value, rand *rand.Rand) {
			// Generate a valid length
			validLengths := []int{32, 40, 64, 128}
			length := validLengths[rand.Intn(len(validLengths))]
			values[0] = reflect.ValueOf(length)
			
			// Generate some hex characters
			hexChars := "0123456789abcdefABCDEF"
			numChars := rand.Intn(10) + 1 // 1-10 characters
			chars := make([]byte, numChars)
			for i := range chars {
				chars[i] = hexChars[rand.Intn(len(hexChars))]
			}
			values[1] = reflect.ValueOf(chars)
		},
	}

	if err := quick.Check(f, config); err != nil {
		t.Error(err)
	}
}

// TestDetectHashAlgorithm tests hash algorithm detection with specific cases.
func TestDetectHashAlgorithm(t *testing.T) {
	tests := []struct {
		name     string
		hashStr  string
		expected []string
	}{
		// Valid hex strings of each length
		{"valid md5", "d41d8cd98f00b204e9800998ecf8427e", []string{AlgorithmMD5}},
		{"valid sha1", "da39a3ee5e6b4b0d3255bfef95601890afd80709", []string{AlgorithmSHA1}},
		{"valid sha256", "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855", []string{AlgorithmSHA256}},
		{"valid sha512", "cf83e1357eefb8bdf1542850d66d8007d620e4050b5715dc83f4a921d36ce9ce47d0d13c5d85f2b0ff8318d2877eec2f63b931bd47417a81a538327af927da3e", []string{AlgorithmSHA512, AlgorithmBLAKE2b}},
		{"valid blake2b (same length as sha512)", "786a02f742015903c6c6fd852552d272912f4740e15847618a86e217f71f5419d25e1031afee585313896444934eb04b903a685b1448b755d56f701afe9be2ce", []string{AlgorithmSHA512, AlgorithmBLAKE2b}},
		
		// Uppercase should work
		{"uppercase md5", "D41D8CD98F00B204E9800998ECF8427E", []string{AlgorithmMD5}},
		{"uppercase sha256", "E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855", []string{AlgorithmSHA256}},
		
		// Mixed case should work
		{"mixed case sha1", "Da39A3ee5E6b4B0d3255BfeF95601890aFd80709", []string{AlgorithmSHA1}},
		
		// Invalid hex characters
		{"invalid hex chars", "g3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855", []string{}},
		{"invalid hex with space", "e3b0c442 8fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855", []string{}},
		{"invalid hex with dash", "e3b0c442-98fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855", []string{}},
		
		// Wrong lengths
		{"too short", "abc123", []string{}},
		{"too long", "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855aa", []string{}},
		{"wrong length 16", "1234567890abcdef", []string{}},
		{"wrong length 48", "123456789012345678901234567890123456789012345678", []string{}},
		
		// Edge cases
		{"empty string", "", []string{}},
		{"single char", "a", []string{}},
		{"all zeros md5", "00000000000000000000000000000000", []string{AlgorithmMD5}},
		{"all zeros sha256", "0000000000000000000000000000000000000000000000000000000000000000", []string{AlgorithmSHA256}},
		{"all f's sha512", "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff", []string{AlgorithmSHA512, AlgorithmBLAKE2b}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DetectHashAlgorithm(tt.hashStr)
			
			// Check length matches
			if len(got) != len(tt.expected) {
				t.Errorf("DetectHashAlgorithm() returned %d algorithms, expected %d", len(got), len(tt.expected))
				t.Errorf("Got: %v, Expected: %v", got, tt.expected)
				return
			}
			
			// Check all expected algorithms are present
			for _, expected := range tt.expected {
				found := false
				for _, actual := range got {
					if actual == expected {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("DetectHashAlgorithm() missing expected algorithm %s", expected)
					t.Errorf("Got: %v, Expected: %v", got, tt.expected)
				}
			}
		})
	}
}
