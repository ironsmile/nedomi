/*
   This small program is used for generating module glue files from templates.
   It scans the modules main directory and outputs the template in the output file
   using as values all directories found in the modules' main directory.
*/
package main

import (
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
)

func init() {
	flag.StringVar(&templateFile, "template", "", "The input template file location")
	flag.StringVar(&outputFile, "output", "",
		"The location of the file which will be generated from the template")
}

func main() {

	flag.Parse()

	if templateFile == "" {
		log.Fatalln("The -template argument is required. See -h.")
	}

	if outputFile == "" {
		log.Fatalln("The -ouput argument is required. See -h.")
	}

	tpl, err := template.ParseFiles(templateFile)

	if err != nil {
		log.Fatalln("Error parsing template file.", err)
	}

	outFile, err := os.Create(outputFile)

	if err != nil {
		log.Fatalln("Error creating output file.", err)
	}

	workingDirectory, err := os.Getwd()

	if err != nil {
		log.Fatalln("Could not get the current directory.", err)
	}

	directories := make([]string, 0)

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

		directories = append(directories, path)

		return nil
	}

	err = filepath.Walk(workingDirectory, walkFunc)

	if err != nil {
		log.Fatalln("Error scaning the module directory", err)
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
