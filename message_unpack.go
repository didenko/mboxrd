package mboxrd

import (
	"encoding/base64"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/mail"
	"os"
	"strings"
	"sync"
)

type getter interface {
	Get(key string) string
}

func partsIterate(head getter, body io.Reader, emlbase string, errors chan error) {

	mediaType, params, err := mime.ParseMediaType(head.Get("Content-Type"))
	if err != nil {
		errors <- MessageError(
			fmt.Sprintf(
				"failed to read a part's Content-Type in the %q message: %s",
				emlbase,
				err.Error()))
		return
	}

	if !strings.HasPrefix(mediaType, "multipart/") {
		return
	}

	mr := multipart.NewReader(body, params["boundary"])

	for {

		p, err := mr.NextPart()
		if err == io.EOF {
			return
		}
		if err != nil {
			errors <- MessageError(
				fmt.Sprintf(
					"Problem opening a part of the %q message: %s",
					emlbase,
					err.Error()))
			return
		}

		partsIterate(p.Header, p, emlbase, errors)
		unpackPart(p, emlbase, errors)
	}
}

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

	case "base64", "BASE64", "Base64":
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

	partsIterate(msg.Header, msg.Body, cutExt(eml), errors)
}
