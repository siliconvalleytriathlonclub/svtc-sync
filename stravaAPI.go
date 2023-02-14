package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"sort"
	"strconv"
	"strings"
)

type StravaAPIModel struct {
	Client *http.Client
	ClubID int
}

// --------------------------------------------------------------------------------------------

func (m *StravaAPIModel) Club(access_token string) (*Club, error) {

	// https://www.strava.com/api/v3/clubs/{id}
	url := "https://www.strava.com/api/v3/clubs/"
	url += strconv.Itoa(m.ClubID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("creation of new GET request to Strava api failed: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+access_token)

	resp, err := m.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("GET request to Strava api failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 401 {
		return nil, fmt.Errorf("401: request to Strava api not authorized")
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("unable to read result of call to strava api: %w", err)
	}

	// A non-200 return code does not cause an error on GET, e.g. for a malformed request string
	if resp.StatusCode != 200 {
		err = errors.New(string(body))
		return nil, fmt.Errorf("non-200 response from strava api: %w", err)
	}

	club := &Club{}

	err = json.Unmarshal(body, club)
	if err != nil {
		return nil, fmt.Errorf("unmarshal json data failed: %w", err)
	}

	return club, nil

}

// --------------------------------------------------------------------------------------------

func (m *StravaAPIModel) AthleteList(count int, access_token string) ([]Athlete, error) {

	// https://www.strava.com/api/v3/clubs/{id}/members
	url := "https://www.strava.com/api/v3/clubs/"
	url += strconv.Itoa(m.ClubID)
	url += "/members"
	url += "?page=1"
	url += "&per_page=" + strconv.Itoa(count)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("creation of new GET request to Strava api failed: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+access_token)

	resp, err := m.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("GET request to Strava api failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 401 {
		return nil, fmt.Errorf("401: request to Strava api not authorized")
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("unable to read result of call to Strava api: %w", err)
	}

	// A non-200 return code does not cause an error on GET, e.g. for a malformed request string
	if resp.StatusCode != 200 {
		err = errors.New(string(body))
		return nil, fmt.Errorf("non-200 response from Strava api: %w", err)
	}

	al := []Athlete{}

	err = json.Unmarshal(body, &al)
	if err != nil {
		return nil, fmt.Errorf("unmarshal json data failed: %w", err)
	}

	return al, nil

}

// --------------------------------------------------------------------------------------------

// Sorts Strava Athlete slice by the First name field in descending order
func (m *StravaAPIModel) Sort(al []Athlete) {

	sort.Slice(al, func(i, j int) bool {
		return strings.ToLower(al[i].FirstName) < strings.ToLower(al[j].FirstName)
	})

}
