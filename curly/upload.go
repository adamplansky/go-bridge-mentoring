package main

import (
	"compress/gzip"
	"fmt"
	"io"
	"mime/multipart"
)

func UploadGZIP(wMultiPart *multipart.Writer, r io.Reader, filename string) error {
	part, err := wMultiPart.CreateFormFile("file", fmt.Sprintf("%s.gz", filename))
	if err != nil {
		return err
	}

	gzipW := gzip.NewWriter(part)
	defer gzipW.Close()

	if _, err = io.Copy(gzipW, r); err != nil {
		return err
	}
	return nil
}
