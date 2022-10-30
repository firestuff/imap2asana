package main

import (
	"fmt"
	"os"
)

func main() {
	ic, err := NewImapClient(
		os.Getenv("IMAP_HOST"),
		os.Getenv("IMAP_USERNAME"),
		os.Getenv("IMAP_PASSWORD"),
		"Asana",
		"Archive",
	)
	if err != nil {
		panic(err)
	}

	defer ic.Close()

	tasks, err := ic.Poll()
	if err != nil {
		panic(err)
	}

	for _, task := range tasks {
		fmt.Printf("%#v\n", task)
	}
}

type Task struct {
	Name      string
	HtmlNotes string
}
