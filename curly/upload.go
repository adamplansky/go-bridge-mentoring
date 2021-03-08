package main

import (
	"compress/gzip"
	"io"
)

func UploadGZIP(r io.Reader, w io.Writer) error {

	gzipW := gzip.NewWriter(w)
	defer gzipW.Close()

	_, err := io.Copy(gzipW, r)
	return err
}
