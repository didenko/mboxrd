package mboxrd

// Filterer interface is used to earmark messages which
// should not be stored in the output directory or anyhow
// further processed.
type Filterer interface {
	Filter(in chan string, out chan string)
}
