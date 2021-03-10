package upload

import (
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
)

type Uploader struct {
	c         *http.Client
	uploadURL string
}

func New(c *http.Client, uploadURL string) *Uploader {
	return &Uploader{
		c:         c,
		uploadURL: uploadURL,
	}
}

func (u *Uploader) Upload(r io.Reader, filename string) (*http.Request, error) {

	rBody, wPipe := io.Pipe()
	writer := multipart.NewWriter(wPipe)
	go func() {
		defer wPipe.Close()
		defer writer.Close()

		gzfilename := fmt.Sprintf("%s.gz", filename)
		part, err := writer.CreateFormFile("file", gzfilename)

		gzipW := gzip.NewWriter(part)
		defer gzipW.Close()

		_, err = io.Copy(gzipW, r)
		if err != nil {
			log.Fatal(err)
		}
	}()

	req, err := http.NewRequest(http.MethodPost, u.uploadURL, rBody)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Encoding", "gzip")
	req.Header.Add("Content-Type", writer.FormDataContentType())

	return req, nil
}
