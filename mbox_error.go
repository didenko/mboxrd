package mboxrd

import "fmt"

type MboxrdError string

func (me MboxrdError) Error() string {
	return fmt.Sprintf("Error parsing the mbox file: %s", string(me))
}
