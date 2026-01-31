package config

import (
	"testing"
)

func TestClassifyArguments(t *testing.T) {
	tests := []struct {
		name          string
		args          []string
		algorithm     string
		wantFiles     int
		wantHashes    int
		wantUnknowns  int
	}{
		{"files only", []string{"config.go", "cli.go"}, "sha256", 2, 0, 0},
		{"hashes only", []string{"e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"}, "sha256", 0, 1, 0},
		{"mixed", []string{"config.go", "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"}, "sha256", 1, 1, 0},
		{"stdin", []string{"-"}, "sha256", 1, 0, 0},
		{"invalid hash length for algo", []string{"d41d8cd98f00b204e9800998ecf8427e"}, "sha256", 0, 0, 1},
		{"invalid hex", []string{"not-a-hash-but-looks-like-one-if-it-had-hex-only-0123456789abcdefg"}, "sha256", 0, 0, 1},
		{"hash like but unknown length", []string{"abcde"}, "sha256", 0, 0, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			files, hashes, unknowns, err := ClassifyArguments(tt.args, tt.algorithm)
			if err != nil {
				t.Errorf("ClassifyArguments() unexpected error = %v", err)
				return
			}
			if len(files) != tt.wantFiles {
				t.Errorf("got %d files, want %d", len(files), tt.wantFiles)
			}
			if len(hashes) != tt.wantHashes {
				t.Errorf("got %d hashes, want %d", len(hashes), tt.wantHashes)
			}
			if len(unknowns) != tt.wantUnknowns {
				t.Errorf("got %d unknowns, want %d", len(unknowns), tt.wantUnknowns)
			}
		})
	}
}

func TestDetectHashAlgorithm(t *testing.T) {
	tests := []struct {
		hash string
		want []string
	}{
		{"d41d8cd98f00b204e9800998ecf8427e", []string{"md5"}},
		{"da39a3ee5e6b4b0d3255bfef95601890afd80709", []string{"sha1"}},
		{"e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855", []string{"sha256"}},
		{"cf83e1357eefb8bdf1542850d66d8007d620e4050b5715dc83f4a921d36ce9ce47d0d13c5d85f2b0ff8318d2877eec2f63b931bd47417a81a538327af927da3e", []string{"sha512", "blake2b"}},
		{"too-short", []string{}},
		{"nothex", []string{}},
	}

	for _, tt := range tests {
		got := detectHashAlgorithm(tt.hash)
		if len(got) != len(tt.want) {
			t.Errorf("detectHashAlgorithm(%s) = %v, want %v", tt.hash, got, tt.want)
			continue
		}
		for i := range got {
			if got[i] != tt.want[i] {
				t.Errorf("detectHashAlgorithm(%s)[%d] = %s, want %s", tt.hash, i, got[i], tt.want[i])
			}
		}
	}
}
