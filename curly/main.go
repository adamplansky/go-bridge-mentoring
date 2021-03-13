package main

import (
	"context"
	"crypto/md5"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"time"

	"go.uber.org/zap"

	"github.com/adamplansky/go-bridge-mentoring/curly/roundtripper"
	"github.com/adamplansky/go-bridge-mentoring/curly/upload"
)

const (
	// floppy disk size = 1.44 * 1000 * 1024
	floppySize = 1_474_560
)

var (
	stdout  = os.Stdout
	stdnull = io.Discard
	stderr  = os.Stderr
)

type Config struct {
	MD5           bool
	ChunkedPrefix string
	Std           io.Writer
	DownloadURL   *url.URL
	Upload        bool
	UploadURL     string
	Verbose       bool
	errs          []error
}

// https://github.com/mayth/go-simple-upload-server
func ParseConfig(log *zap.SugaredLogger) *Config {

	var cfg Config

	//var outputFlag, outputChunked string

	flag.Func("output", "output is downloaded to file, if value is '-' output is stdout, if output is not specified file is printed to /dev/null", func(outputFlag string) error {
		switch {
		case outputFlag == "-":
			cfg.Std = stdout
		case len(outputFlag) > 0:
			f, err := os.Create(outputFlag)
			if err != nil {
				log.Fatal("unable to create os file: %w", err)
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
	flag.StringVar(&cfg.UploadURL, "uploadurl", "", "upload url")
	flag.BoolVar(&cfg.Verbose, "verbose", false, "verbose output")

	flag.Parse()

	if len(flag.Args()) == 0 {
		log.Fatal("no file to download")
	}
	var err error
	cfg.DownloadURL, err = url.Parse(flag.Args()[0])
	if err != nil {
		log.Fatal("unable parse arg flag: %w", err)
	}

	if cfg.Upload && cfg.UploadURL == "" {
		log.Fatal("no upload url specified")
	}

	return &cfg
}

func main() {
	logger, _ := zap.NewDevelopment()
	log := logger.Sugar()
	defer logger.Sync() // flushes buffer, if any

	cfg := ParseConfig(log)

	t := http.DefaultTransport
	if cfg.Verbose {
		t = roundtripper.NewDebug(t, logger.Sugar())
	}

	c := &http.Client{
		Transport: t,
		Timeout:   10 * time.Second,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, cfg.DownloadURL.String(), nil)
	if err != nil {
		logger.Fatal(err.Error())
	}

	resp, err := c.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	r := io.Reader(resp.Body)
	h := md5.New()

	if len(cfg.ChunkedPrefix) > 0 {
		chunker, err := NewFileChunker(cfg.ChunkedPrefix)
		if err != nil {
			log.Fatal(err)
		}
		defer chunker.Close()
		chunked := NewChunked(chunker, floppySize)
		r = io.TeeReader(resp.Body, chunked)
	}

	if cfg.MD5 {
		r = io.TeeReader(r, h)
	}

	if cfg.Upload {
		u := upload.New(c, cfg.UploadURL)
		n := path.Base(cfg.DownloadURL.Path)

		req, err := u.Upload(r, n)
		if err != nil {
			log.Fatal(err)
		}

		resp, err := c.Do(req)
		if err != nil {
			log.Fatal(err)
		}

		defer resp.Body.Close()
	}

	if _, err := io.Copy(cfg.Std, r); err != nil {
		log.Fatal(err)
	}

	if cfg.MD5 {
		msg := fmt.Sprintf("MD5 sum: %x", h.Sum(nil))
		log.Errorw(msg)
	}

	os.Exit(0)
}
