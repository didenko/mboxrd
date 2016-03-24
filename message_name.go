package mboxrd

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"
)

const timestampFormat = "060102150405"

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
	t, er := time.Parse("Mon, 2 Jan 2006 15:04:05 -0700", line)
	if er == nil {
		return t.In(loc).Format(timestampFormat), nil
	}

	t, er = time.Parse("Mon, 2 Jan 2006 15:04:05 -0700 (MST)", line)
	if er == nil {
		return t.In(loc).Format(timestampFormat), nil
	}

	t, er = time.Parse("2 Jan 2006 15:04:05 -0700", line)
	if er == nil {
		return t.In(loc).Format(timestampFormat), nil
	}

	return "", er
}

// NameFromTimeUser returns a closed function used to extract
// a message file name based on the message timestamp and sender's
// username part of the email.
//
// It is an example on how to construct the file name from multiple
// headers.
func NameFromTimeUser(format string, errors chan error) ByLineName {
	const (
		datePrefix = "Date: "
		fromPrefix = "From: "
		headPrefix = "From "
	)

	var (
		ts, fr, hd string
		fromRE = regexp.MustCompile("(.*<)?(.*)(@.*)")
		headRE = regexp.MustCompile("^From (.*)(@.*)")
	)

	return func(line string, errors chan error) string {
		var er error

		if ts == "" && strings.HasPrefix(line, datePrefix){
			ts, er = TimeNorm(strings.TrimPrefix(line, datePrefix), errors)
			if er != nil {
				errors <- fmt.Errorf(
					"Failed to parse the timestamp. Error: %s",
					er.Error())
				return ""
			}
		}

		if hd == "" && strings.HasPrefix(line, headPrefix) {
			parsedHead := headRE.FindStringSubmatch(line)

			if parsedHead == nil {
				errors <- fmt.Errorf(
					"Failed to extract sender ID from the header line %q",
					line)
				return ""
			}

			hd_temp := parsedHead[1]
			hd = hd_temp[len(hd_temp)-8:]
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

		return fmt.Sprintf(format, ts, hd, fr)
	}
}
