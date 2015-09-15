package mboxrd

import (
	"fmt"
	"regexp"
)

type Criterion struct {
	OnlyHeaders bool
	RE          *regexp.Regexp
}

func AdmitAnyPattern(criteria []Criterion, errors chan error) ByLineAdmit {

	var (
		admitted  = false
		inHeaders = true
	)

	return func(line string, errors chan error) bool {

		if admitted {
			return true
		}

		if inHeaders && line == "" {
			inHeaders = false
		}

		for _, permit := range criteria {

			if permit.OnlyHeaders && !inHeaders {
				continue
			}

			if admitted = permit.RE.MatchString(line); admitted {
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

	criteria := make([]Criterion, len(emails)*2)

	for i, email := range emails {
		pos := i * 2
		criteria[pos] = Criterion{true, regexp.MustCompile(fmt.Sprintf(sdrFmt, email))}
		criteria[pos+1] = Criterion{true, regexp.MustCompile(fmt.Sprintf(rvrFmt, email))}
	}

	return AdmitAnyPattern(criteria, errors)
}
