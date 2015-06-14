// pounce.go
package main

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/guelfey/go.dbus"
	"io"
	"net/http"
	"os"
	"time"
)

var (
	transactions map[string]Transaction
)

type Transaction struct {
	Url         string
	Destination string
	Multi       bool
}

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

func create(filename string) *os.File {
	file, err := os.Create(filename)
	if err != nil {
		fmt.Printf("Unable to create file!")
		panic(err)
	}

	return file
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

		if url == "" || inFile == "" {
			fmt.Println(errors.New("Missing input arguments, please see 'pounce --help'\n"))
			return
		} else if outFile == "" || outDir == "" {
			fmt.Println(errors.New("Missing output arguments, please see 'pounce --help'\n"))
			return
		}

		startTime := time.Now()

		respChan := make(chan *http.Response)

		if inFile != "" && outDir != "" {
			urls := readFile(inFile)
			for _, url := range urls {
				transaction := Transaction{Url: url, Destination: outDir, Multi: true}
				transactions[url] = transaction
				go download(url, respChan)
			}
		}

		if url != "" && outFile != "" {
			go download(url, respChan)
		}

		for {
			select {
			case r := <-respChan:
				u, err := r.Location()
				if err != nil {
					fmt.Printf("Caught error: %v\n", err)
				}
				if transaction, ok := transactions[u.Path]; ok == true {
					defer r.Body.Close()
					f := create(transaction.Destination)
					defer f.Close()
					bytesWritten := save(f, r)
					endTime := time.Now()
					msg := fmt.Sprintf("Downloaded %vkb file '%v' in %v\n",
						bytesWritten/1024, f.Name, endTime.Sub(startTime))
					go notify(msg)
				}
			}
		}

	}
	app.Run(os.Args)
}
