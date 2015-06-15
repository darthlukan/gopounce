# GoPounce

> Author: Brian Tomlinson <brian.tomlinson@linux.com>

## Description

> ```gopounce``` is a simple file downloader written in Go and is nothing more than a personal learning tool.  I don't
expect it to work outside of the most simple of cases. Your mileage may vary.


## Installation

```
    $ go get github.com/darthlukan/gopounce
```


## Usage

```
    $ gopounce -u http://www.google.com -o Downloads/google.html

    OR

    $ gopounce --url http://www.google.com --outfile Downloads/google.html

    OR

    $ gopounce -f /path/to/file/with/links -d Downloads/

    OR

    $ gopounce --file /path/to/file/with/links --dir Downloads/
```


## License

> The Unlicense, see the included LICENSE file.

> Basically, do what you want with it :)
