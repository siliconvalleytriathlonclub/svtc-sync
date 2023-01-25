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

func (m *ClubCSVModel) IsMember(reference []*MemberSVTC, source string, data interface{}) bool {

	switch source {

	case "strava":

		mStrava := data.(Athlete)
		for _, m := range reference {
			if strings.EqualFold(mStrava.FirstName, m.FirstName) &&
				strings.EqualFold(string(mStrava.LastName[0]), string(m.LastName[0])) {
				return true
			}
		}

	case "slack":

		mSlack := data.(Member)
		for _, m := range reference {
			if (strings.EqualFold(mSlack.Profile.FirstName, m.FirstName) &&
				strings.EqualFold(mSlack.Profile.LastName, m.LastName)) ||
				strings.EqualFold(mSlack.Profile.Email, m.Email) {
				return true
			}
		}

	}

	return false

}

// --------------------------------------------------------------------------------------------