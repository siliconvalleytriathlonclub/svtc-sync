package main

import (
	"encoding/csv"
	"fmt"
	"io"
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

// Validation of csv file according to RFC 4180.
// Returns an array of parsing errors, which includes the error msg and the record number of the error
// Also returns an error in case a read error was unable to be asserted to a parsing error type.
func (m *ClubCSVModel) Validate() ([]csv.ParseError, error) {

	r := csv.NewReader(m.File)
	r.FieldsPerRecord = 0 // Records must have the same number of fields as the first one read
	r.LazyQuotes = true   // A quote may appear in an unquoted field and a non-doubled quote may appear in a quoted field

	ErrList := []csv.ParseError{}

	for {
		_, err := r.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			pe, ok := err.(*csv.ParseError)
			if !ok {
				return nil, err // Return with no error list if parsing error
			}
			ErrList = append(ErrList, *pe)
			continue
		}
	}

	// Reset pointer to beginning of file
	m.File.Seek(0, io.SeekStart)

	return ErrList, nil

}

// --------------------------------------------------------------------------------------------
