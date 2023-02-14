package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"sort"
	"strings"
)

type SlackAPIModel struct {
	Client *http.Client
}

// --------------------------------------------------------------------------------------------

func (m *SlackAPIModel) UserList(access_token string) ([]Member, error) {

	url := "https://slack.com/api/users.list"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("creation of new GET request to Slack web api failed: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+access_token)

	resp, err := m.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("GET request to Slack web api failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("unable to read result of call to Slack web api: %w", err)
	}

	response := ResponseMember{}

	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, fmt.Errorf("unmarshal json data failed: %w", err)
	}

	// Handle rerror conditions, independent of HTTP return codes
	if !response.Ok {
		err = errors.New(response.Error)
		return nil, fmt.Errorf("non-OK response status from Slack web api: %w", err)
	}

	return response.Members, nil

}

// --------------------------------------------------------------------------------------------

// Sorts a Slack Workspace User slice by the First Name field in descending order
func (m *SlackAPIModel) Sort(ml []Member) {

	sort.Slice(ml, func(i, j int) bool {
		return strings.ToLower(ml[i].Profile.FirstName) < strings.ToLower(ml[j].Profile.FirstName)
	})

}
