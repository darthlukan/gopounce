/* pounce.go
	Contains all of the functions necessary to download
	a remote file given a URL and a location on disk to
	save that file.  Can also take a text file containing
	a URL on each line and save each file to a specified
	directory on disk.  Downloading via file input uses
	a very simplistic method of finding filenames, so
	URLs should contain the filename and extension in the
	URL.

	Single File Example:
		gopounce http://www.google.com/index.html /tmp/index.html
	
	File Input Example:
		gopounce /path/to/file.txt /tmp
	
		<file.txt contents>
			http://www.google.com/index.html
			http://www.brianctomlinson.com/index.html
		</>

	NOTES:
	1. This software has not been tested on platforms other than Linux.
	2. The '.txt' extension on the input file is not a necessity.
	3. Feel free to do whatever you'd like with this program, I wrote
	it to help me learn Go, not to sell nor to solve some awesome problem. :)
*/
package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/guelfey/go.dbus"
	"io"
	"net/http"
	"os"
	"strings"
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

	file, err := os.Create(filename)

	if err != nil {
		panic(err)
	}

	fmt.Printf("Created %v ...\n", filename)

	return file
}

func multiDownload(url, destination string, channel chan<- bool) {

	fmt.Printf("Downloading...\n")
	resp, err := http.Get(url)
	createDone := make(chan bool)
	go multiCreate(destination, resp, createDone)

	if err != nil {
		panic(err)
	}
	fmt.Printf("Download complete!\n")
	channel <- true
	<-createDone
}

func multiCreate(destination string, response *http.Response, createDone chan<- bool) {

	fmt.Printf("Creating %v ...\n", destination)
	file, err := os.Create(destination)
	defer file.Close()

	if err != nil {
		panic(err)
	}

	fmt.Printf("Writing file %v...\n", destination)
	complete, err := io.Copy(file, response.Body)
	defer response.Body.Close()
	fmt.Printf("Wrote %v bytes to file %v\n", complete, destination)

	if err != nil {
		fmt.Println(err)
	}

	createDone <- true
}

func readFile(infilename, destination string) {

	file, err := os.Open(infilename)

	if err != nil {
		panic(err)
	}

	scanner := bufio.NewScanner(file)

	lineCount := 0
	channel := make(chan bool)
	for scanner.Scan() {
		lineCount++

		line := scanner.Text()
		splitCount := strings.Count(line, "/")
		endOfSplit := strings.SplitAfterN(line, "/", splitCount)
		filename := string(endOfSplit[len(endOfSplit)-1])

		if strings.Count(filename, "/") > 0 {
			filename = string(strings.SplitAfterN(filename, "/", 2)[1])
		}

		toSave := fmt.Sprintf("%v%v", destination, filename)
		go multiDownload(line, toSave, channel)
	}

	if err := scanner.Err(); err != nil {
		fmt.Println(err)
	}

	for i := 0; i < lineCount; i++ {
		<-channel
	}
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

// notify takes a message string and channel as an
// argument and displays a dbus notification.
func notify(msg string, done chan<- bool) {

	conn, err := dbus.SessionBus()

	if err != nil {
		fmt.Println(err)
	}

	obj := conn.Object("org.freedesktop.Notifications", "/org/freedesktop/Notifications")
	call := obj.Call("org.freedesktop.Notifications.Notify", 0, "", uint32(0),
		"", "GoPounce", msg, []string{}, map[string]dbus.Variant{}, int32(5000))

	if call.Err != nil {
		fmt.Println(call.Err)
	}
	done <- true
}

func main() {

	var input string
	var output string
	var msg string

	done := make(chan bool)

	flag.StringVar(&input, "url", "", "URL to get")
	flag.StringVar(&input, "infile", "", "An input file with URLs on each line that are newline delimitted.")
	flag.StringVar(&output, "filename", "", "Destination file to write.")
	flag.StringVar(&output, "directory", "", "Destination directory to write into with trailing '/' (only works with -infile flag)")
	flag.Parse()
	flag.Args()

	if input == "" {
		input = flag.Arg(0)
	}

	if output == "" {
		output = flag.Arg(1)
	}

	start := time.Now()

	// TODO: Refactor, this doesn't conform to DRY.
	if strings.Contains(input, "http") || strings.Contains(input, "www") {
		resp := download(input)
		defer resp.Body.Close()
		file := create(output)
		defer file.Close()
		complete := save(file, resp)

		end := time.Now()
		size := complete / 1024
		msg = fmt.Sprintf("Downloaded ~%v kb in %v \n", size, end.Sub(start))
	} else {
		readFile(input, output)
		end := time.Now()
		msg = fmt.Sprintf("Downloaded contents of %v in %v\n", input, end.Sub(start))
	}

	go notify(msg, done)
	fmt.Printf(msg)
	<-done
}
