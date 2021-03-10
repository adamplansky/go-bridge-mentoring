package main

import (
	"fmt"
	"io"
	"os"
)

var _ io.Writer = (*Chunked)(nil)

type Chunked struct {
	chunker Chunker
	size    int
	maxSize int
}

func NewChunked(chunker Chunker, chunkSize int) io.Writer {
	return &Chunked{
		chunker: chunker,
		size:    0,
		maxSize: chunkSize,
	}
}

func (c *Chunked) Write(p []byte) (int, error) {
	if len(p) < c.maxSize-c.size {
		n, err := c.chunker.Write(p)
		if err != nil {
			return 0, err
		}
		c.size += n
		return n, nil
	}

	off := c.maxSize - c.size
	n, err := c.chunker.Write(p[:off])
	if err != nil {
		return n, err
	}
	err = c.chunker.NewChunk()
	if err != nil {
		return n, err
	}
	c.size = 0
	return c.Write(p[off:])
}

type Chunker interface {
	io.WriteCloser
	NewChunk() error
}

type fileChunker struct {
	file   *os.File
	prefix string
	idx    int
}

func NewFileChunker(prefix string) (Chunker, error) {
	chunker := fileChunker{
		prefix: prefix,
		idx:    0,
	}
	var err error
	chunker.file, err = os.Open(filename(chunker.prefix, chunker.idx))
	return &chunker, err
}

func (f *fileChunker) NewChunk() error {
	f.idx++
	var err error
	if err = f.file.Close(); err != nil {
		return err
	}
	f.file, err = os.Create(filename(f.prefix, f.idx))
	return err
}

func (f *fileChunker) Write(p []byte) (int, error) {
	return f.file.Write(p)
}

func (f *fileChunker) Close() error {
	return f.file.Close()
}

func filename(prefix string, idx int) string {
	return fmt.Sprintf("%s.%d", prefix, idx)
}
