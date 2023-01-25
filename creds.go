package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

type CredsModel struct {
	Client *http.Client
}

// --------------------------------------------------------------------------------------------

func (m *CredsModel) CheckStravaExp(id int, ufile, cfile string) (string, error) {

	c, err := m.ReadStravaUserCreds(ufile)
	if err != nil {
		return "", fmt.Errorf("unable to read Strava user credentials %w", err)
	}

	if c.Expires_At < int(time.Now().Unix()) {

		rc, err := m.RefreshStravaAccess(c.Refresh_Token, cfile)
		if err != nil {
			return "", fmt.Errorf("refresh Strava access failed: %w", err)
		}
		log.Printf("[RefreshStravaAccess] Refreshed Strava Access Token for athlete ID %d \n", id)

		err = m.WriteStravaUserCreds(rc, ufile)
		if err != nil {
			return "", fmt.Errorf("write Strava user creds failed: %w", err)
		}
		log.Printf("[WriteStravaUserCreds] Saved new Strava User Creds to file \n")

		return rc.Access_Token, nil

	}

	return c.Access_Token, nil
}

// --------------------------------------------------------------------------------------------

func (m *CredsModel) RefreshStravaAccess(refresh_token string, file string) (*StravaCreds, error) {

	creds, err := m.ReadStravaClientCreds(file)
	if err != nil {
		return nil, fmt.Errorf("unable to read Strava client creds: %w", err)
	}

	req_url := "http://www.strava.com/oauth/token"

	query := map[string]string{
		"client_id":     strconv.Itoa(creds.Client_ID),
		"client_secret": creds.Client_Secret,
		"refresh_token": refresh_token,
		"grant_type":    "refresh_token",
	}

	post, err := json.Marshal(query)
	if err != nil {
		return nil, fmt.Errorf("marshal json query data failed: %w", err)
	}

	req, err := http.NewRequest("POST", req_url, bytes.NewBuffer(post))
	if err != nil {
		return nil, fmt.Errorf("creation of new POST request to Strava api failed: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := m.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("POST request to Strava api execution failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 401 {
		return nil, fmt.Errorf("401: request to Strava api not authorized")
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("unable to read result of call to Strava api: %w", err)
	}

	err = json.Unmarshal(body, &creds)
	if err != nil {
		return nil, fmt.Errorf("unmarshal json data failed: %w", err)
	}

	return creds, nil
}

// --------------------------------------------------------------------------------------------

func (m *CredsModel) ReadStravaClientCreds(fname string) (*StravaCreds, error) {

	creds := &Creds{}

	file, err := os.Open(fname)
	if err != nil {
		return nil, fmt.Errorf("file open failed: %w", err)
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("file read failed: %w", err)
	}

	err = json.Unmarshal(data, &creds)
	if err != nil {
		return nil, fmt.Errorf("unmarshal json data failed: %w", err)
	}

	return &creds.Strava, nil
}

// --------------------------------------------------------------------------------------------

func (m *CredsModel) ReadStravaUserCreds(fname string) (*StravaCreds, error) {

	creds := &StravaCreds{}

	file, err := os.Open(fname)
	if err != nil {
		return nil, fmt.Errorf("file open failed: %w", err)
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("file read failed: %w", err)
	}

	err = json.Unmarshal(data, &creds)
	if err != nil {
		return nil, fmt.Errorf("unmarshal json data failed: %w", err)
	}

	return creds, nil
}

// --------------------------------------------------------------------------------------------

func (m *CredsModel) WriteStravaUserCreds(creds *StravaCreds, fname string) error {

	data, err := json.MarshalIndent(creds, "", " ")
	if err != nil {
		return fmt.Errorf("marshal json data failed: %w", err)
	}

	// log.Printf("\n%s \n", string(data))

	err = ioutil.WriteFile(fname, data, 0644)
	if err != nil {
		return fmt.Errorf("file write failed: %w", err)
	}

	return nil
}

// --------------------------------------------------------------------------------------------
