package main

import (
	"bytes"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"testing"
)

var _ Chunker = &testChunked{}

type testChunked struct {
	multiBuffer []*bytes.Buffer
	buf         *bytes.Buffer
	idx         int
}

func NewTestChunked() *testChunked {
	var buf bytes.Buffer
	return &testChunked{
		multiBuffer: []*bytes.Buffer{
			&buf,
		},
		buf: &buf,
		idx: 0,
	}
}

func (_ *testChunked) Close() error {
	return nil
}

func (c *testChunked) Write(p []byte) (n int, err error) {
	return c.buf.Write(p)
}

func (c *testChunked) NewChunk() error {
	c.idx++
	var err error
	var buf bytes.Buffer
	c.buf = &buf
	c.multiBuffer = append(c.multiBuffer, &buf)
	return err
}

func TestChunked_Write(t *testing.T) {
	tests := []struct {
		name    string
		maxSize int
		input   [][]byte
		want    [][]byte
	}{
		{
			name:    "io test",
			maxSize: 5,
			input: [][]byte{
				[]byte("123"),
				[]byte("456"),
				[]byte("789"),
				[]byte("123"),
			},
			want: [][]byte{
				[]byte("12345"),
				[]byte("67891"),
				[]byte("23"),
			},
		},
		{
			name:    "max size 10",
			maxSize: 10,
			input: [][]byte{
				[]byte("123"),
				[]byte("456"),
				[]byte("789"),
				[]byte("123"),
			},
			want: [][]byte{
				[]byte("1234567891"),
				[]byte("23"),
			},
		},
		{
			name:    "max size 3",
			maxSize: 3,
			input: [][]byte{
				[]byte("1234567891234"),
			},
			want: [][]byte{
				[]byte("123"),
				[]byte("456"),
				[]byte("789"),
				[]byte("123"),
				[]byte("4"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tChunker := NewTestChunked()
			c := NewChunked(tChunker, tt.maxSize)

			for _, p := range tt.input {
				_, err := c.Write(p)
				assert.NoError(t, err)
			}

			for i := range tt.want {
				diff := cmp.Diff(tt.want[i], tChunker.multiBuffer[i].Bytes())
				assert.Empty(t, diff)
			}
		})
	}
}
