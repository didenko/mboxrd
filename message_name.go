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
	var l string

	if li := strings.LastIndex(line, ` (`); li == -1 {
		l = line
	} else {
		l = line[:li]
	}

	l = strings.TrimSpace(strings.Replace(strings.TrimSuffix(l, " UT"), `GMT`, ``, 1))

	t, er := time.Parse("Mon, 2 Jan 2006 15:04:05 -0700", l)
	if er == nil {
		return t.In(loc).Format(timestampFormat), nil
	}

	t, er = time.Parse("2 Jan 2006 15:04:05 -0700", l)
	if er == nil {
		return t.In(loc).Format(timestampFormat), nil
	}

	t, er = time.Parse("2 Jan 2006 15:04:05 MST", l)
	if er == nil {
		return t.In(loc).Format(timestampFormat), nil
	}

	t, er = time.Parse("Mon, 2 Jan 2006 15:04:05 MST", l)
	if er == nil {
		return t.In(loc).Format(timestampFormat), nil
	}

	t, er = time.Parse("Mon, 2 Jan 2006 15:04:05", l)
	if er == nil {
		return t.In(loc).Format(timestampFormat), nil
	}

	t, er = time.Parse("2006-01-02 15:04:05 -0700", l)
	if er == nil {
		return t.In(loc).Format(timestampFormat), nil
	}

	//Wed, 6 Aug 2014 09:59:18 GMT-07:00

	t, er = time.Parse("Mon, 2 Jan 2006 15:04:05 Z07:00", l)
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
	const headPrefix = "From "

	var (
		ts, fr, fr_acc, hd string

		in_from = false

		addrRE = regexp.MustCompile(`(\w{1,})["']?(@.*)`)
		headRE = regexp.MustCompile(`^From (.*)(@.*)`)
		dateRE = regexp.MustCompile(`(?i)^date: (.*)`)
		fromRE = regexp.MustCompile(`(?i)^(from:)(.*)`)
		indentRE = regexp.MustCompile(`^\s`)
	)

	return func(line string, errors chan error) string {
		var er error

		if ts == "" && dateRE.MatchString(line) {
			parsedTS := dateRE.FindStringSubmatch(line)
			if parsedTS == nil {
				errors <- fmt.Errorf(
					"Failed to parse the timestamp from the line %q",
					line)
				return ""
			}

			ts, er = TimeNorm(parsedTS[1], errors)
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

		if fr == "" {
			if parsedFrom := fromRE.FindStringSubmatch(line); parsedFrom != nil {
				in_from = true
				fr_acc = strings.TrimSpace( parsedFrom[2] )

			} else if in_from && indentRE.MatchString(line) {
				fr_acc = fr_acc + strings.TrimSpace(line)

			} else if in_from {
				in_from = false

				parsedEmail := addrRE.FindStringSubmatch(fr_acc)
				if parsedEmail == nil {
					errors <- fmt.Errorf(
						"Failed to extract user name from the email address %q",
						fr_acc)
					return ""
				}

				fr = parsedEmail[1]
			}
		}

		if ts == "" || fr == "" || hd == "" {
			return ""
		}

		return fmt.Sprintf(format, ts, hd, fr)
	}
}
