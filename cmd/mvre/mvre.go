package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
)

var (
	find    string
	findRE  *regexp.Regexp
	repl    string
	dir     string
	preview bool
)

func init() {
	flag.StringVar(&find, "find", "", "A required regular expression to search for. See https://golang.org/pkg/regexp/syntax/ for details.")
	flag.StringVar(&repl, "repl", "", "A string with placeholders to use for replacement. Defaults to an empty string. See https://golang.org/pkg/regexp/#Regexp.ReplaceAllString for details.")
	flag.StringVar(&dir, "dir", ".", "A directory to search for matching files.")
	flag.BoolVar(&preview, "preview", false, "Preview only, do not change matching files.")
	flag.Parse()

	if find == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	findRE = regexp.MustCompile(find)
}

func main() {
	err := filepath.Walk(dir, func(oldpath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		var (
			newname string
			newpath string
			oldname = info.Name()
		)

		if newname = findRE.ReplaceAllString(oldname, repl); newname != oldname {
			newpath = filepath.Join(filepath.Dir(oldpath), newname)

			if preview {
				fmt.Printf("%q => %q\n", oldpath, newpath)
			} else {
				er := os.Rename(oldpath, newpath)
				if er != nil {
					return er
				}
			}
		}
		return nil
	})

	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}
