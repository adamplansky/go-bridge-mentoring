package main

import (
	"github.com/stretchr/testify/assert"
	"io"
	"testing"
)

func ioFloppyCreateFn() ([]*testChunked, func() (io.WriteCloser, error)) {
	ww := make([]*testChunked, 0)
	return ww, func() (io.WriteCloser, error) {
		wc := &testChunked{Buf: make([]byte, 0)}
		ww = append(ww, wc)
		return wc, nil
	}
}

var _ io.WriteCloser = &testChunked{}

type testChunked struct {
	Buf []byte
}

func (_ *testChunked) Close() error {
	return nil
}

func (c *testChunked) Write(p []byte) (n int, err error) {
	c.Buf = append(c.Buf, p...)
	return len(p), nil
}

func TestChunked_Write(t *testing.T) {
	tests := []struct {
		name     string
		maxSize  int
		p        [][]byte
		want     [][]byte
	}{
		{
			name:     "io test",
			maxSize:  5,
			p: [][]byte{
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, floppyFn := ioFloppyCreateFn()
			w, err := floppyFn()
			assert.NoError(t, err)
			c := &Chunked{
				w:              w,
				size:           0,
				maxSize:        tt.maxSize,
				floppyCreateFn: floppyFn,
			}
			for _, p := range tt.p {
				_, err = c.Write(p)
			}

			assert.NoError(t, err)
			for i := range result {
				assert.Equal(t, tt.want[i], result[i].Buf)
			}
		})
	}
}
