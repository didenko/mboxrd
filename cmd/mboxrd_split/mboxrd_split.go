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
	addr   string
	emlWG  sync.WaitGroup
	origWG sync.WaitGroup
	workWG sync.WaitGroup

	messages = make(chan chan string)
	emlNames = make(chan string)
	errors   = make(chan error)
)

func logErrors(lg *log.Logger, errors chan error) {
	for err := range errors {
		lg.Println(err)
	}
}

func init() {

	flag.StringVar(&dir, "dir", "", "A directory to put the resulting messages to")
	flag.StringVar(&mbox, "mbox", "", "An mbox file to process")
	flag.StringVar(&addr, "email", "", "An email which correspondence to be captured")
	flag.Parse()

	if dir == "" || mbox == "" {
		lg.Fatal("Parameters 'dir' and 'mbox' are required")
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

	go logErrors(lg, errors)

	go mboxrd.Extract(mboxFile, messages, errors)

	workWG.Add(1)
	go func() {
		defer workWG.Done()
		for message := range messages {

			origWG.Add(1)

			var admit mboxrd.ByLineAdmit
			if addr != "" {
				admit = mboxrd.AllWith([]string{addr}, errors)
			} else {
				admit = mboxrd.AllWith([]string{}, errors)
			}

			go mboxrd.WriteOriginal(
				message,
				emlNames,
				errors,
				dir,
				admit,
				mboxrd.NameFromTimeUser("%s_%s_%s.eml", errors),
				&origWG)
		}
	}()

	go func() {
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
