package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"svtc-sync/pkg/models"
)

type ExpressMemberModel struct {
	Client *http.Client
}

// SVTC Club ID for ClubExpress api requests
const ClubIDexpress = 325779 // SVTC Club ID for ClubExpress api requests

// URL to ClubExpress JSON file containing currently active member data
const activesurl = "https://s3.amazonaws.com/ClubExpressClubFiles/325779/json/wremawat.json"

// --------------------------------------------------------------------------------------------

// Function to call the ClubExpress member_status API endpoint t in order to obtain the current
// status of a member record.
//
// Possible return values acc. to their documentation are:
// 0 = internal error
// 1 = active
// 2 = expired
// 3 = dropped
// 5 = frozen
// 6 = bulk loaded
// 7 = pending
// 8 = prospective
// 10 = trial
// -1 = more than one match found (applies to email lookup only)
// -2 = no match found
// -3 = invalid request
//
// We are currently not using this api due its limited scope, but the code has been developed and tested
// to enable it, including storage and reading of the api credentials (access_key).
func (m *ExpressMemberModel) GetStatus(Num int, Email string, access_key string) (int, error) {

	// https://ws.clubexpress.com/member_status.ashx?cid={clubid}&key={access_key}
	url := "https://ws.clubexpress.com/member_status.ashx"
	url += "?cid=" + strconv.Itoa(ClubIDexpress)
	url += "&key=" + access_key
	url += "&n=" + strconv.Itoa(Num)
	url += "&e=" + Email

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, fmt.Errorf("creation of new GET request to ClubExpress api failed: %w", err)
	}

	resp, err := m.Client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("GET request to ClubExpress api failed: %w", err)
	}
	defer resp.Body.Close()

	// Error if ClubID or Key is missing or they don't match
	if resp.StatusCode == 403 {
		return 0, fmt.Errorf("403: request to ClubExpress api not authorized")
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("unable to read result of call to ClubExpress api: %w", err)
	}

	ret, err := strconv.Atoi(string(body))
	if err != nil {
		return 0, fmt.Errorf("converstion of return code failed: %w", err)
	}

	return ret, nil

}

// --------------------------------------------------------------------------------------------

// Function to read a JSON file via the provided url and deserialize it into a memberSVTC struct.
// A new version of this file is placed at the AWS S3 hosted URL every day at 11:00 am GMT (3am PST)
// and contains an export of all currently active members in the ClubExpress platform for SVTC.
func (m *ExpressMemberModel) GetActives() ([]*models.MemberSVTC, error) {

	ml := []*models.MemberSVTC{}

	resp, err := http.Get(activesurl)
	if err != nil {
		return nil, fmt.Errorf("could not GET active member JSON from ClubExpress api: %w", err)
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("unable to read result of call to ClubExpress api: %w", err)
	}

	err = json.Unmarshal(data, &ml)
	if err != nil {
		return nil, fmt.Errorf("unmarshal json data failed: %w", err)
	}

	return ml, nil

}

// --------------------------------------------------------------------------------------------

// Function reads the GET request header if a supplied url and returns the assoc key-value pairs.
func (m *ExpressMemberModel) GetHeader() (map[string][]string, error) {

	resp, err := http.Get(activesurl)
	if err != nil {
		return nil, fmt.Errorf("could not GET header info from ClubExpress api: %w", err)
	}
	defer resp.Body.Close()

	return resp.Header, nil

}

// --------------------------------------------------------------------------------------------

// Function to read the GET request body of a url and return the data as raw slice of bytes.
func (m *ExpressMemberModel) GetActivesRaw() ([]byte, error) {

	resp, err := http.Get(activesurl)
	if err != nil {
		return nil, fmt.Errorf("could not GET active member JSON from ClubExpress api: %w", err)
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("unable to read result of call to ClubExpress api: %w", err)
	}

	return data, nil

}

// --------------------------------------------------------------------------------------------
