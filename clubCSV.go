package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"sort"
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

	// Iterate over refrence data to apply filter criteria
	for _, m := range reference {

		if cfg.output == "EXP" {
			// Ignore records that are not expired status
			if m.Status != "Expired" {
				continue
			}
			if cfg.expire != "1/1/01" {
				// Ignore records with expire dates preceeding config exp date
				if GetDate(m.Expired).Before(GetDate(cfg.expire)) {
					continue
				}
			}
		}

		switch source {

		case "slack":

			// Assert type to Member for Slack workspace users
			mSlack := data.(Member)

			// Match on either firstname and lastname or match on email
			if !((strings.EqualFold(mSlack.Profile.FirstName, m.FirstName) &&
				strings.EqualFold(mSlack.Profile.LastName, m.LastName)) ||
				strings.EqualFold(mSlack.Profile.Email, m.Email)) {

				continue
			}

		case "strava":

			// Assert type to Athlete for Strava club members
			mStrava := data.(Athlete)

			// Trim leading or trailing white space from Strava names
			// Match on firstname and first letter lastname
			if !(strings.EqualFold(strings.TrimSpace(mStrava.FirstName), m.FirstName) &&
				strings.EqualFold(strings.TrimSpace(string(mStrava.LastName[0])), string(m.LastName[0]))) {

				continue
			}

		}

		ml = append(ml, m)

	}

	return m.Sort(ml, "exp")

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

// Sorts a MemberSVTC slice by the Expired date field in descending order
func (m *ClubCSVModel) Sort(ml []*MemberSVTC, sortby string) []*MemberSVTC {

	switch sortby {

	case "exp":
		sort.Slice(ml, func(i, j int) bool {
			return GetDate(ml[i].Expired).After(GetDate(ml[j].Expired))
		})

	case "lname":
		sort.Slice(ml, func(i, j int) bool {
			return ml[i].LastName < ml[j].LastName
		})

	}

	return ml
}

// --------------------------------------------------------------------------------------------
