package mboxrd

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"sync"
)

// WriteMessage receives a message text from the `message` channel
// and writes it into a file in the destination `dir` directory.
// The message file name is constructed by the `nameFunc` parameter
// function.
//
// All error are posted in the parameter channel. The `WaitGroup`
// parameter must be properly initialised and incremented prior
// to calling this function, or be supplied as `nil` if not needed.
func WriteMessage(
	message chan string,
	errors chan error,
	dir string,
	wg *sync.WaitGroup,
	nameFunc func(string, chan error) string) {

	if wg != nil {
		defer wg.Done()
	}

	var msgFile string

	tempFile, err := ioutil.TempFile(dir, "_msg_")
	if err != nil {
		errors <- err
		return
	}

	for line := range message {
		tempFile.WriteString(line + crlf)
		if msgFile == "" {
			msgFile = nameFunc(line, errors)
		}
	}

	if msgFile == "" {
		errors <- MessageError(
			fmt.Sprintf(
				"File name did not constuct, the message left in the %q file",
				tempFile.Name()))
		return
	}

	msgPath := path.Join(dir, msgFile)

	_, err = os.Stat(msgPath)
	if err == nil {
		errors <- MessageError(
			fmt.Sprintf(
				"The message file %q already exists, the message left in the %q file",
				msgPath,
				tempFile.Name()))
		return
	}

	err = os.Rename(tempFile.Name(), msgPath)
	if err != nil {
		errors <- MessageError(
			fmt.Sprintf(
				"Problem renaming %q into %q, the file may have either of the names. Error: %s",
				tempFile.Name(),
				msgPath,
				err.Error()))
	}
}
