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

// Function to deserialize records from a CSV file into a slice of MemberSVTC structs
func (m *ClubCSVModel) MemberList() ([]*MemberSVTC, error) {

	ml := []*MemberSVTC{}

	err := gocsv.UnmarshalFile(m.File, &ml)
	if err != nil {
		return nil, fmt.Errorf("unmarshal csv data failed: %w", err)
	}

	return ml, nil

}

// --------------------------------------------------------------------------------------------

// Function to compare a member record based on a specified source platform and data object, with a reference data set.
// Returns a result set of matches based on the specified criteria
// Config setting from the main function are passed via the cfg paramater
func (m *ClubCSVModel) CheckMember(reference []*MemberSVTC, cfg config, source string, data interface{}) []*MemberSVTC {

	ml := []*MemberSVTC{}

	switch source {

	case "slack":

		// Assert type to Member for Slack workspace users
		mSlack := data.(Member)

		// Iterate over refrence data to apply filter criteria
		for _, m := range reference {

			// Match on either firstname and lastname or match on email matches
			if (strings.EqualFold(mSlack.Profile.FirstName, m.FirstName) &&
				strings.EqualFold(mSlack.Profile.LastName, m.LastName)) ||
				strings.EqualFold(mSlack.Profile.Email, m.Email) {

				// Ignore records with Expired dates preceeding config exp date spec
				if GetDate(m.Expired).After(GetDate(cfg.expire)) {
					ml = append(ml, m)
				}

			}

		}

	case "strava":

		// Assert type to Athlete for Strava club members
		mStrava := data.(Athlete)

		// Iterate over refrence data to apply filter criteria
		for _, m := range reference {

			// Trim leading or trailing white space from Strava names
			// Match on firstname and first letter lastname
			if strings.EqualFold(strings.TrimSpace(mStrava.FirstName), m.FirstName) &&
				strings.EqualFold(strings.TrimSpace(string(mStrava.LastName[0])), string(m.LastName[0])) {

				// Ignore records with Expired dates preceeding config exp date spec
				if GetDate(m.Expired).After(GetDate(cfg.expire)) {
					ml = append(ml, m)
				}

			}

		}

	}

	return ml

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
