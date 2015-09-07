package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/didenko/mboxrd"
)

var (
	lg        *log.Logger = log.New(os.Stderr, "", log.Lshortfile)
	loc       *time.Location
	err       error
	mbox, dir string
)

func init() {
	loc, err = time.LoadLocation("UTC")
	if err != nil {
		lg.Fatal(err)
	}

	flag.StringVar(&dir, "dir", "", "A directory to put the resulting messages to")
	flag.StringVar(&mbox, "mbox", "", "An mbox file to process")
	flag.Parse()

	if dir == "" || mbox == "" {
		lg.Fatal("Both dir and mbox parameters are required")
	}
}

func main() {
	err = mboxrd.Extract(mbox, dir, lg, func(line string) string {

		const datePrefix = "Date: "

		if strings.HasPrefix(line, datePrefix) {

			t, er := time.Parse("Mon, 2 Jan 2006 15:04:05 -0700", strings.TrimPrefix(line, datePrefix))
			if er != nil {
				lg.Println(
					fmt.Errorf(
						"Failed to parse the timestamp. Error: %s",
						er.Error()))
				return ""
			}

			return t.In(loc).Format("060102150405") + ".eml"
		}

		return ""
	})

	if err != nil {
		lg.Fatal(err)
	}
}
