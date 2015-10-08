package main

import (
	"flag"
	"log"
	"os"
	"sync"

	"github.com/didenko/mboxrd"
)

var (
	lg     *log.Logger = log.New(os.Stderr, "", log.LstdFlags)
	err    error
	mbox   string
	dir    string
	email  string
	emlWG  sync.WaitGroup
	origWG sync.WaitGroup
	workWG sync.WaitGroup
)

func logErrors(lg *log.Logger, errors chan error) {
	for err := range errors {
		lg.Println(err)
	}
}

func init() {

	flag.StringVar(&dir, "dir", "", "A directory to put the resulting messages to")
	flag.StringVar(&mbox, "mbox", "", "An mbox file to process")
	flag.StringVar(&email, "email", "", "An email which correspondence to be captured")
	flag.Parse()

	if dir == "" || mbox == "" || email == "" {
		lg.Fatal("All of: dir, mbox, and email parameters are required")
	}

	fi, err := os.Stat(dir)
	if err != nil {
		lg.Fatalf("Failed to open the path %q: %s\n", dir, err)
	}

	if !fi.Mode().IsDir() {
		lg.Fatalf("Error: the %q path is not a directory", dir)
	}
}

func main() {

	mboxFile, err := os.Open(mbox)
	if err != nil {
		lg.Panic(err)
	}
	defer mboxFile.Close()

	messages := make(chan chan string)
	emlNames := make(chan string)
	errors := make(chan error)

	go logErrors(lg, errors)

	go mboxrd.Extract(mboxFile, messages, errors)

	workWG.Add(1)
	go func() {
		defer workWG.Done()
		for message := range messages {
			origWG.Add(1)
			go mboxrd.WriteOriginal(
				message,
				emlNames,
				errors,
				dir,
				mboxrd.AllWith([]string{email}, errors),
				mboxrd.NameFromTimeUser("%s_%s.eml", errors),
				&origWG)
		}
	}()

	go func() {
		defer workWG.Done()
		for eml := range emlNames {
			emlWG.Add(1)
			go mboxrd.UnpackMessage(
				eml,
				errors,
				&emlWG)
		}
	}()

	workWG.Wait()
	origWG.Wait()
	emlWG.Wait()
}
