package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
)

const (
	Megabyte = 1 << 20
	Kilobyte = 1 << 10
)

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	f, err := loadFile("pic.png")
	check(err)

	fi, err := f.Stat()
	check(err)
	fmt.Println(fi.Size())

	var kilobytes int64
	kilobytes = (fi.Size() / 1024)
	fmt.Println("size in kb: ", kilobytes, " lr size: ", Kilobyte)

	r := io.MultiReader(bytes.NewReader([]byte("concat bytes + file reader: ")), f)

	var buf1, buf2 bytes.Buffer
	r = io.TeeReader(r, &buf1)
	r = io.TeeReader(r, &buf2)
	w := bufio.NewWriter(os.Stderr)
	r = io.TeeReader(r, w)

	lr := io.LimitReader(r, Kilobyte)

	n, err := io.Copy(os.Stderr, lr)
	fmt.Println("io copy output: ", n, err)


	//fmt.Println("")
	//log.Printf("buf1: %s<---\n", buf1.String())
	//log.Printf("buf2: %s<---\n", buf2.String())
	check(w.Flush())
	fmt.Printf("\nbufio len: %d\n", w.Size())
	fmt.Println("io copy output: ", n, err)

	//fmt.Println(buf1.Len())
	//fmt.Println(len("concat bytes + file reader: hello"))

}

func loadFile(filename string) (*os.File, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	return f, nil
}
