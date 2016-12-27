// This small program is used for generating module glue files from templates.
// It scans the modules main directory and outputs the template in the output file
// using as values all directories found in the modules' main directory.
package main

import (
	"bufio"
	"flag"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
)

var (
	templateFile string
	outputFile   string
	inputFile    string
)

func init() {
	flag.StringVar(&inputFile, "inputlist", "",
		`The input list location, if given every line will be an argument to the template.
If the file doesn't exist that will be logged but will not stop the go generate.`)
	flag.StringVar(&templateFile, "template", "",
		"The input template file location")
	flag.StringVar(&outputFile, "output", "",
		"The location of the file which will be generated from the template")
}

func main() {
	flag.Parse()

	if templateFile == "" {
		log.Fatalln("The -template argument is required. See -h.")
	}

	if outputFile == "" {
		log.Fatalln("The -output argument is required. See -h.")
	}

	tpl, err := template.ParseFiles(templateFile)

	if err != nil {
		log.Fatalln("Error parsing template file.", err)
	}

	var directories []pkg
	if inputFile == "" {
		directories = walkDirectories()
	} else {
		directories = readLinesFromInputFile()
		if len(directories) == 0 {
			path, _ := filepath.Abs(inputFile)
			log.Printf("%s is empty or missing", path)
			return
		}
	}

	outFile, err := os.Create(outputFile)
	if err != nil {
		log.Fatalln("Error creating output file.", err)
	}

	err = tpl.Execute(outFile, directories)

	if err != nil {
		log.Fatalln("Error executing the template.", err)
	}

	cmd := exec.Command("go", "fmt", outputFile)

	if err := cmd.Run(); err != nil {
		log.Fatalln("Error executing go fmt.", err)
	}
}

type pkg string

func (p pkg) String() string {
	return (string)(p)
}

func (p pkg) PkgName() string {
	return filepath.Base(p.String())
}

func readLinesFromInputFile() (directories []pkg) {
	file, err := os.Open(inputFile)
	if err != nil {
		if os.IsNotExist(err) {
			return
		}
		log.Fatalln("error opening input file", err)
	}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		directories = append(directories, pkg(scanner.Text()))
	}
	if err := scanner.Err(); err != nil {
		log.Fatalln("reading from the input file:", err)
	}
	return
}

func walkDirectories() (directories []pkg) {
	workingDirectory, err := os.Getwd()
	walkFunc := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Printf("Error in entry %s: %s\n", path, err)
			return err
		}

		if !info.IsDir() {
			return nil
		}

		path = strings.TrimPrefix(path, workingDirectory)

		if path == "" {
			return nil
		}

		path = strings.TrimPrefix(path, "/")

		if path == "" {
			return nil
		}

		if strings.Contains(path, "/") {
			return nil
		}

		directories = append(directories, pkg(path))

		return nil
	}

	err = filepath.Walk(workingDirectory, walkFunc)

	if err != nil {
		log.Fatalln("Error scaning the module directory", err)
	}
	return
}
