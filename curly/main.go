package main

import (
	"context"
	"crypto/md5"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"time"

	"github.com/adamplansky/go-bridge-mentoring/curly/request"

	"go.uber.org/zap"

	"github.com/adamplansky/go-bridge-mentoring/curly/roundtripper"
)

const (
	// floppy disk size = 1.44 * 1000 * 1024
	floppySize = 1_474_560
)

var (
	stdout  = os.Stdout
	stdnull = io.Discard
)

type Config struct {
	MD5           bool
	ChunkedPrefix string
	Std           io.Writer
	DownloadURL   *url.URL
	Upload        bool
	UploadURL     *url.URL
	Verbose       bool
}

// https://github.com/mayth/go-simple-upload-server
func ParseConfig(log *zap.SugaredLogger) (*Config, error) {

	cfg := Config{
		// default value to prevent panic nil std.writer
		Std: stdnull,
	}
	flag.Func("output", "output is downloaded to file, if value is '-' output is stdout, if output is not specified file is printed to /dev/null", func(outputFlag string) error {
		switch {
		case outputFlag == "-":
			cfg.Std = stdout
		case len(outputFlag) > 0:
			f, err := os.Create(outputFlag)
			if err != nil {
				return fmt.Errorf("unable to create os file: %w", err)
			}
			cfg.Std = f
		default:
			cfg.Std = stdnull
		}
		return nil
	})
	flag.StringVar(&cfg.ChunkedPrefix, "output-chunked", "", "FILEPREFIX, content is splitted to 3.5 Mb files FILEPREFIX.0 FILEPREFIX.1")
	flag.BoolVar(&cfg.MD5, "md5", false, "prints md5 sum of file into stderr")
	flag.BoolVar(&cfg.Upload, "upload", false, "upload file true/false")
	flag.Func("uploadurl", "upload url", func(uploadURL string) error {
		u, err := url.Parse(uploadURL)
		if err != nil {
			return fmt.Errorf("unable to parse upload url: %w", err)
		}
		cfg.UploadURL = u
		return nil
	})
	flag.BoolVar(&cfg.Verbose, "verbose", false, "verbose output")

	flag.Parse()

	if len(flag.Args()) == 0 {
		return nil, fmt.Errorf("no file to download")
	}
	var err error
	cfg.DownloadURL, err = url.Parse(flag.Args()[0])
	if err != nil {
		return nil, fmt.Errorf("unable parse arg flag: %w", err)

	}

	if cfg.Upload && cfg.UploadURL == nil {
		return nil, fmt.Errorf("no upload url specified")
	}

	return &cfg, nil
}

func run() error {
	logger, _ := zap.NewDevelopment()
	log := logger.Sugar()
	defer logger.Sync() // flushes buffer, if any

	cfg, err := ParseConfig(log)
	if err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}

	t := http.DefaultTransport
	if cfg.Verbose {
		t = roundtripper.NewDebug(t, logger.Sugar())
	}

	c := &http.Client{
		Transport: t,
		Timeout:   10 * time.Second,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, cfg.DownloadURL.String(), nil)
	if err != nil {
		return err
	}

	resp, err := c.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	r := io.Reader(resp.Body)

	if len(cfg.ChunkedPrefix) > 0 {
		chunker, err := NewFileChunker(cfg.ChunkedPrefix)
		if err != nil {
			return err
		}
		defer chunker.Close()
		chunked := NewChunked(chunker, floppySize)
		r = io.TeeReader(resp.Body, chunked)
	}

	h := md5.New()
	if cfg.MD5 {
		r = io.TeeReader(r, h)
	}

	if cfg.Upload {
		fname := path.Base(cfg.DownloadURL.Path)
		req, err := request.UploadGZIPZeroMemory(cfg.UploadURL.String(), fname, r)
		if err != nil {
			return fmt.Errorf("unable to create UploadGZIPZeroMemory request: %w", err)

		}

		_, err = c.Do(req)
		if err != nil {
			return fmt.Errorf("upload Do failed: %w", err)

		}
	}

	if _, err := io.Copy(cfg.Std, r); err != nil {
		return fmt.Errorf("io.Copy failed: %w", err)
	}
	log.Debugf("download has finished successfuly: %s", cfg.DownloadURL)

	if cfg.MD5 {
		msg := fmt.Sprintf("MD5 sum: %x", h.Sum(nil))
		log.Errorw(msg)
	}
	return nil
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
	os.Exit(0)
}
