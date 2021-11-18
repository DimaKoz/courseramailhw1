package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
)

// Node represents a node of a tree of files.
type Node interface {
	fmt.Stringer
}

// DirInfo represents the name and contents of a directory.
type DirInfo struct {
	name     string
	children []Node
}

// FileInfo represents the name and the size of a file.
type FileInfo struct {
	name string
	size int64
}

// Join concatenates the elements to create a single string.
func concatStrings(strings ...string) (string, error) {
	var buf bytes.Buffer
	var err error
	for _, item := range strings {
		_, err = buf.WriteString(item)
		if err != nil {
			return "", err
		}
	}
	return buf.String(), err
}

// Return a string presentation of the FileInfo
// or empty string if any error happened
func (fileInfo FileInfo) String() string {
	if fileInfo.size == 0 {
		result, err := concatStrings(fileInfo.name, " (empty)")
		if err != nil {
			return ""
		}
		return result
	}
	result, err := concatStrings(fileInfo.name, " (", strconv.FormatInt(fileInfo.size, 10), "b)")
	if err != nil {
		return ""
	}
	return result
}

// Return a string presentation of the DirInfo
func (directory DirInfo) String() string {
	return directory.name
}

// Read a directory
func readDir(path string, nodes *[]Node, withFiles bool) (result *[]Node, err error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	files, err := file.Readdir(0)
	defer func() {
		cErr := file.Close()
		if err == nil {
			err = cErr
		}
	}()

	sort.Slice(files, func(i, j int) bool {
		return files[i].Name() < files[j].Name()
	})

	for _, info := range files {
		if !(info.IsDir() || withFiles) {
			continue
		}

		var newNode Node
		if info.IsDir() {
			children, _ := readDir(filepath.Join(path, info.Name()), &[]Node{}, withFiles)
			newNode = DirInfo{info.Name(), *children}
		} else {
			newNode = FileInfo{info.Name(), info.Size()}
		}

		*nodes = append(*nodes, newNode)
	}

	return nodes, err
}

// Print nodes of a tree of files.
func printDir(out io.Writer, nodes []Node, prefixes []string) error {
	if len(nodes) == 0 {
		return nil
	}

	totalPrefixes, err := concatStrings(prefixes...)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(out, "%s", totalPrefixes)
	if err != nil {
		return err
	}

	node := nodes[0]

	if len(nodes) == 1 {
		_, err = fmt.Fprintf(out, "%s%s\n", "└───", node)
		if err != nil {
			return err
		}
		if directory, ok := node.(DirInfo); ok {
			return printDir(out, directory.children, append(prefixes, "\t"))
		}
		return nil
	}

	_, err = fmt.Fprintf(out, "%s%s\n", "├───", node)
	if err != nil {
		return err
	}
	if directory, ok := node.(DirInfo); ok {
		err = printDir(out, directory.children, append(prefixes, "│\t"))
		if err != nil {
			return err
		}
	}

	return printDir(out, nodes[1:], prefixes)
}

func dirTree(out io.Writer, path string, isPrintFiles bool) error {
	nodes, err := readDir(path, &[]Node{}, isPrintFiles)
	if err != nil {
		return err
	}
	if nodes == nil {
		return errors.New("no nodes for you")
	}

	return printDir(out, *nodes, []string{})
}

func main() {
	out := os.Stdout
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
