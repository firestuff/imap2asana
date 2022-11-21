package main

import (
	"fmt"
	"math/rand"
	"mime"
	"net/mail"
	"strings"
	"time"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
)

type ImapClient struct {
	cli     *client.Client
	updates chan client.Update

	fromFolder string
	toFolder   string
}

func NewImapClient(host, user, pass, fromFolder, toFolder string) (*ImapClient, error) {
	c, err := client.DialTLS(host, nil)
	if err != nil {
		return nil, err
	}

	updates := make(chan client.Update, 100)
	c.Updates = updates

	err = c.Login(user, pass)
	if err != nil {
		c.Logout()
		return nil, err
	}

	_, err = c.Select(fromFolder, false)
	if err != nil {
		c.Logout()
		return nil, err
	}

	return &ImapClient{
		cli:        c,
		updates:    updates,
		fromFolder: fromFolder,
		toFolder:   toFolder,
	}, nil
}

func (ic *ImapClient) Close() {
	ic.cli.Logout()
}

func (ic *ImapClient) Poll() ([]*Task, error) {
	mbox, err := ic.cli.Status(ic.fromFolder, []imap.StatusItem{imap.StatusMessages})
	if err != nil {
		return nil, err
	}

	ic.drainUpdates()

	if mbox.Messages < 1 {
		return []*Task{}, nil
	}

	seqset := &imap.SeqSet{}
	seqset.AddRange(1, mbox.Messages)

	section := &imap.BodySectionName{}

	msgs, err := ic.Fetch(seqset, []imap.FetchItem{imap.FetchEnvelope, imap.FetchUid, section.FetchItem()})
	if err != nil {
		return nil, err
	}

	ret := []*Task{}

	for _, msg := range msgs {
		m, err := mail.ReadMessage(msg.GetBody(section))
		if err != nil {
			return nil, err
		}

		d, err := m.Header.Date()
		if err != nil {
			return nil, err
		}

		wd := &mime.WordDecoder{}
		s, err := wd.DecodeHeader(m.Header.Get("Subject"))
		if err != nil {
			return nil, err
		}

		ret = append(ret, &Task{
			Name: s,
			HtmlNotes: fmt.Sprintf(
				"<body>From: %s\nTo: %s\nDate: %s</body>",
				ic.escape(m.Header.Get("From")),
				ic.escape(m.Header.Get("To")),
				ic.escape(d.Local().Format("Monday, 2006-01-02 15h04 -0700")),
			),
			Uid: msg.Uid,
		})
	}

	return ret, nil
}

func (ic *ImapClient) Archive(task *Task) error {
	seqset := &imap.SeqSet{}
	seqset.AddNum(task.Uid)
	return ic.Move(seqset, ic.toFolder)
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

func (ic *ImapClient) Move(seqset *imap.SeqSet, dest string) error {
	return ic.cli.UidMove(seqset, dest)
}

func (ic *ImapClient) Wait() error {
	stop := make(chan struct{})
	done := make(chan error, 1)
	go func() {
		done <- ic.cli.Idle(stop, &client.IdleOptions{
			LogoutTimeout: -1,
			PollInterval:  -1,
		})
	}()

	select {
	case err := <-done:
		// Never times out, so error
		return err

	case <-time.After(time.Duration(rand.Intn(60)) * time.Second):
		close(stop)
		<-done
		return nil

	case <-ic.updates:
		close(stop)
		<-done
		return nil
	}
}

func (ic *ImapClient) drainUpdates() {
	for {
		select {
		case <-ic.updates:

		default:
			return
		}
	}
}

func (ic *ImapClient) escape(in string) string {
	in = strings.ReplaceAll(in, "<", "&lt;")
	in = strings.ReplaceAll(in, ">", "&gt;")
	in = strings.ReplaceAll(in, `"`, "&quot;")
	return in
}
