package main

import (
	"fmt"
	"os"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
)

func main() {
	ic, err := NewImapClient(
		os.Getenv("IMAP_HOST"),
		os.Getenv("IMAP_USERNAME"),
		os.Getenv("IMAP_PASSWORD"),
	)
	if err != nil {
		panic(err)
	}

	defer ic.Close()

	mboxs, err := ic.List("", "*")
	if err != nil {
		panic(err)
	}

	for _, mbox := range mboxs {
		fmt.Printf("%#v\n", mbox)
	}
}

// go-imap's API is a pile of stupid; wrap it
type ImapClient struct {
	cli *client.Client
}

func NewImapClient(host, user, pass string) (*ImapClient, error) {
	c, err := client.DialTLS(host, nil)
	if err != nil {
		return nil, err
	}

	err = c.Login(user, pass)
	if err != nil {
		c.Logout()
		return nil, err
	}

	return &ImapClient{
		cli: c,
	}, nil
}

func (ic *ImapClient) Close() {
	ic.cli.Logout()
}

func (ic *ImapClient) List(ref, name string) ([]*imap.MailboxInfo, error) {
	ch := make(chan *imap.MailboxInfo, 10)
	done := make(chan error)
	go func() {
		done <- ic.cli.List(ref, name, ch)
	}()

	ret := []*imap.MailboxInfo{}
	for mbox := range ch {
		ret = append(ret, mbox)
	}

	err := <-done
	if err != nil {
		return nil, err
	} else {
		return ret, nil
	}
}
