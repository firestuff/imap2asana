package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

type AsanaClient struct {
	cli       *http.Client
	token     string
	workspace string
	assignee  string
}

type addTaskRequest struct {
	Data *addTaskRequestInt `json:"data"`
}

type addTaskRequestInt struct {
	Name      string `json:"name"`
	HtmlNotes string `json:"html_notes"`
	Workspace string `json:"workspace"`
	Assignee  string `json:"assignee"`
}

func NewAsanaClient() *AsanaClient {
	return &AsanaClient{
		cli:       &http.Client{},
		token:     os.Getenv("ASANA_TOKEN"),
		workspace: os.Getenv("ASANA_WORKSPACE"),
		assignee:  os.Getenv("ASANA_ASSIGNEE"),
	}
}

func (ac *AsanaClient) CreateTask(name string, notes string) error {
	body := &addTaskRequest{
		Data: &addTaskRequestInt{
			Name:      name,
			HtmlNotes: notes,
			Workspace: ac.workspace,
			Assignee:  ac.assignee,
		},
	}

	js, err := json.Marshal(body)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", "https://app.asana.com/api/1.0/tasks", bytes.NewReader(js))
	if err != nil {
		return err
	}

	ac.addAuth(req)
	req.Header.Add("Content-Type", "application/json")

	resp, err := ac.cli.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != 201 {
		msg, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("%s: %s", resp.Status, msg)
	}

	return nil
}

func (ac *AsanaClient) addAuth(req *http.Request) {
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", ac.token))
}
