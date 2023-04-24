package api

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

	"svtc-sync/pkg/models"
)

type CredsModel struct {
	Client *http.Client
}

const (
	cfileStrava  = "./.secret/user_creds_strava.json"
	cfileSlack   = "./.secret/bot_creds_slack.json"
	cfileExpress = "./.secret/club_creds_express.json"
	cfileAPI     = "./.secret/api_creds.json"
)

// --------------------------------------------------------------------------------------------

func (m *CredsModel) CheckStravaExp() (string, error) {

	c, err := m.ReadStravaUserCreds()
	if err != nil {
		return "", fmt.Errorf("unable to read Strava user credentials %w", err)
	}

	if c.Expires_At < int(time.Now().Unix()) {

		rc, err := m.RefreshStravaAccess(c.Refresh_Token)
		if err != nil {
			return "", fmt.Errorf("refresh Strava access failed: %w", err)
		}
		log.Printf("[RefreshStravaAccess] Refreshed Strava Access Token \n")

		err = m.WriteStravaUserCreds(rc)
		if err != nil {
			return "", fmt.Errorf("write Strava user creds failed: %w", err)
		}
		log.Printf("[WriteStravaUserCreds] Saved new Strava User Creds to file \n")

		return rc.Access_Token, nil

	}

	return c.Access_Token, nil
}

// --------------------------------------------------------------------------------------------

func (m *CredsModel) RefreshStravaAccess(refresh_token string) (*models.StravaCreds, error) {

	creds, err := m.ReadStravaClientCreds()
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

func (m *CredsModel) ReadStravaClientCreds() (*models.StravaCreds, error) {

	creds := &models.Creds{}

	file, err := os.Open(cfileAPI)
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

func (m *CredsModel) ReadStravaUserCreds() (*models.StravaCreds, error) {

	creds := &models.StravaCreds{}

	file, err := os.Open(cfileStrava)
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

func (m *CredsModel) WriteStravaUserCreds(creds *models.StravaCreds) error {

	data, err := json.MarshalIndent(creds, "", " ")
	if err != nil {
		return fmt.Errorf("marshal json data failed: %w", err)
	}

	err = ioutil.WriteFile(cfileStrava, data, 0644)
	if err != nil {
		return fmt.Errorf("file write failed: %w", err)
	}

	return nil
}

// --------------------------------------------------------------------------------------------

func (m *CredsModel) GetSlackAccess() (string, error) {

	creds, err := m.ReadSlackBotCreds()
	if err != nil {
		return "", fmt.Errorf("unable to read Slack user credentials %w", err)
	}

	return creds.Access_Token, nil
}

// --------------------------------------------------------------------------------------------

func (m *CredsModel) ReadSlackBotCreds() (*models.SlackCreds, error) {

	creds := &models.SlackCreds{}

	file, err := os.Open(cfileSlack)
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
/*
	// Read ClubExpress Access Token from credentials file.
	express_access_key, err := app.creds.GetExpressAccess(cfg.ccfileexpress)
	if err != nil {
		log.Printf("[GetExpressAccess] Unable to read ClubeExpress credentials %s", err)
		return
	}
	log.Printf("[main] Retrieved ClubExpress api access key from file [%s] \n", express_access_key)
*/
//

func (m *CredsModel) GetExpressAccess() (string, error) {

	creds, err := m.ReadExpressCreds()
	if err != nil {
		return "", fmt.Errorf("unable to read ClubExpress credentials %w", err)
	}

	return creds.Access_Key, nil
}

// --------------------------------------------------------------------------------------------

func (m *CredsModel) ReadExpressCreds() (*models.ExpressCreds, error) {

	creds := &models.ExpressCreds{}

	file, err := os.Open(cfileExpress)
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
