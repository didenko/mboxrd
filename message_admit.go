package mboxrd

import (
	"fmt"
	"regexp"
)

type Criterion struct {
	OnlyHeaders bool
	RE          *regexp.Regexp
}

const (
	sdrFmt = "^(From:|Sender:|Reply-To:).*[\\s<,:]%s[\\s>,\\]].*"
	rvrFmt = "^(To:\\s|Cc:\\s|Bcc:\\s|(\\s+)).*[<,:]%s[\\s>,\\]].*"
)

var (
	banChat = regexp.MustCompile(`^X-Gmail-Labels\: Chat`)
)

func AdmitAnyPattern(criteria []Criterion, vetos []Criterion, errors chan error) ByLineAdmit {

	var (
		admitted  = false
		banned = false
		inHeaders = true
	)

	return func(line string, errors chan error) bool {

		if banned {
			return false
		}

		if admitted {
			return true
		}

		if inHeaders && line == "" {
			inHeaders = false
		}

		for _, veto := range vetos {

			if veto.OnlyHeaders && !inHeaders {
				continue
			}

			if veto.RE.MatchString(line) {
				banned = true
				return false
			}
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

func AllWith(addrs []string, errors chan error) ByLineAdmit {

	criteria := make([]Criterion, len(addrs)*2)
	vetos := make([]Criterion, 1)

	for i, addr := range addrs {
		pos := i * 2
		criteria[pos] = Criterion{true, regexp.MustCompile(fmt.Sprintf(sdrFmt, addr))}
		criteria[pos+1] = Criterion{true, regexp.MustCompile(fmt.Sprintf(rvrFmt, addr))}
	}

	vetos[0] = Criterion{true, banChat}

	return AdmitAnyPattern(criteria, vetos, errors)
}
