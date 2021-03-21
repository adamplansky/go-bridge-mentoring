package main

import (
	"fmt"
	"io"
	"log"
	"os"
)

func main() {

	// pipe writer to reader
	r, w := io.Pipe()
	_, _ = r, w
	go func() {
		fmt.Fprint(w, "some io.Reader stream to be read\n")
		w.Close()
	}()

	if _, err := io.Copy(os.Stdout, r); err != nil {
		log.Fatal("2: ", err)
	}
}
