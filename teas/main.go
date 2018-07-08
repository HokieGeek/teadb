package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"gitlab.com/hokiegeek.net/teadb"
)

func main() {
	// Command flags: load
	loadCommand := flag.NewFlagSet("load", flag.ExitOnError)
	loadFilePtr := loadCommand.String("file", "", "The filename to use")

	// Command flags: save
	saveCommand := flag.NewFlagSet("save", flag.ExitOnError)
	saveFilePtr := saveCommand.String("file", "", "The filename to use")

	command := os.Args[1]

	switch command {
	case "load":
		loadCommand.Parse(os.Args[2:])
		if err := loadFromJSON(*loadFilePtr); err != nil {
			panic(err)
		}
	case "save":
		saveCommand.Parse(os.Args[2:])
		if err := saveToFile(*saveFilePtr); err != nil {
			panic(err)
		}
	default:
		log.Fatalf("Unrecognized command: %s\n", command)
	}
}

func loadFromJSON(filename string) error {
	raw, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	var teas []teadb.Tea
	json.Unmarshal(raw, &teas)

	return createTeas(teas)
}

func createTeas(teas []teadb.Tea) error {
	for _, tea := range teas {
		if err := teadb.CreateTea(tea); err != nil {
			fmt.Printf("Could not create tea %d: %v\n", tea.ID, err)
		} else {
			fmt.Printf("Tea (%d): %s\n", tea.ID, tea.Name)
		}
	}

	return nil
}

func saveToFile(filename string) error {
	teas, err := teadb.GetAllTeas()
	if err != nil {
		return err
	}

	numEntries := 0
	for _, t := range teas {
		numEntries += len(t.Entries)
	}

	fmt.Printf("Retrieved %d teas and %d entries\n", len(teas), numEntries)

	teasJSON, err := json.Marshal(teas)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(filename, teasJSON, 0644)
}
