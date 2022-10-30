package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"
)

func main() {
	rand.Seed(time.Now().UnixNano())

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

	err = Poll(ic)
	if err != nil {
		log.Printf("%s", err)
	}

	for {
		time.Sleep(time.Duration(rand.Intn(60)) * time.Second)

		err := Poll(ic)
		if err != nil {
			log.Printf("%s", err)
		}
	}
}

type Task struct {
	Name      string
	HtmlNotes string
}

func Poll(ic *ImapClient) error {
	tasks, err := ic.Poll()
	if err != nil {
		return err
	}

	for _, task := range tasks {
		fmt.Printf("%#v\n", task)
	}

	return nil
}
