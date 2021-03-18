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
		return nil, fmt.Errorf("create form file: %w", err)
	}
	gzipW := gzip.NewWriter(part)
	_, err = io.Copy(gzipW, r)
	if err != nil {
		return nil, fmt.Errorf("upload io.copy: %w", err)
	}

	gzipW.Close()
	writer.Close()

	return createPostRequest(uploadURL, writer, body)
}

func UploadGZIPZeroMemory(uploadURL string, filename string, r io.Reader) (*http.Request, error) {
	pipeR, pipeW := io.Pipe()
	writer := multipart.NewWriter(pipeW)
	gzfilename := fmt.Sprintf("%s.gz", filename)

	go func() {
		var err error
		defer func() {
			if err != nil {
				_ = pipeW.CloseWithError(err)
			} else {
				_ = pipeW.Close()
			}
		}()

		defer pipeW.Close()
		defer writer.Close()
		part, err := writer.CreateFormFile("file", gzfilename)
		if err != nil {
			err = fmt.Errorf("create form file: %w", err)
			return
		}
		gzipW := gzip.NewWriter(part)
		defer gzipW.Close()
		_, err = io.Copy(gzipW, r)
		if err != nil {
			err = fmt.Errorf("UploadGZIPZeroMemory io.copy: %w", err)
			return
		}
	}()

	return createPostRequest(uploadURL, writer, pipeR)
}

func createPostRequest(uploadURL string, w *multipart.Writer, r io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(http.MethodPost, uploadURL, r)
	if err != nil {
		return nil, fmt.Errorf("upload io.copy: %w", err)
	}

	req.Header.Add("Content-Encoding", "gzip")
	req.Header.Add("Content-Type", w.FormDataContentType())
	return req, nil
}
