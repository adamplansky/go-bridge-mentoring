package main

import (
	"crypto/md5"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
)

const (
	// floppy disk size = 1.44 * 1000 * 1024
	floppySize = 1_474_560
)

var (
	stdout = os.Stdout
	//stdnull = os.NewFile(0, os.DevNull)
	stdnull = io.Discard
	stderr  = os.Stderr
)

type Config struct {
	MD5           bool
	ChunkedPrefix string
	Std           io.Writer
	Url           *url.URL
}

func ParseConfig() (*Config, error) {
	var outputFlag, outputChunked string
	var md5Flag bool
	flag.StringVar(&outputFlag, "output", "", "output is downloaded to file, if value is '-' output is stdout, if output is not specified file is printed to /dev/null")
	flag.StringVar(&outputChunked, "output-chunked", "", "FILEPREFIX, content is splitted to 3.5 Mb files FILEPREFIX.0 FILEPREFIX.1")
	flag.BoolVar(&md5Flag, "md5", false, "prints md5 sum of file into stderr")
	flag.Parse()

	if len(flag.Args()) == 0 {
		return nil, errors.New("no file to download")
	}
	u, err := url.Parse(flag.Args()[0])
	if err != nil {
		return nil, fmt.Errorf("unable parse arg flag: %w", err)
	}

	var std io.Writer
	switch {
	case outputFlag == "-":
		std = stdout
	case len(outputFlag) > 0:
		f, err := os.Create(outputFlag)
		if err != nil {
			return nil, fmt.Errorf("unable to create os file: %w", err)
		}
		std = f
	default:
		std = stdnull
	}

	return &Config{
		MD5:           md5Flag,
		Std:           std,
		Url:           u,
		ChunkedPrefix: outputChunked,
	}, nil
}

func main() {
	cfg, err := ParseConfig()
	if err != nil {
		logErr(err)
	}

	c := &http.Client{}
	req, err := http.NewRequest(http.MethodGet, cfg.Url.String(), nil)
	if err != nil {
		logErr(err)
	}

	resp, err := c.Do(req)
	if err != nil {
		logErr(err)
	}
	defer resp.Body.Close()

	r := io.Reader(resp.Body)
	h := md5.New()

	if len(cfg.ChunkedPrefix) > 0 {
		chunker, err := NewFileChunker(cfg.ChunkedPrefix)
		if err != nil {
			logErr(err)
		}
		defer chunker.Close()
		chunked := NewChunked(chunker, floppySize)


		r = io.TeeReader(resp.Body, chunked)
	}

	if cfg.MD5 {
		r = io.TeeReader(r, h)
	}

	if _, err := io.Copy(cfg.Std, r); err != nil {
		logErr(err)
	}

	if cfg.MD5 {
		_, _ = fmt.Fprintf(stderr, "%x\n", h.Sum(nil))
	}

	os.Exit(0)
}

func logErr(err error) {
	_, _ = stderr.Write([]byte(err.Error()))
	os.Exit(1)
}
