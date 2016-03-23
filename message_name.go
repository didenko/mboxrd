package mboxrd

import (
	"fmt"
	"log"
	"regexp"
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

func TimeNorm(line string, errors chan error) (string, error) {
	const datePrefix = "Date: "

	if strings.HasPrefix(line, datePrefix) {

		t, er := time.Parse("Mon, 2 Jan 2006 15:04:05 -0700", strings.TrimPrefix(line, datePrefix))
		if er == nil {
			return t.In(loc).Format("060102150405"), nil
		}

		t, er = time.Parse("Mon, 2 Jan 2006 15:04:05 -0700 (MST)", strings.TrimPrefix(line, datePrefix))
		if er == nil {
			return t.In(loc).Format("060102150405"), nil
		}

		t, er = time.Parse("2 Jan 2006 15:04:05 -0700", strings.TrimPrefix(line, datePrefix))
		if er == nil {
			return t.In(loc).Format("060102150405"), nil
		}

		return "", er
	}

	return "", nil
}

// NameFromTimeUser returns a closed function used to extract
// a message file name based on the message timestamp and sender's
// username part of the email.
//
// It is an example on how to construct the file name from multiple
// headers.
func NameFromTimeUser(format string, errors chan error) ByLineName {
	const fromPrefix = "From: "

	var (
		ts, fr string
		fromRE = regexp.MustCompile("(.*<)?(.*)(@.*)")
	)

	return func(line string, errors chan error) string {
		var er error

		if ts == "" {
			ts, er = TimeNorm(line, errors)
			if er != nil {
				errors <- fmt.Errorf(
					"Failed to parse the timestamp. Error: %s",
					er.Error())
				return ""
			}
		}

		if fr == "" && strings.HasPrefix(line, fromPrefix) {

			email := strings.TrimPrefix(line, fromPrefix)
			parsedEmail := fromRE.FindStringSubmatch(email)

			if parsedEmail == nil {
				errors <- fmt.Errorf(
					"Failed to extract user name from the email address %q",
					email)
				return ""
			}

			fr = parsedEmail[2]
		}

		if ts == "" || fr == "" {
			return ""
		}

		return fmt.Sprintf(format, ts, fr)
	}
}
