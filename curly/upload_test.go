package main

import (
	"bytes"
	b64 "encoding/base64"
	"github.com/stretchr/testify/assert"
	"io"
	"strings"
	"testing"
)

func TestUpload(t *testing.T) {
	tests := []struct {
		name string
		r    io.Reader
		want string
	}{
		{
			name: "hello",
			r:    strings.NewReader("hello"),
			want: "H4sIAAAAAAAA/8pIzcnJBwQAAP//hqYQNgUAAAA=",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := UploadGZIP(tt.r, &buf)
			assert.NoError(t, err)
			sEnc := b64.StdEncoding.EncodeToString(buf.Bytes())
			assert.Equal(t, tt.want, sEnc)
		})
	}
}
