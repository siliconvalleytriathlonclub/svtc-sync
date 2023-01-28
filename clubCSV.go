package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/gocarina/gocsv"
)

type ClubCSVModel struct {
	File *os.File
}

// --------------------------------------------------------------------------------------------

func (m *ClubCSVModel) MemberList() ([]*MemberSVTC, error) {

	ml := []*MemberSVTC{}

	err := gocsv.UnmarshalFile(m.File, &ml)
	if err != nil {
		return nil, fmt.Errorf("unmarshal csv data failed: %w", err)
	}

	return ml, nil

}

// --------------------------------------------------------------------------------------------

func (m *ClubCSVModel) Validate() bool {

	return true

}

// --------------------------------------------------------------------------------------------

func (m *ClubCSVModel) CheckMember(reference []*MemberSVTC, source string, data interface{}) *MemberSVTC {

	switch source {

	case "slack":

		// Assert type to Member for Slack workspace users
		mSlack := data.(Member)

		// Match on either firstname and lastname or match on email matches
		for _, m := range reference {
			if (strings.EqualFold(mSlack.Profile.FirstName, m.FirstName) &&
				strings.EqualFold(mSlack.Profile.LastName, m.LastName)) ||
				strings.EqualFold(mSlack.Profile.Email, m.Email) {
				return m
			}
		}

	case "strava":

		// Assert type to Athlete for Strava club members
		mStrava := data.(Athlete)

		// Trim leading or trailing white space from Strava names
		// Match on firstname and first letter lastname
		for _, m := range reference {
			if strings.EqualFold(strings.TrimSpace(mStrava.FirstName), m.FirstName) &&
				strings.EqualFold(strings.TrimSpace(string(mStrava.LastName[0])), string(m.LastName[0])) {
				return m
			}
		}

	}

	return nil

}

// --------------------------------------------------------------------------------------------
