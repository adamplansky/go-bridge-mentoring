package rest

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

func (h *Handlers) Upload(w http.ResponseWriter, r *http.Request) {
	srcFile, _, err := r.FormFile("file")
	if err != nil {
		h.log.Error("failed to acquire the uploaded content", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		writeError(w, err)
		return
	}
	defer srcFile.Close()
	size, err := getSize(srcFile)
	if err != nil {
		h.log.Error("failed to get the size of the uploaded content", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		writeError(w, err)
		return
	}

	body, err := ioutil.ReadAll(srcFile)
	if err != nil {
		h.log.Error("failed to read the uploaded content", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		writeError(w, err)
		return
	}
	filename := uuid.New().String()

	dstPath := path.Join(h.uploadFolder, filename)
	dstFile, err := os.OpenFile(dstPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		h.log.Error("failed to open the file", zap.Error(err), zap.String("path", dstPath))
		w.WriteHeader(http.StatusInternalServerError)
		writeError(w, err)
		return
	}
	defer dstFile.Close()
	if written, err := dstFile.Write(body); err != nil {
		h.log.Error("failed to write the content", zap.Error(err), zap.String("path", dstPath))
		w.WriteHeader(http.StatusInternalServerError)
		writeError(w, err)
		return
	} else if int64(written) != size {
		h.log.Error("uploaded file size and written size differ", zap.Int64("size", size), zap.Int("written", written))
		w.WriteHeader(http.StatusInternalServerError)
		writeError(w, fmt.Errorf("the size of uploaded content is %d, but %d bytes written", size, written))
	}
	uploadedURL := strings.TrimPrefix(dstPath, h.uploadFolder)
	if !strings.HasPrefix(uploadedURL, "/") {
		uploadedURL = "/" + uploadedURL
	}
	uploadedURL = "/files" + uploadedURL
	h.log.Info("file uploaded by POST",
		zap.String("path", dstPath),
		zap.String("url", uploadedURL),
		zap.Int64("size", size),
	)

	w.WriteHeader(http.StatusOK)
	writeSuccess(w, uploadedURL)
}
