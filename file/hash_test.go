package file

import (
	"bytes"
	"context"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"errors"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/elvinchan/util-collects/as"
)

func TestHasher(t *testing.T) {
	testCases := []struct {
		name        string
		meta        []HashMeta
		input       HashType
		expectValid bool
	}{
		{
			name:        "single type",
			meta:        []HashMeta{{HashMD5, md5.New}},
			input:       HashMD5,
			expectValid: true,
		},
		{
			name:        "valid and invalid mix types",
			meta:        []HashMeta{{HashMD5, md5.New}},
			input:       HashMD5 | 0x1000,
			expectValid: false,
		},
		{
			name:        "no registered type",
			meta:        []HashMeta{},
			input:       0x1000,
			expectValid: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			hasher := NewHasher(WithHashMeta(tc.meta...))

			as.Equal(t, tc.expectValid, hasher.IsValid(tc.input))

			_, _, err := hasher.ReaderHash(context.Background(), strings.NewReader("test"), tc.input)
			if tc.expectValid {
				as.NoError(t, err)
			} else {
				as.Error(t, err)
			}
		})
	}
}

func TestReaderHash(t *testing.T) {
	testContent := []byte("test data for hashing")
	expectedMD5 := md5.Sum(testContent)
	expectedSHA1 := sha1.Sum(testContent)

	tests := []struct {
		name        string
		input       io.Reader
		ht          HashType
		wantResults map[HashType]string
		wantErr     error
		setupHasher func() *Hasher
	}{
		{
			name:  "single hash type",
			input: bytes.NewReader(testContent),
			ht:    HashMD5,
			wantResults: map[HashType]string{
				HashMD5: hex.EncodeToString(expectedMD5[:]),
			},
			setupHasher: func() *Hasher {
				return NewHasher()
			},
		},
		{
			name:  "multiple hash types",
			input: bytes.NewReader(testContent),
			ht:    HashMD5 | HashSHA1,
			wantResults: map[HashType]string{
				HashMD5:  hex.EncodeToString(expectedMD5[:]),
				HashSHA1: hex.EncodeToString(expectedSHA1[:]),
			},
			setupHasher: func() *Hasher {
				return NewHasher()
			},
		},
		{
			name:  "empty reader",
			input: bytes.NewReader(nil),
			ht:    HashMD5,
			setupHasher: func() *Hasher {
				return NewHasher()
			},
		},
		{
			name:  "custom buffer size",
			input: bytes.NewReader(make([]byte, 5*1024*1024)),
			ht:    HashMD5,
			setupHasher: func() *Hasher {
				return NewHasher(WithBufferSize(128 * 1024))
			},
			wantResults: func() map[HashType]string {
				h := md5.New()
				h.Write(make([]byte, 5*1024*1024))
				return map[HashType]string{
					HashMD5: hex.EncodeToString(h.Sum(nil)),
				}
			}(),
		},
		{
			name:  "context cancellation",
			input: &slowReader{data: make([]byte, 1*1024*1024)},
			ht:    HashMD5,
			setupHasher: func() *Hasher {
				return NewHasher()
			},
			wantErr: context.Canceled,
		},
		{
			name:    "invalid hash type",
			input:   bytes.NewReader(testContent),
			ht:      0,
			wantErr: errors.New("at least one hash type required"),
			setupHasher: func() *Hasher {
				return NewHasher()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			if tt.wantErr == context.Canceled {
				cancelCtx, cancel := context.WithCancel(ctx)
				defer cancel()
				ctx = cancelCtx
				go func() {
					time.Sleep(10 * time.Millisecond)
					cancel()
				}()
			}

			h := tt.setupHasher()

			results, n, err := h.ReaderHash(ctx, tt.input, tt.ht)

			if (err != nil) != (tt.wantErr != nil) {
				t.Fatalf("unexpected error: got %v, want %v", err, tt.wantErr)
			}
			if tt.wantErr != nil && !strings.Contains(err.Error(), tt.wantErr.Error()) {
				t.Fatalf("error mismatch: got %v, want %v", err, tt.wantErr)
			}

			if tt.wantResults != nil {
				if len(results) != len(tt.wantResults) {
					t.Fatalf("result count mismatch: got %d, want %d", len(results), len(tt.wantResults))
				}

				for k, v := range tt.wantResults {
					if got := results[k]; got != v {
						t.Errorf("hash %v mismatch:\n got %s\nwant %s", uint(k), got, v)
					}
				}

				if n != int64(tt.input.(*bytes.Reader).Size()) && tt.wantErr == nil {
					t.Errorf("byte count mismatch: got %d, want %d", n, tt.input.(*bytes.Reader).Size())
				}
			}
		})
	}
}

func TestFileHash(t *testing.T) {
	createTempFile := func(content []byte) string {
		f, err := os.CreateTemp("", "hash-test-")
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { os.Remove(f.Name()) })

		if _, err := f.Write(content); err != nil {
			t.Fatal(err)
		}
		if err := f.Close(); err != nil {
			t.Fatal(err)
		}
		return f.Name()
	}

	tests := []struct {
		name        string
		setupFile   func() string
		ht          HashType
		wantResults map[HashType]string
		wantErr     error
	}{
		{
			name: "valid file",
			setupFile: func() string {
				return createTempFile([]byte("test file content"))
			},
			ht: HashSHA256 | HashSHA512,
			wantResults: func() map[HashType]string {
				content := []byte("test file content")

				h256 := sha256.New()
				h256.Write(content)
				sha256Hash := hex.EncodeToString(h256.Sum(nil))

				h512 := sha512.New()
				h512.Write(content)
				sha512Hash := hex.EncodeToString(h512.Sum(nil))

				return map[HashType]string{
					HashSHA256: sha256Hash,
					HashSHA512: sha512Hash,
				}
			}(),
		},
		{
			name: "non-existent file",
			setupFile: func() string {
				return "/path/to/nonexistent/file"
			},
			ht:      HashMD5,
			wantErr: os.ErrNotExist,
		},
		{
			name: "empty file",
			setupFile: func() string {
				return createTempFile(nil)
			},
			ht: HashMD5,
			wantResults: map[HashType]string{
				HashMD5: hex.EncodeToString(md5.New().Sum(nil)),
			},
		},
		{
			name: "large file",
			setupFile: func() string {
				f := createTempFile(nil)
				if err := os.Truncate(f, 256*1024*1024); err != nil {
					t.Fatal(err)
				}
				return f
			},
			ht: HashMD5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filePath := tt.setupFile()
			h := NewHasher()

			results, n, err := h.FileHash(context.Background(), filePath, tt.ht)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("error mismatch: got %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			for k, v := range tt.wantResults {
				if got := results[k]; got != v {
					t.Errorf("hash %v mismatch:\n got %s\nwant %s", uint(k), got, v)
				}
			}

			if tt.name == "large file" {
				fi, err := os.Stat(filePath)
				if err != nil {
					t.Fatal(err)
				}
				if n != fi.Size() {
					t.Errorf("file size mismatch: got %d, want %d", n, fi.Size())
				}
			}
		})
	}
}

type slowReader struct {
	data []byte
	pos  int
}

func (r *slowReader) Read(p []byte) (n int, err error) {
	if r.pos >= len(r.data) {
		return 0, io.EOF
	}
	time.Sleep(10 * time.Millisecond)
	n = copy(p, r.data[r.pos:r.pos+1])
	r.pos += n
	return n, nil
}

func TestBufferSize(t *testing.T) {
	tests := []struct {
		name     string
		input    int
		expected int
		poolInit bool
	}{
		{"Negative Value", -1, 0, false},
		{"Zero Value", 0, 0, false},
		{"Under Min", 500, minBufferSize, true},
		{"Normal Value", 32 * 1024, 32 * 1024, true},
		{"Over Max", 256 * 1024 * 1024, maxBufferSize, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewHasher(WithBufferSize(tt.input))

			if tt.poolInit != (h.bufPool.New != nil) {
				t.Errorf("pool init mismatch, expect %v got %v",
					tt.poolInit, h.bufPool.New != nil)
			}

			if h.bufSize != tt.expected {
				t.Errorf("buffer size mismatch, expect %d got %d",
					tt.expected, h.bufSize)
			}
		})
	}
}

func TestHashingWithoutBuffer(t *testing.T) {
	h := NewHasher(WithBufferSize(-1))

	data := strings.Repeat("test_data", 1000)
	r := strings.NewReader(data)

	results, n, err := h.ReaderHash(context.Background(), r, HashMD5)

	as.NoError(t, err)
	as.Equal(t, int64(len(data)), n)

	expected := md5.Sum([]byte(data))
	as.Equal(t, hex.EncodeToString(expected[:]), results[HashMD5])
}

func TestBufferReuse(t *testing.T) {
	h := NewHasher()
	data := make([]byte, 64*1024)
	r := bytes.NewReader(data)

	allocs := testing.AllocsPerRun(100, func() {
		r.Seek(0, io.SeekStart)
		_, _, _ = h.ReaderHash(context.Background(), r, HashMD5)
	})

	if allocs > 10 {
		t.Errorf("expected <=10 allocs, got %f", allocs)
	}
}
