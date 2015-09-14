package mboxrd

import (
	"fmt"
	"regexp"
)

func AdmitAnyPattern(criteria []*regexp.Regexp, errors chan error) ByLineAdmit {

	var admitted = false

	return func(line string, errors chan error) bool {

		if admitted {
			return true
		}

		for _, permit := range criteria {
			if admitted = permit.MatchString(line); admitted {
				return true
			}
		}
		return false
	}
}

const (
	sdrFmt = "^(From:|Sender:|Reply-To:).*[\\s<,:]%s[\\s>,\\]].*"
	rvrFmt = "^(To:|Cc:|Bcc:).*[\\s<,:]%s[\\s>,\\]].*"
)

func AllWith(emails []string, errors chan error) ByLineAdmit {

	criteria := make([]*regexp.Regexp, len(emails)*2)

	for i, email := range emails {
		pos := i * 2
		criteria[pos] = regexp.MustCompile(fmt.Sprintf(sdrFmt, email))
		criteria[pos+1] = regexp.MustCompile(fmt.Sprintf(rvrFmt, email))
	}

	return AdmitAnyPattern(criteria, errors)
}
