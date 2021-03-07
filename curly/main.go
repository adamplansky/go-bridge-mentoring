package main

import (
	"crypto/md5"
	"curly/rounttripper"
	"errors"
	"flag"
	"fmt"
	"go.uber.org/multierr"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"time"
)

const (
	//megabyte    = 1 << 20
	floppySize = 1_474_560 // floppy disk size
	// 1.44 * 1000 * 1024
)

var (
	stdout = os.Stdout
	//stdnull = os.NewFile(0, os.DevNull)
	stdnull = io.Discard
	stderr  = os.Stderr
)

type Config struct {
	MD5                   bool
	ChunkedPrefix         string
	Std                   io.Writer
	DownloadURL           *url.URL
	Upload                bool
	UploadFile, UploadURL string
	Verbose               bool
	errs                  []error
}

// https://github.com/mayth/go-simple-upload-server
func ParseConfig() *Config {
	var cfg Config

	//var outputFlag, outputChunked string

	flag.Func("output", "output is downloaded to file, if value is '-' output is stdout, if output is not specified file is printed to /dev/null", func(outputFlag string) error {
		var err error
		cfg.DownloadURL, err = url.Parse(flag.Args()[0])
		if err != nil {
			log.Fatal("unable parse arg flag: %w", err)
		}

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
	flag.StringVar(&cfg.UploadFile, "uploadfile", "", "upload file")
	flag.StringVar(&cfg.UploadURL, "uploadurl", "", "upload url")
	flag.BoolVar(&cfg.Verbose, "verbose", false, "verbose output")

	flag.Parse()

	if !cfg.Upload && len(flag.Args()) == 0 {
		log.Fatal("no file to download")
	}

	return &cfg
}

func (cfg *Config) ValidateUpload() error {
	if cfg.UploadURL == "" {
		cfg.errs = append(cfg.errs, errors.New("uploadurl not specified"))
	}
	if cfg.UploadFile == "" {
		cfg.errs = append(cfg.errs, errors.New("uploadfile not specified"))
	}
	return multierr.Combine(cfg.errs...)
}

func main() {
	cfg := ParseConfig()
	if cfg.Upload {
		err := cfg.ValidateUpload()
		if err != nil {
			log.Fatal(err)
		}

		file, err := os.Open(cfg.UploadFile)
		if err != nil {
			log.Fatal(err)
		}

		rBody, wPipe := io.Pipe()
		//defer wPipe.Close()

		wMultiPart := multipart.NewWriter(wPipe)

		go func() {
			defer wPipe.Close()
			defer wMultiPart.Close()

			if err = UploadGZIP(wMultiPart, file, cfg.UploadFile); err != nil {
				log.Fatal(err)
			}
		}()

		c := &http.Client{
			Transport: rounttripper.New(),
			Timeout:   10 * time.Second,
		}
		req, err := http.NewRequest(http.MethodPost, cfg.UploadURL, rBody)
		if err != nil {
			log.Fatal(err)
		}
		req.Header.Set("Content-Type", wMultiPart.FormDataContentType())
		req.Header.Add("Content-Encoding", "gzip")


		if cfg.Verbose {
			log.Printf("")
		}

		resp, err := c.Do(req)
		if err != nil {
			log.Fatal(err)
		}

		defer resp.Body.Close()
		log.Printf("response upload: %#v\n", resp)

		os.Exit(0)
	}

	c := &http.Client{}
	req, err := http.NewRequest(http.MethodGet, cfg.DownloadURL.String(), nil)
	if err != nil {
		log.Fatal(err)
	}

	resp, err := c.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	r := io.Reader(resp.Body)
	h := md5.New()

	if len(cfg.ChunkedPrefix) > 0 {
		chunked, err := NewChunked(cfg.ChunkedPrefix)
		if err != nil {
			log.Fatal(err)
		}
		r = io.TeeReader(resp.Body, chunked)
	}

	if cfg.MD5 {
		r = io.TeeReader(r, h)
	}

	if _, err := io.Copy(cfg.Std, r); err != nil {
		log.Fatal(err)
	}

	if cfg.MD5 {
		_, _ = fmt.Fprintf(stderr, "%x\n", h.Sum(nil))
	}

	os.Exit(0)
}

var _ io.Writer = (*Chunked)(nil)

type Chunked struct {
	w              io.WriteCloser
	size           int
	maxSize        int
	floppyCreateFn func() (io.WriteCloser, error)
}

func fileFloppyFn(p string) func() (io.WriteCloser, error) {
	i := -1
	prefix := p
	return func() (io.WriteCloser, error) {
		i++
		return os.Create(fmt.Sprintf("%s.%d", prefix, i))
	}
}

func NewChunked(prefix string) (*Chunked, error) {
	floppyFunc := fileFloppyFn(prefix)
	w, err := floppyFunc()
	return &Chunked{
		w:              w,
		size:           0,
		maxSize:        floppySize,
		floppyCreateFn: floppyFunc,
	}, err
}

func (c *Chunked) Write(p []byte) (int, error) {
	if len(p) < c.maxSize-c.size {
		n, err := c.w.Write(p)
		if err != nil {
			return 0, err
		}
		c.size += n
		return n, nil
	}

	off := c.maxSize - c.size
	n, err := c.w.Write(p[:off])
	if err != nil {
		return n, err
	}
	c.w, err = c.floppyCreateFn()
	if err != nil {
		return n, err
	}
	c.size = len(p) - off
	n, err = c.w.Write(p[off:])
	if err != nil {
		return n, err
	}
	return n, nil

}
