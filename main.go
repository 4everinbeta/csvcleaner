package main

import (
	// "encoding/csv"
	"flag"
	"fmt"
	"os"
)

func main() {
	pathPtr := flag.String("file", "", "Please provide the path to the csv file you would like me to parse...")
	flag.Parse()
	if *pathPtr == "" {
		fmt.Println("Please provide a path to the file you would like me to parse...")
		os.Exit(1)
	}
	if _, err := os.Stat(*pathPtr); err != nil {
		fmt.Printf("I was unable to find the file %s. Please check that it exists and try again.\n", *pathPtr)
		os.Exit(1)
	}
	fmt.Printf("File to parse: %s\n", *pathPtr)

	file, err := os.Open(*pathPtr)
	if err != nil {
		// err is printable
		// elements passed are separated by space automatically
		fmt.Println("Error:", err)
		return
	}
	// automatically call Close() at the end of current method
	defer file.Close()

	ps := new(person)
	rdr, err := NewReadIter(file, ps)

	if err != nil {
		fmt.Printf("Error creating reader: %v\n", err)
	}

	for rdr.Get() {
		fmt.Println(ps.String())
	}
}
