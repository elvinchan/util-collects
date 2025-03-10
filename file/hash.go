package file

import (
	"context"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"errors"
	"hash"
	"io"
	"os"
	"sync"
)

type HashType uint

const (
	HashMD5 HashType = 1 << iota
	HashSHA1
	HashSHA256
	HashSHA512
)

type HashMeta struct {
	Type     HashType
	HashFunc func() hash.Hash
}

// HashOption defines configuration options for hasher
type HashOption func(*hashOptions)

type hashOptions struct {
	bufSize uint64
	meta    []HashMeta
}

const minBufferSize = 1 * 1024          // 1KB
const maxBufferSize = 128 * 1024 * 1024 // 128MB

func WithBufferSize(n uint64) HashOption {
	return func(o *hashOptions) {
		if n < minBufferSize {
			n = minBufferSize
		}
		if n > maxBufferSize {
			n = maxBufferSize
		}
		o.bufSize = n
	}
}

func WithHashMeta(m HashMeta) HashOption {
	return func(o *hashOptions) {
		for i, existing := range o.meta {
			if existing.Type == m.Type {
				o.meta[i] = m
				return
			}
		}
		o.meta = append(o.meta, m)
	}
}

type Hasher struct {
	bufPool sync.Pool
	bufSize uint64
	meta    []HashMeta
}

func NewHasher(opts ...HashOption) *Hasher {
	o := hashOptions{
		bufSize: 32 * 1024,
		meta: []HashMeta{
			{HashMD5, md5.New},
			{HashSHA1, sha1.New},
			{HashSHA256, sha256.New},
			{HashSHA512, sha512.New},
		},
	}
	for _, opt := range opts {
		opt(&o)
	}
	return &Hasher{
		bufPool: sync.Pool{
			New: func() interface{} {
				buf := make([]byte, o.bufSize)
				return &buf
			},
		},
		bufSize: o.bufSize,
		meta:    o.meta,
	}
}

// ReaderHash returns a map of hash types to their hex encoded hash values.
func (h *Hasher) ReaderHash(ctx context.Context, r io.Reader, ht HashType) (
	map[HashType]string, int64, error) {
	return readerHash(ctx, r, ht, h)
}

// FileHash returns a map of hash types to their hex encoded hash values.
func (h *Hasher) FileHash(ctx context.Context, path string, ht HashType) (
	map[HashType]string, int64, error) {
	return fileHash(ctx, path, ht, h)
}

var defaultHasher = NewHasher()

// ReaderHash returns a map of hash types to their hex encoded hash values.
func ReaderHash(ctx context.Context, r io.Reader, ht HashType) (
	map[HashType]string, int64, error) {
	return readerHash(ctx, r, ht, defaultHasher)
}

// FileHash returns a map of hash types to their hex encoded hash values.
func FileHash(ctx context.Context, path string, ht HashType) (
	map[HashType]string, int64, error) {
	return fileHash(ctx, path, ht, defaultHasher)
}

func readerHash(ctx context.Context, r io.Reader, ht HashType, h *Hasher) (
	map[HashType]string, int64, error) {
	if ht == 0 {
		return nil, 0, errors.New("at least one hash type required")
	}

	hashers := make(map[HashType]hash.Hash)
	for _, meta := range h.meta {
		if ht&meta.Type != 0 {
			hashers[meta.Type] = meta.HashFunc()
		}
	}
	if len(hashers) == 0 {
		return nil, 0, errors.New("unsupported hash type")
	}
	writers := make([]io.Writer, 0, len(hashers))
	for t := range hashers {
		writers = append(writers, hashers[t])
	}
	multiWriter := io.MultiWriter(writers...)

	cr := &chunkReader{
		r:   r,
		ctx: ctx,
	}

	bufPtr := h.bufPool.Get().(*[]byte)
	buf := *bufPtr
	defer h.bufPool.Put(bufPtr)

	n, err := io.CopyBuffer(multiWriter, cr, buf[:h.bufSize])
	if err != nil {
		return nil, n, err
	}

	results := make(map[HashType]string)
	for t, h := range hashers {
		results[t] = hex.EncodeToString(h.Sum(nil))
	}
	return results, n, nil
}

func fileHash(ctx context.Context, path string, ht HashType, h *Hasher) (
	map[HashType]string, int64, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, 0, err
	}
	defer f.Close()

	return readerHash(ctx, f, ht, h)
}

type chunkReader struct {
	r   io.Reader
	ctx context.Context
}

func (c *chunkReader) Read(p []byte) (n int, err error) {
	if err := c.ctx.Err(); err != nil {
		return 0, err
	}
	n, err = c.r.Read(p)
	return
}
