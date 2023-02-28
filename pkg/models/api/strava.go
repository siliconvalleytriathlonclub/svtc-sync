package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"svtc-sync/pkg/models"
)

type StravaAthleteModel struct {
	Client *http.Client
}

const (
	Athleteid    = 112729399 // ID of Strava user under who this app is registered
	ClubIDstrava = 449951    // Strava Club ID for SVTC
)

// --------------------------------------------------------------------------------------------

// Function to query the Strava public Club API endpoint to obtain information on SVTC (based on the CLub ID)
func (m *StravaAthleteModel) GetClub(access_token string) (*models.Club, error) {

	// https://www.strava.com/api/v3/clubs/{id}
	url := "https://www.strava.com/api/v3/clubs/"
	url += strconv.Itoa(ClubIDstrava)

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

	club := &models.Club{}

	err = json.Unmarshal(body, club)
	if err != nil {
		return nil, fmt.Errorf("unmarshal json data failed: %w", err)
	}

	return club, nil

}

// --------------------------------------------------------------------------------------------

// Function to query the Strava public Club API endpoint to obtain a list of athletes affilated
// with SVTC (based on the CLub ID). The number of records to query is based on the number of members
// returned by the GetClub function.
func (m *StravaAthleteModel) List(count int, access_token string) ([]models.Athlete, error) {

	// https://www.strava.com/api/v3/clubs/{id}/members
	url := "https://www.strava.com/api/v3/clubs/"
	url += strconv.Itoa(ClubIDstrava)
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

	al := []models.Athlete{}

	err = json.Unmarshal(body, &al)
	if err != nil {
		return nil, fmt.Errorf("unmarshal json data failed: %w", err)
	}

	return al, nil

}

// --------------------------------------------------------------------------------------------

// Sorts Strava Athlete slice by the First name field in descending order
func (m *StravaAthleteModel) Sort(al []models.Athlete) {

	sort.Slice(al, func(i, j int) bool {
		return strings.ToLower(al[i].FirstName) < strings.ToLower(al[j].FirstName)
	})

}
