package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

const (
	connector = "├───"
	end       = "└───"
	vertical  = "│"
	indent    = "\t"
	hereDot   = "."
)

func printLine(out io.Writer, prefix string, entry os.DirEntry) {
	str := fmt.Sprint(prefix + entry.Name())

	if !entry.IsDir() {
		if info, err := entry.Info(); err == nil {
			if info.Size() > 0 {
				str += fmt.Sprintf(" (%db)", info.Size())
			} else {
				str += fmt.Sprint(" (empty)")
			}
		}
	}

	fmt.Fprintln(out, str)
}

func filter(d []os.DirEntry, printFiles bool) []os.DirEntry {
	res := make([]os.DirEntry, 0)
	for _, v := range d {
		if !v.IsDir() && !printFiles {
			continue
		}
		res = append(res, v)
	}
	return res
}

func goWalk(out io.Writer, prefix string, basePath string, d []os.DirEntry, printFiles bool) {
	filteredEntries := filter(d, printFiles)

	for i, entry := range filteredEntries {
		isLast := i == len(filteredEntries)-1

		var lineSymbol, nextPrefix string
		if isLast {
			lineSymbol = end
			nextPrefix = prefix + indent
		} else {
			lineSymbol = connector
			nextPrefix = prefix + vertical + indent
		}

		printLine(out, prefix+lineSymbol, entry)

		if entry.IsDir() {
			path := filepath.Join(basePath, entry.Name())
			allFiles, _ := os.ReadDir(path)
			goWalk(out, nextPrefix, path, allFiles, printFiles)
		}
	}
}

// `dirTree` prints to `out` a directory tree starting at `path`.
//
// `printFiles` boolean flag determines if normal files should be included as well.
func dirTree(out io.Writer, path string, printFiles bool) error {

	path = filepath.Clean(path)

	fileInfo, err := os.Stat(path)
	if err != nil {
		return err
	}

	if !fileInfo.IsDir() {
		return fmt.Errorf("%s is not a directory", path)
	}

	allFiles, err := os.ReadDir(path)
	if err != nil {
		return err
	}

	prefix := ""

	goWalk(out, prefix, path, allFiles, printFiles)

	return nil
}

func main() {
	out := bufio.NewWriter(os.Stdout)
	defer out.Flush()

	if !(len(os.Args) == 2 || len(os.Args) == 3) {
		panic("usage go run main.go . [-f]")
	}

	path := os.Args[1]

	printFiles := len(os.Args) == 3 && os.Args[2] == "-f"

	err := dirTree(out, path, printFiles)
	if err != nil {
		panic(err.Error())
	}
}
