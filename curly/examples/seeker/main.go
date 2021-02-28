package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

func main() {
	s := strings.NewReader("some_io.Reader_stream_to_be_read")
	//s := io.NewSectionReader(r, 5, int64(r.Len()))

	if _, err := io.Copy(os.Stdout, s); err != nil {
		log.Fatal(err)
	}
	fmt.Println("\n--")

	if _, err := s.Seek(3, io.SeekStart); err != nil {
		log.Fatal(err)
	}

	if _, err := io.Copy(os.Stdout, s); err != nil {
		log.Fatal(err)
	}

	fmt.Println("\n--")

	if _, err := s.Seek(-5, io.SeekEnd); err != nil {
		log.Fatal(err)
	}

	if _, err := io.Copy(os.Stdout, s); err != nil {
		log.Fatal(err)
	}

}
