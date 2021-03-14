package request

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
)

//type Uploader struct {
//	c         *http.Client
//	uploadURL string
//}
//
//func New(c *http.Client, uploadURL string) (*Uploader, error) {
//	if c == nil {
//		return nil, fmt.Errorf("client is nil")
//	}
//	if uploadURL == "" {
//		return nil, fmt.Errorf("uploadURL is empty")
//	}
//	return &Uploader{
//		c:         c,
//		uploadURL: uploadURL,
//	}, nil
//}

func UploadGZIP(uploadURL string, filename string, r io.Reader) (*http.Request, error) {

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)

	gzfilename := fmt.Sprintf("%s.gz", filename)
	part, err := writer.CreateFormFile("file", gzfilename)
	if err != nil {
		return nil, err
	}
	gzipW := gzip.NewWriter(part)

	_, err = io.Copy(gzipW, r)
	if err != nil {
		return nil, fmt.Errorf("upload io.copy: %w", err)
	}

	gzipW.Close()
	writer.Close()

	req, err := http.NewRequest(http.MethodPost, uploadURL, body)
	if err != nil {
		return nil, fmt.Errorf("upload io.copy: %w", err)
	}

	req.Header.Add("Content-Encoding", "gzip")
	req.Header.Add("Content-Type", writer.FormDataContentType())

	return req, nil
}
