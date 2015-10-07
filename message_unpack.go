package mboxrd

import (
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"mime"
	"mime/multipart"
	"net/mail"
	"os"
	"strings"
	"sync"
)

func unpackPart(part *multipart.Part, emlbase string, errors chan error) {

	defer part.Close()

	partFileName := part.FileName()
	if partFileName == "" {
		return
	}

	attachmentFileName := emlbase + " " + partFileName

	attachmentFile, err := os.Create(attachmentFileName)
	if err != nil {
		errors <- MessageError(
			fmt.Sprintf(
				"Problem opening the %q file: %s",
				attachmentFileName,
				err.Error()))
		return
	}
	defer attachmentFile.Close()

	enc := part.Header.Get("Content-Transfer-Encoding")

	var partReader io.Reader

	switch enc {
	case "", "7bit", "8bit":
		partReader = part

	case "base64":
		partReader = base64.NewDecoder(base64.StdEncoding, part)

	default:
		errors <- MessageError(
			fmt.Sprintf(
				"Attachment %q: unknown encoging %q",
				attachmentFileName,
				enc))
		return
	}

	_, err = io.Copy(attachmentFile, partReader)
	if err != nil {
		errors <- MessageError(
			fmt.Sprintf(
				"Problem copying the %q part of the %q message: %s",
				attachmentFile,
				emlbase,
				err.Error()))
		return
	}
	fmt.Printf("Message: %q, part %q\n", emlbase, attachmentFileName)
}

func cutExt(fname string) string {
	dotIdx := strings.LastIndex(fname, ".")
	if dotIdx < 0 {
		return fname
	}

	return fname[:dotIdx]
}

func UnpackMessage(eml string, errors chan error, wg *sync.WaitGroup) {

	if wg != nil {
		defer wg.Done()
	}

	f, err := os.Open(eml)
	if err != nil {
		errors <- MessageError(
			fmt.Sprintf(
				"Problem opening the %q message file for unpacking: %s",
				eml,
				err.Error()))
		return
	}
	defer f.Close()

	msg, err := mail.ReadMessage(f)
	if err != nil {
		errors <- MessageError(
			fmt.Sprintf(
				"Problem opening the %q message file for unpacking: %s",
				eml,
				err.Error()))
		return
	}

	mediaType, params, err := mime.ParseMediaType(msg.Header.Get("Content-Type"))
	if err != nil {
		log.Fatal(err)
	}

	if strings.HasPrefix(mediaType, "multipart/") {

		mr := multipart.NewReader(msg.Body, params["boundary"])

		for {

			part, err := mr.NextPart()
			if err == io.EOF {
				return
			}
			if err != nil {
				errors <- MessageError(
					fmt.Sprintf(
						"Problem opening a part of the %q message: %s",
						eml,
						err.Error()))
				return
			}

			unpackPart(part, cutExt(eml), errors)
		}
	}
}
