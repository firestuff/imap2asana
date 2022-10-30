package main

import (
	"net/mail"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
)

type ImapClient struct {
	cli *client.Client

	fromFolder string
	toFolder   string
}

func NewImapClient(host, user, pass, fromFolder, toFolder string) (*ImapClient, error) {
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
		cli:        c,
		fromFolder: fromFolder,
		toFolder:   toFolder,
	}, nil
}

func (ic *ImapClient) Close() {
	ic.cli.Logout()
}

func (ic *ImapClient) Poll() ([]*Task, error) {
	mbox, err := ic.Select(ic.fromFolder, false)
	if err != nil {
		return nil, err
	}

	if mbox.Messages < 1 {
		return []*Task{}, nil
	}

	seqset := &imap.SeqSet{}
	seqset.AddRange(1, mbox.Messages)

	section := &imap.BodySectionName{}

	msgs, err := ic.Fetch(seqset, []imap.FetchItem{section.FetchItem()})
	if err != nil {
		return nil, err
	}

	ret := []*Task{}

	for _, msg := range msgs {
		msg, err := mail.ReadMessage(msg.GetBody(section))
		if err != nil {
			return nil, err
		}

		ret = append(ret, &Task{
			Name: msg.Header.Get("Subject"),
		})
	}

	err = ic.Move(seqset, ic.toFolder)
	if err != nil {
		return nil, err
	}

	return ret, nil
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

func (ic *ImapClient) Fetch(seqset *imap.SeqSet, toFetch []imap.FetchItem) ([]*imap.Message, error) {
	ch := make(chan *imap.Message, 10)
	done := make(chan error)
	go func() {
		done <- ic.cli.Fetch(seqset, toFetch, ch)
	}()

	ret := []*imap.Message{}
	for msg := range ch {
		ret = append(ret, msg)
	}

	err := <-done
	if err != nil {
		return nil, err
	} else {
		return ret, nil
	}
}

func (ic *ImapClient) Select(name string, readonly bool) (*imap.MailboxStatus, error) {
	return ic.cli.Select(name, readonly)
}

func (ic *ImapClient) Move(seqset *imap.SeqSet, dest string) error {
	return ic.cli.Move(seqset, dest)
}