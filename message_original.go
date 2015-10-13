package mboxrd

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
)

type (
	ByLineAdmit func(string, chan error) bool
	ByLineName  func(string, chan error) string
)

// WriteOriginal receives a message text from the `message` channel
// and writes it into a file in the destination `dir` directory.
//
// All error are posted in the `error` parameter channel.
//
// An `admit` parameter allows to determine if the message is left
// in the target directory. The function is called for each line
// in the message, uncluding headers. The value returned by the
// `admit` function determines if the message is kept in the
// directory.
//
// The message file name is constructed by the `name` parameter
// function. The function is called for each line in the
// message, uncluding headers, until it returns a non-empty
// string. If `name` parameter is `nill` then messages will stay
// in randomly named temporary files starting with `_msg_` prefix
//
// The `WaitGroup` parameter must be properly initialised and
// incremented prior to calling this function, or be supplied as
// `nil` if not needed.
func WriteOriginal(
	message chan string,
	emlName chan string,
	errors chan error,
	dir string,
	admit ByLineAdmit,
	name ByLineName,
	wg *sync.WaitGroup) {

	if wg != nil {
		defer wg.Done()
	}

	var (
		msgFile string
		allowed = true
	)

	tempFile, err := ioutil.TempFile(dir, "_msg_")
	if err != nil {
		errors <- err
		return
	}

	for line := range message {

		if admit != nil {
			allowed = admit(line, errors)
		}

		tempFile.WriteString(line + crlf)
		if name != nil && msgFile == "" {
			msgFile = name(line, errors)
		}
	}

	if !allowed {
		defer func() {

			if err := tempFile.Close(); err != nil {
				errors <- MessageError(
					fmt.Sprintf(
						"Problem while closing the %s temporary file: %s",
						tempFile.Name(),
						err.Error()))
			}

			if err := os.Remove(tempFile.Name()); err != nil {
				errors <- MessageError(
					fmt.Sprintf(
						"Problem while deleting the %s temporary file: %s",
						tempFile.Name(),
						err.Error()))
			}
		}()
		return
	}

	tempFileEml := filepath.Base(tempFile.Name()) + ".eml"
	if name != nil && msgFile == "" {
		msgFile = tempFileEml
		errors <- MessageError(
			fmt.Sprintf(
				"File name did not constuct, moving message to the %q file",
				msgFile))
	}

	msgPath := filepath.Join(dir, msgFile)

	_, err = os.Stat(msgPath)
	if err == nil {

		if msgFile != tempFileEml {

			msgFile = tempFileEml
			errors <- MessageError(
				fmt.Sprintf(
					"The message file %q already exists, moving message to the %q file",
					msgPath,
					msgFile))

			msgPath = filepath.Join(dir, msgFile)
			_, err = os.Stat(msgPath)
		}

		if err == nil {
			errors <- MessageError(
				fmt.Sprintf(
					"The message file %q already exists, the message left in the %q file",
					msgPath,
					tempFile.Name()))
			emlName <- tempFile.Name()
			return
		}
	}

	tempFile.Close()
	err = os.Rename(tempFile.Name(), msgPath)
	if err != nil {
		errors <- MessageError(
			fmt.Sprintf(
				"Problem renaming %q into %q, the file may have either of the names. Error: %s",
				tempFile.Name(),
				msgPath,
				err.Error()))
		return
	}

	emlName <- msgPath
}
