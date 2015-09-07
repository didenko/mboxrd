package mboxrd

import "fmt"

type MessageError string

func (me MessageError) Error() string {
	return fmt.Sprintf("Error processing the message: %s", string(me))
}
