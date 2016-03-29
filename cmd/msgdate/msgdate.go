package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/didenko/mboxrd"
)

var (
	dirName   string
	msgSuffix string
	loc       *time.Location
	err       error
	lg        = log.New(os.Stderr, "", log.Lshortfile)
	prefixRE  = regexp.MustCompile("([[:digit:]]{12})(.*)")
	dateRE    = regexp.MustCompile(`(?i)^date: (.*)`)
)

func init() {
	var tz string
	var defaults bool
	flag.StringVar(&dirName, "dir", ".", "A directory name to scan")
	flag.StringVar(&tz, "loc", "UTC", "A location name from the IANA Time Zone database")
	flag.StringVar(&msgSuffix, "ext", ".eml", "A file extension (including dot) to be recognised as a message file")
	flag.BoolVar(&defaults, "defaults", false, "Proceed with default values for parameters without a warning")
	flag.Parse()

	if flag.NFlag() < 1 && !defaults {
		flag.PrintDefaults()

		fmt.Fprintf(os.Stderr, "\n%s\n", "Kill this process (usually Ctrl-C) to avoid running it with the default parameter values.")
		fmt.Fprintf(os.Stderr, "%s\n", "Press Enter to continue processing.")
		fmt.Scanln()
	}

	loc, err = time.LoadLocation(tz)
	if err != nil {
		lg.Fatal(err)
	}
}

func stampInFile(fn string) (string, error) {

	f, err := os.Open(fn)
	if err != nil {
		return "", err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if line := scanner.Text(); dateRE.MatchString(line) {
			return mboxrd.TimeFromLine(line, loc)
		}
	}
	if err = scanner.Err(); err != nil {
		return "", err
	}
	return "", fmt.Errorf("Timestamp not found in file: %q", fn)
}

type trans struct {
	tsNew string
	files []string
}

type message_heap map[string]*trans

func scanMessages(dir string) message_heap {
	files, err := ioutil.ReadDir(dirName)
	if err != nil {
		lg.Fatal(err)
	}

	messages := make(map[string]*trans)

	for _, file := range files {

		if !file.Mode().IsRegular() {
			continue
		}

		fn := file.Name()
		parts := prefixRE.FindStringSubmatch(fn)
		if parts == nil {
			lg.Printf("Skipping %q\n", fn)
			continue
		}

		if msgInfo, ok := messages[parts[1]]; ok {
			msgInfo.files = append(msgInfo.files, parts[2])
		} else {
			msgInfo = &trans{"", []string{parts[2]}}
			messages[parts[1]] = msgInfo
		}

		if strings.HasSuffix(fn, msgSuffix) {
			stamp, err := stampInFile(fn)
			if err != nil {
				lg.Fatal(err)
			}
			messages[parts[1]].tsNew = stamp
		}
	}
	return messages
}

func processMessages(msgs message_heap) {
	for tsOld, msg := range msgs {
		for _, file := range msg.files {
			if tsOld != msg.tsNew {
				oldName := tsOld + file
				newName := msg.tsNew + file
				os.Rename(oldName, newName)
				lg.Printf("%s => %s", oldName, newName)
			}
		}
	}
}

func main() {
	processMessages(scanMessages(dirName))
}
