package mboxrd

import "fmt"

// MboxError type is returned when there are errors occurred
// reading or splitting a mboxrd archive.
type MboxError string

func (mbe MboxError) Error() string {
	return fmt.Sprintf("MBox error: %s", string(mbe))
}
