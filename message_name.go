package mboxrd

import (
	"fmt"
	"log"
	"strings"
	"time"
)

var (
	loc *time.Location
)

func init() {
	var err error
	loc, err = time.LoadLocation("UTC")
	if err != nil {
		log.Fatal(err)
	}
}

// NameFromTimestamp returns a message file name. It currently
// only considers one line at a time and should be rewritten to
// construct a file name based on multiple headers.
func NameFromTimestamp(line string, errors chan error) string {

	const datePrefix = "Date: "

	if strings.HasPrefix(line, datePrefix) {

		t, er := time.Parse("Mon, 2 Jan 2006 15:04:05 -0700", strings.TrimPrefix(line, datePrefix))
		if er != nil {
			errors <- fmt.Errorf(
				"Failed to parse the timestamp. Error: %s",
				er.Error())
			return ""
		}

		return t.In(loc).Format("060102150405") + ".eml"
	}

	return ""
}
