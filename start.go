package main

import (
	"bufio"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"strings"

	"github.com/dustin/go-humanize"
)

func main() {
	file, err := os.Open("links.txt")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		link := strings.TrimSpace(scanner.Text())

		fileResp, err := http.Get(link)
		if err != nil {
			panic(err)
		}
		defer fileResp.Body.Close()

		filename := getFilename(link)
		if contentDisposition := fileResp.Header.Get("Content-Disposition"); contentDisposition != "" {
			_, params, err := mime.ParseMediaType(contentDisposition)
			if err == nil {
				if filenameFromHeader, ok := params["filename"]; ok {
					filename = filenameFromHeader
				}
			}
		}

		if _, err := os.Stat("Download/"); os.IsNotExist(err) {
			errDir := os.MkdirAll("Download/", 0755)
			if errDir != nil {
				panic(errDir)
			}
		}

		file, err := os.Create("Download/" + filename)
		if err != nil {
			panic(err)
		}
		defer file.Close()

		fmt.Printf("%s\n", filename)
		size, err := io.Copy(file, &ReaderWithProgress{reader: fileResp.Body, size: fileResp.ContentLength})
		if err != nil {
			fmt.Println(size)
			panic(err)
		}
		fmt.Printf("\n\n")
	}

	if err := scanner.Err(); err != nil {
		panic(err)
	}
}

type ReaderWithProgress struct {
	reader io.Reader
	size   int64
	total  int64
}

func (r *ReaderWithProgress) Read(p []byte) (n int, err error) {
	n, err = r.reader.Read(p)
	r.total += int64(n)
	if r.total >= r.size {
		fmt.Printf("\rDownloaded %s", humanize.Bytes(uint64(r.total)))
	} else {
		fmt.Printf("\rDownloading %s/%s", humanize.Bytes(uint64(r.total)), humanize.Bytes(uint64(r.size)))
	}
	return
}

func getFilename(path string) string {
	parts := strings.Split(path, "/")
	filename := strings.Split(parts[len(parts)-1], "?")[0]
	if len(filename) == 0 {
		filename = parts[len(parts)-2] + ".exe"
	}
	return filename
}
