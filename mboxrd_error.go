package mboxrd

import "fmt"

// MboxrdError type is returned when there are errors occurred
// reading or splitting a mboxrd archive.
type MboxrdError string

func (me MboxrdError) Error() string {
	return fmt.Sprintf("MBox error: %s", string(me))
}
