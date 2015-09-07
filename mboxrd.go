package mboxrd

import (
	"bufio"
	"io"
	"regexp"
)

const (
	crlf = string("\r\n")
)

var (
	reNewMessage = regexp.MustCompile(`^From `)
	reUnescape   = regexp.MustCompile(`^\>+From `)
)

// Extract processes all lines from the the mboxrd reader
// and puts resulting messages each as its own channel into
// the provided messages channel.
//
// It will stop only if it runs into non-empty lines prior
// to a message header. Otherwise it will continue processing
// the lines in assumption that the message archive format
// is correct.
//
// Each message's channel and the parent messages' channel
// are closed after the mbox data is exhausted.
func Extract(mboxrd io.Reader, messages chan chan string, errors chan error) {

	var (
		line      string
		prevEmpty = true
		message   chan string
	)

	scanner := bufio.NewScanner(mboxrd)

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
			messages <- message

			message <- line
			prevEmpty = false

		case reUnescape.MatchString(line):

			line = line[1:]

			if message == nil {
				errors <- MboxError("Data found before a message beginning")
				return
			}

			if prevEmpty {
				message <- ""
			}
			message <- line
			prevEmpty = false

		default:

			if message == nil {
				errors <- MboxError("Data found before a message beginning")
				return
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

	if err := scanner.Err(); err != nil {
		errors <- err
	}

	close(messages)

	return
}
