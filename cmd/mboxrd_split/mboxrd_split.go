package main

import (
	"flag"
	"log"
	"os"
	"sync"

	"github.com/didenko/mboxrd"
)

var (
	lg    *log.Logger = log.New(os.Stderr, "", log.Lshortfile)
	err   error
	mbox  string
	dir   string
	msgWG sync.WaitGroup
)

func logErrors(lg *log.Logger, errors chan error) {
	for err := range errors {
		lg.Println(err)
	}
}

func init() {

	flag.StringVar(&dir, "dir", "", "A directory to put the resulting messages to")
	flag.StringVar(&mbox, "mbox", "", "An mbox file to process")
	flag.Parse()

	if dir == "" || mbox == "" {
		lg.Fatal("Both dir and mbox parameters are required")
	}
}

func main() {

	mboxFile, err := os.Open(mbox)
	if err != nil {
		lg.Panic(err)
	}
	defer mboxFile.Close()

	messages := make(chan chan string)
	errors := make(chan error)

	go logErrors(lg, errors)

	go mboxrd.Extract(mboxFile, messages, errors)

	for message := range messages {
		msgWG.Add(1)
		go mboxrd.WriteMessage(
			message,
			errors,
			dir,
			&msgWG,
			mboxrd.NameFromTimeUser("%s_%s.eml", errors))
	}

	msgWG.Wait()
}
