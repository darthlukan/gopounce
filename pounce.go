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
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/guelfey/go.dbus"
	"io"
	"net/http"
	"os"
)

func notify(msg string) {
	conn, err := dbus.SessionBus()
	if err != nil {
		fmt.Printf("Unable to connect to dbus session bus: %v\n", err)
	}

	obj := conn.Object("org.freedesktop.Notifications", "/org/freedesktop/Notifications")
	call := obj.Call("org.freedesktop.Notifications.Notify", 0, "", uint32(0),
		"", "pounce", msg, []string{}, map[string]dbus.Variant{}, int32(5000))
	if call.Err != nil {
		fmt.Printf("Error while trying to send notification: %v\n", call.Err)
	}
}

func download(url string, response chan<- *http.Response) {
	fmt.Printf("Retreiving from %v\n", url)
	resp, err := http.Get(url)

	if err != nil {
		fmt.Printf("Unable to get resource from '%v'. Error: %v\n", url, err)
	}
	response <- resp
}

func multiDownload(urls []string, destination string) {
	resp := make(chan *http.Response)
	for _, url := range urls {
		go download(url, resp)
	}
	for {
		select {
		case r := <-resp:
			defer r.Body.Close()
		}
	}
}

func readFile(filename string) []string {
	var urls []string

	file, err := os.Open(filename)
	if err != nil {
		fmt.Printf("Problem opening input file!")
		panic(err)
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		urls = append(urls, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("Problem reading input file!")
		panic(err)
	}

	return urls
}

func save(file io.Writer, resp *http.Response) int64 {
	complete, err := io.Copy(file, resp.Body)
	if err != nil {
		fmt.Printf("Unable to write file, error: %v\n", err)
	}
	return complete
}

func main() {
	app := cli.NewApp()
	app.Name = "pounce"
	app.Version = "0.0.2"
	app.Usage = "A very simple file downloader in the vein of wget."
	app.Authors = []cli.Author{cli.Author{
		Name:  "Brian Tomlinson",
		Email: "darthlukan@gmail.com",
	}}
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "url, u",
			Value: "",
			Usage: "The input URL",
		},
		cli.StringFlag{
			Name:  "file, f",
			Value: "",
			Usage: "The input file",
		},
		cli.StringFlag{
			Name:  "outfile, o",
			Value: "",
			Usage: "The output file",
		},
		cli.StringFlag{
			Name:  "dir, d",
			Value: "",
			Usage: "The output directory",
		},
	}
	app.Action = func(c *cli.Context) {
		url := c.String("url")
		inFile := c.String("file")
		outFile := c.String("outfile")
		outDir := c.String("dir")

		if url == "" && inFile == "" {
			url = c.Args()[0]
		}

		if outFile == "" && outDir == "" {
			outFile = fmt.Sprintf("%v/pounced.txt", os.Getenv("HOME"))
		}
	}
	app.Run(os.Args)
}
