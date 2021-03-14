package upload

import (
	"compress/gzip"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
)

type Uploader struct {
	c         *http.Client
	uploadURL *url.URL
}

func New(c *http.Client, uploadURL *url.URL) (*Uploader, error) {
	if c == nil {
		return nil, fmt.Errorf("client is nil")
	}
	if uploadURL == nil {
		return nil, fmt.Errorf("uploadURL is nil")
	}
	return &Uploader{
		c:         c,
		uploadURL: uploadURL,
	}, nil
}

func (u *Uploader) Upload(r io.Reader, filename string) (*http.Request, error, io.Reader) {
	rBody, wPipe := io.Pipe()
	r = io.TeeReader(r, wPipe)
	writer := multipart.NewWriter(wPipe)

	go func() {
		defer wPipe.Close()
		defer writer.Close()

		gzfilename := fmt.Sprintf("%s.gz", filename)
		part, err := writer.CreateFormFile("file", gzfilename)
		if err != nil {
			_ = wPipe.CloseWithError(err)
		}
		gzipW := gzip.NewWriter(part)
		defer gzipW.Close()

		_, err = io.Copy(gzipW, r)
		if err != nil {
			_ = wPipe.CloseWithError(fmt.Errorf("upload io.copy: %w", err))
		}
	}()

	req, err := http.NewRequest(http.MethodPost, u.uploadURL.String(), rBody)
	if err != nil {
		return nil, fmt.Errorf("upload io.copy: %w", err), r
	}

	req.Header.Add("Content-Encoding", "gzip")
	req.Header.Add("Content-Type", writer.FormDataContentType())

	return req, nil, r
}
