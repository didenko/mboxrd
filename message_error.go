package mboxrd

import "fmt"

// MessageError type is returned when there are errors occurred
// writing a mesage to filesystem.
type MessageError string

func (me MessageError) Error() string {
	return fmt.Sprintf("Message error: %s", string(me))
}
