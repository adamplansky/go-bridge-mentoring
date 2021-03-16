package request

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestUpload(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		filename string
		want     string
	}{
		{
			name:     "hello",
			input:    "hello1\nhello2\nhello3\nhello4\nhello5\nhello6",
			filename: "my-name",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				err := r.ParseMultipartForm(10) // limit your max input length!
				assert.NoError(t, err)

				file, header, err := r.FormFile("file")
				assert.NoError(t, err)
				defer file.Close()
				name := strings.Split(header.Filename, ".")
				fmt.Printf("File name %s\n", name[0])

				var bufOriginal bytes.Buffer
				tr := io.TeeReader(file, &bufOriginal)

				var buf bytes.Buffer
				zr, err := gzip.NewReader(tr)
				assert.NoError(t, err)

				_, err = io.Copy(&buf, zr)
				assert.NoError(t, err)

				assert.NotEqual(t, buf, bufOriginal)
				assert.Equal(t, tt.input, buf.String())
			}))
			defer ts.Close()

			c := &http.Client{
				Timeout: 10 * time.Second,
			}

			r := strings.NewReader(tt.input)
			req, err := UploadGZIP(ts.URL, tt.filename, r)
			assert.NoError(t, err)

			resp, err := c.Do(req)
			assert.NoError(t, err)
			defer resp.Body.Close()

		})
	}
}

func BenchmarkUploadAlloc(b *testing.B) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseMultipartForm(5) // limit your max input length!
		assert.NoError(b, err)

		file, header, err := r.FormFile("file")
		assert.NoError(b, err)
		defer file.Close()

		name := strings.Split(header.Filename, ".")
		fmt.Printf("File name %s\n", name[0])

		fmt.Println(io.ReadAll(r.Body))

		_, err = io.Copy(w, file)
		assert.NoError(b, err)
	}))
	defer ts.Close()

	c := &http.Client{Timeout: 10 * time.Second}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	lr := io.LimitReader(r, 1024)
	req, err := UploadGZIP(ts.URL, "some-name", lr)
	assert.NoError(b, err)

	resp, err := c.Do(req)
	assert.NoError(b, err)
	defer resp.Body.Close()

	_, err = io.Copy(os.Stdout, resp.Body)
	assert.NoError(b, err)
	fmt.Println()

}
