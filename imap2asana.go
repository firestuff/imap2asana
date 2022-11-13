package main

import (
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

	ac := NewAsanaClient()

	err = Poll(ic, ac)
	if err != nil {
		log.Printf("%s", err)
	}

	for {
		time.Sleep(time.Duration(rand.Intn(60)) * time.Second)

		err := Poll(ic, ac)
		if err != nil {
			log.Printf("%s", err)
		}
	}
}

type Task struct {
	Name      string
	HtmlNotes string
	Uid       uint32
}

func Poll(ic *ImapClient, ac *AsanaClient) error {
	tasks, err := ic.Poll()
	if err != nil {
		return err
	}

	if len(tasks) < 1 {
		return nil
	}

	for _, task := range tasks {
		log.Printf("%s", task.Name)

		err = ac.CreateTask(task.Name, task.HtmlNotes)
		if err != nil {
			return err
		}

		err = ic.Archive(task)
		if err != nil {
			return err
		}
	}

	return nil
}
