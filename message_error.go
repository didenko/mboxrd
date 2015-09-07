package mboxrd

import "fmt"

// MessageError type is returned when there are errors occurred
// writing a mesage to filesystem.
type MessageError string

func (msge MessageError) Error() string {
	return fmt.Sprintf("Message error: %s", string(msge))
}
