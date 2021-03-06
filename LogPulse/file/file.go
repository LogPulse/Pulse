// Package file is used for reading and writing to files.
package file

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"io"
	"mime/multipart"
	"os"
)

// Read will read filename line by line and each line be returned to channel.
func Read(filename string, lineOut chan<- string) {

	go func() {
		inFile, err := os.Open(filename)

		defer func() {
			inFile.Close()
			close(lineOut)
		}()

		if err != nil {
			panic(fmt.Errorf("file.Read: %s", err))
		}
		scanner := bufio.NewScanner(inFile)
		for scanner.Scan() {
			lineOut <- scanner.Text()
		}
	}()
}

// StreamRead will read from io.Reader line by line and each line be returned to channel.
func StreamRead(reader multipart.File, lineOut chan<- string) {
	go func() {
		scanner := bufio.NewScanner(reader)
		for scanner.Scan() {
			val := scanner.Text()
			fmt.Println(val)
			lineOut <- val
		}
		lineOut <- "EOF"
	}()
}

// Write will append or create filename and write the slice of strings separated by a new line.
func Write(filename string, line string) {
	outFile, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	defer outFile.Close()
	if err != nil {
		panic(fmt.Errorf("file.Write: %s", err))
	}
	if _, err = outFile.WriteString(line + "\n"); err != nil {
		panic(fmt.Errorf("file.Write: %s", err))
	}
}

// UnGZip will decompress the file from filename.gz to filename
func UnGZip(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("file.UnGZip: %s", err)
	}
	defer file.Close()
	newfile := filename[0 : len(filename)-len(".gz")]
	out, err := os.Create(newfile)
	if err != nil {
		return fmt.Errorf("file.UnGZip: %s", err)
	}

	defer out.Close()

	r, err := gzip.NewReader(file)
	io.Copy(out, r)
	r.Close()

	return nil
}
