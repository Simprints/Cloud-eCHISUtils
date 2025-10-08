package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/simprints/cloud-echisutils/subjectactions/pkg/subjectactions"
)

func main() {
	var file, input string
	flag.StringVar(&file, "file", "", "file to read the input from, one of 'file' or 'input' must be provided")
	flag.StringVar(&input, "input", "", "the input string, one of 'file' or 'input' must be provided")
	flag.Parse()
	if file == "" && input == "" {
		fmt.Println("one of --file or --input must be provided")
		os.Exit(1)
	}
	if file != "" {
		fileInput, err := os.ReadFile(file)
		if err != nil {
			fmt.Printf("failed to read file: %v\n", err)
			os.Exit(1)
		}
		input = string(fileInput)
	}
	spec, err := subjectactions.Check(input)
	if err != nil {
		fmt.Printf("invalid input: %v\n", err)
		os.Exit(1)
	}
	jsonSpec, err := json.Marshal(spec)
	if err != nil {
		fmt.Printf("failed to marshal subject specification: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(string(jsonSpec))
}
