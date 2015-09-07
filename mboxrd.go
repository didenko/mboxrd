package mboxrd

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"regexp"
)

const (
	crlf = string("\r\n")
)

var (
	reNewMessage = regexp.MustCompile(`^From `)
	reUnescape   = regexp.MustCompile(`^\>+From `)
)

type MakeName func(line string) (name string)

func Extract(inFile string, outDir string, lg *log.Logger, nameFunc MakeName) error {

	var (
		line      string
		prevEmpty bool = true
		message   chan string
	)

	fmt.Println("in Extract")

	mbox, err := os.Open(inFile)
	if err != nil {
		return err
	}
	defer mbox.Close()

	scanner := bufio.NewScanner(mbox)

	for scanner.Scan() {

		line = scanner.Text()

		switch {

		case line == "":

			if message != nil {
				if prevEmpty {
					message <- ""
				}
				prevEmpty = true
			}

		case reNewMessage.MatchString(line) && prevEmpty:

			if message != nil {
				close(message)
			}

			message = make(chan string)
			go digest(message, outDir, lg, nameFunc)

			message <- line
			prevEmpty = false

		case reUnescape.MatchString(line):

			line = line[1:]

			if message == nil {
				return MboxrdError("Data found before a message beginning")
			}

			if prevEmpty {
				message <- ""
			}
			message <- line
			prevEmpty = false

		default:

			if message == nil {
				return MboxrdError("Data found before a message beginning")
			}

			if prevEmpty {
				message <- ""
			}
			message <- line
			prevEmpty = false
		}
	}

	if message != nil {
		close(message)
	}

	if err = scanner.Err(); err != nil {
		return err
	}

	return nil
}

func digest(message chan string, outDir string, lg *log.Logger, nameFunc MakeName) {

	var fileName string

	tempFile, err := ioutil.TempFile(outDir, "_msg_")
	if err != nil {
		log.Println(err)
		return
	}

	for line := range message {
		tempFile.WriteString(line + crlf)
		if fileName == "" {
			fileName = nameFunc(line)
		}
	}

	if fileName == "" {
		lg.Println(
			MessageError(
				fmt.Sprintf(
					"No file name received, the message was left in the %q file",
					tempFile.Name())))
		return
	}

	fullName := path.Join(outDir, fileName)
	err = os.Rename(tempFile.Name(), fullName)
	if err != nil {
		lg.Println(
			MessageError(
				fmt.Sprintf(
					"Problem renaming %q into %q, the file may have either of the names. Error: %s",
					tempFile.Name(),
					fullName,
					err.Error())))
	}
}
