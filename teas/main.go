package main

import (
	"fmt"

	"gitlab.com/hokiegeek/teadb"
)

func main() {
	teadb.CreateTea(44)

	tea42, err := teadb.GetTeaByID(42)
	if err != nil {
		panic(err)
	}
	fmt.Printf("tea42 (%d): %q\n", tea42.ID, tea42.Comments)

	teas, err := teadb.GetAllTeas()
	if err != nil {
		panic(err)
	}

	for _, tea := range teas {
		fmt.Printf("Tea (%d): %q\n", tea.ID, tea.Comments)
	}
}
