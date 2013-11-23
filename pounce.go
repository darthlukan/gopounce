// pounce.go
package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// download takes a url of type string as an argument
// and returns a response pointer.
func download(url string) *http.Response {

	fmt.Printf("Downloading...\n")
	resp, err := http.Get(url)

	if err != nil {
		panic(err)
	}

	fmt.Printf("Download complete!\n")

	return resp
}

// create takes a filename of type string as an argument
// and returns a file pointer.  It basically calls
// os.Create(filename) and checks err.
func create(filename string) *os.File {

	fmt.Printf("Creating %v ...\n", filename)
	file, err := os.Create(filename)

	if err != nil {
		panic(err)
	}

	fmt.Printf("%v created!\n", filename)

	return file
}

// save takes a file of type io.Writer and a response pointer
// as an argument so that it can return a size as type int64.
// It calls io.Copy, handles any errors, and returns.
func save(file io.Writer, resp *http.Response) int64 {

	fmt.Printf("Writing file...\n")
	complete, err := io.Copy(file, resp.Body)

	if err != nil {
		panic(err)
	}

	return complete
}

func main() {

	var url string
	var filename string

	flag.StringVar(&url, "url", "", "URL to get")
	flag.StringVar(&filename, "filename", "", "Destination file to write.")
	flag.Parse()
	flag.Args()

	if url == "" {
		url = flag.Arg(0)
	}

	if filename == "" {
		filename = flag.Arg(1)
	}
	start := time.Now()
	resp := download(url)
	defer resp.Body.Close()

	file := create(filename)
	defer file.Close()

	complete := save(file, resp)
	end := time.Now()
	size := complete / 1024
	fmt.Printf("Downloaded ~%v kb in %v \n", size, end.Sub(start))
}
