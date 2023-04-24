package app

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"sort"
	"strings"

	"svtc-sync/pkg/helpers"
	"svtc-sync/pkg/models"
	"svtc-sync/pkg/models/api"
	"svtc-sync/pkg/models/sqlite"

	_ "github.com/mattn/go-sqlite3"
)

// --------------------------------------------------------------------------------------------

type Configuration struct {
	DBfile  string // SQL database reference file
	Source  string // Source data to check against master Member reference
	Output  string // NF (not Found), Expired status, Duplicates or nil
	Expire  string // Date in the format M/D/YY to filter out earlier expire dates
	Email   bool   // Emails only formatted with delimiter
	Actives bool   // Get active member update from ClubExpress and sync with reference data
	Raw     bool   // Output raw JSON records as read from ClubExpress API JSON file
	Preview bool   // Option to only preview results of active member sync, ie not commit to DB
}

type Application struct {
	ErrorLog         *log.Logger
	InfoLog          *log.Logger
	Config           *Configuration
	Creds            *api.CredsModel         // API Credentials
	MemberSQL        *sqlite.MemberModel     // SVTC ClubExpress based SQL DB reference data
	ExpressMemberAPI *api.ExpressMemberModel // ClubExpress API member data
	StravaAthleteAPI *api.StravaAthleteModel // Strava Club API athlete object
	SlackMemberAPI   *api.SlackMemberModel   // Slack Web API workspace member data
}

// --------------------------------------------------------------------------------------------

func (app *Application) ActivesRaw() error {

	// Get Header information of http file Get Request and print key / value pairs
	vals, err := app.ExpressMemberAPI.GetHeader()
	if err != nil {
		app.ErrorLog.Printf("[GetHeader] %s", err)
		return err
	}
	for k, v := range vals {
		fmt.Printf("[%s] %s \n", k, v)
	}

	// Retrieve and print  raw data of active members from ClubExpress API JSON file
	data, err := app.ExpressMemberAPI.GetActivesRaw()
	if err != nil {
		app.ErrorLog.Printf("[GetActives] %s", err)
		return err
	}
	fmt.Printf("%s", string(data))

	return nil

}

// --------------------------------------------------------------------------------------------

func (app *Application) ActivesSync() error {

	// Get Header of http request and print file date: [Last-Modified][Fri, 24 Feb 2023 11:00:04 GMT]
	vals, err := app.ExpressMemberAPI.GetHeader()
	if err != nil {
		app.ErrorLog.Printf("[GetHeader] %s", err)
		return err
	}
	app.InfoLog.Printf("[ActivesSync] JSON File Date: %s \n", vals["Last-Modified"][0])

	// Get list of active members from ClubExpress API JSON file
	mlJSON, err := app.ExpressMemberAPI.GetActives()
	if err != nil {
		app.ErrorLog.Printf("[GetActives] %s", err)
		return err
	}
	app.InfoLog.Printf("[ActivesSync] Created list of %d Active club members from JSON File", len(mlJSON))

	// Iterate over list and compare records to DB, output differences
	app.InfoLog.Printf("[ActivesSync] Comparing list of active members to DB")

	// Log output type and format as appropriate
	if app.Config.Preview {
		app.InfoLog.Printf("[ActivesSync] Preview flag set: NOT making changes to DB \n")
	}

	// Get last day of the year to be used for new and updated active member expired date
	dstr := helpers.GetLastDateStr()
	app.InfoLog.Printf("[ActivesSync] Expired dates will be set to last day of this year \n\n")

	for _, m := range mlJSON {

		// Select member record by JSON file's member number.
		// If number not found, create and assign a new member record with Status as New as result
		mSQL, err := app.MemberSQL.Get(m.Num)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				mSQL = &models.MemberSVTC{Num: m.Num, Status: "New"}
			} else {
				app.ErrorLog.Printf("[Get] %s", err)
				return err
			}
		}

		// Insert New member record with Status Active and end of year as expired date
		if mSQL.Status == "New" {

			if app.Config.Preview {
				fmt.Printf("[%s] %s %s (New) -> (Active) %s \n", m.Num, m.FirstName, m.LastName, dstr)
			} else {
				m.Status = "Active"
				m.Active = true
				m.Expired = dstr
				err = app.MemberSQL.Insert(m)
				if err != nil {
					app.ErrorLog.Printf("[Insert] %s", err)
					continue
				}
				app.InfoLog.Printf("[ActivesSync] Inserted new club member with status Active: %s", m.Num)
			}

			continue
		}

		// Update existing non-active member record to reflect Active status and set expired date
		if mSQL.Status != "Active" {

			if app.Config.Preview {
				fmt.Printf("[%s] %s %s (%s) -> (Active) %s \n", m.Num, m.FirstName, m.LastName, mSQL.Status, dstr)
			} else {
				err = app.MemberSQL.UpdateStatus(m.Num, "Active", dstr)
				if err != nil {
					app.ErrorLog.Printf("[UpdateStatus] %s", err)
					continue
				}
				app.InfoLog.Printf("[ActivesSync] Updated club member. Set status to Active: %s", m.Num)
			}

			// continue
		}

	}

	return nil

}

// --------------------------------------------------------------------------------------------

func (app *Application) ListAlias() error {

	// Query alias table for members
	ml, al, err := app.MemberSQL.ListAlias()
	if err != nil {
		app.ErrorLog.Printf("[Alias SQL] %s", err)
		return err
	}

	for i, _ := range al {
		fmt.Printf("[%s %s %s] \n", al[i].FirstName, al[i].LastName, al[i].Email)
		fmt.Printf("\t[%s] %s %s (%s) \n", ml[i].Num, ml[i].FirstName, ml[i].LastName, ml[i].Email)
	}

	return nil

}

// --------------------------------------------------------------------------------------------

func (app *Application) ListMembers() error {

	// Query alias table for members
	ml, err := app.MemberSQL.ListMembers()
	if err != nil {
		app.ErrorLog.Printf("[Alias SQL] %s", err)
		return err
	}

	for _, m := range ml {
		fmt.Printf("%s %s %s %s %s %s \n", m.Num, m.FirstName, m.LastName, m.Email, m.Status, m.Expired)
	}

	return nil

}

// --------------------------------------------------------------------------------------------

func (app *Application) CheckSlackMembers() error {

	// Read Slack Bot Access Token from credentials file.
	slack_access_token, err := app.Creds.GetSlackAccess()
	if err != nil {
		app.ErrorLog.Printf("[ReadSlackAccess] Unable to read Slack bot credentials %s", err)
		return err
	}
	app.InfoLog.Printf("[CheckSlackMembers] Retrieved Slack api access token from file \n")

	// Get list of Slack team members of workspace that app is installed in
	mlSlack, err := app.SlackMemberAPI.List(slack_access_token)
	if err != nil {
		app.ErrorLog.Printf("[ListUsers] %s", err)
		return err
	}
	app.InfoLog.Printf("[CheckSlackMembers] Requested list of %d workspace users from Slack web api", len(mlSlack))

	// Sort workspace user list by first name (ignore upper / lowercase)
	app.SlackMemberAPI.Sort(mlSlack)
	app.InfoLog.Printf("[CheckSlackMembers] Sorted workspace user list alphabetically by firstname")

	// Log output type and format as appropriate
	app.InfoLog.Printf("[CheckSlackMembers] Generating %s output of matches with %s \n\n", app.Config.Output, app.Config.DBfile)

	// Iterate over list of Slack workspace users/members and check against reference member DB
	for _, mSlack := range mlSlack {

		// Ignore bot or app records (this flag is set to false for those in Slack)
		if !mSlack.Is_Email_Confirmed {
			continue
		}

		// Populate a new search member struct with query criteria
		ms := models.MemberSVTC{
			FirstName: strings.ToLower(mSlack.Profile.FirstName),
			LastName:  strings.ToLower(mSlack.Profile.LastName),
			Email:     strings.ToLower(mSlack.Profile.Email),
			Status:    models.StatusMap[app.Config.Output],
			Expired:   app.Config.Expire,
		}

		// Query list of members from Sqlite3 DB using query criteria according to (firstname AND lastname) OR email
		ml, err := app.MemberSQL.ListMatch("slack", &ms)
		if err != nil {
			app.ErrorLog.Printf("[ListMembers SQL] %s", err)
			return err
		}

		// Query alias table for members using same search criteria
		ma, err := app.MemberSQL.GetAlias(&ms)
		if err != nil {
			app.ErrorLog.Printf("[Alias SQL] %s", err)
			return err
		}

		// Append matches from alias table to result set
		ml = append(ml, ma...)

		// Sort results of comparison by expiration date
		app.sort(ml, "exp")

		// Determine output based on configuration settings and print results of comnparison
		switch app.Config.Output {

		case "NF":

			// Print record not found in reference data
			if len(ml) == 0 {
				fmt.Printf("[%s %s (%s)] Not Found \n", mSlack.Profile.FirstName, mSlack.Profile.LastName, mSlack.Profile.Email)
			}

		case "DUP":

			if len(ml) > 1 {
				// Print all records where there is more than one match, gouped by Slack User record
				fmt.Printf("[%s %s (%s)] \n", mSlack.Profile.FirstName, mSlack.Profile.LastName, mSlack.Profile.Email)
				for _, m := range ml {
					fmt.Printf("\t[%s] %s %s (%s) - %s [%s] \n", m.Num, m.FirstName, m.LastName, m.Email, m.Status, m.Expired)
				}
			}

		case "EXP", "ACT", "TRI":

			// Print records that have a non-active status. When email flag is set, print in RFC 5322 format
			if len(ml) > 0 {
				if !app.Config.Email {
					fmt.Printf("[%s %s (%s)] \n", mSlack.Profile.FirstName, mSlack.Profile.LastName, mSlack.Profile.Email)
				}
				for _, m := range ml {
					if app.Config.Email {
						fmt.Printf("%s %s <%s>,\n", m.FirstName, m.LastName, m.Email)
					} else {
						fmt.Printf("\t[%s] %s %s (%s) - %s [%s] \n", m.Num, m.FirstName, m.LastName, m.Email, m.Status, m.Expired)
					}
				}
			}

		default:

			// Print all records (incl. duplicates and not found) grouped by Slack User record
			fmt.Printf("[%s %s (%s)] \n", mSlack.Profile.FirstName, mSlack.Profile.LastName, mSlack.Profile.Email)
			for _, m := range ml {
				fmt.Printf("\t[%s] %s %s (%s) - %s [%s] \n", m.Num, m.FirstName, m.LastName, m.Email, m.Status, m.Expired)
			}

		}

	}

	return nil

}

// --------------------------------------------------------------------------------------------

func (app *Application) CheckStravaMembers() error {

	// Check Strava Access Token expiration date. Request new one if expired.
	app.InfoLog.Printf("[CheckStravaMembers] Check for expiration of Strava api access token \n")
	strava_access_token, err := app.Creds.CheckStravaExp()
	if err != nil {
		app.ErrorLog.Printf("[CheckStravaExp] refresh Strava authorization failed %s", err)
		return err
	}

	// Get club information from Strava to obtain member count
	cStrava, err := app.StravaAthleteAPI.GetClub(strava_access_token)
	if err != nil {
		app.ErrorLog.Printf("[Get] %s", err)
		return err
	}
	app.InfoLog.Printf("[CheckStravaMembers] Requested data from Strava api for %s \n", cStrava.Name)

	// Get list of athletes (club members) of Strava club
	mlStrava, err := app.StravaAthleteAPI.List(cStrava.MemberCount, strava_access_token)
	if err != nil {
		app.ErrorLog.Printf("[ListAthletes] %s", err)
		return err
	}
	app.InfoLog.Printf("[CheckStravaMembers] Requested list of %d club athletes from Strava api", len(mlStrava))

	// Sort athlete list (club members) by first name (ignore upper/lower case)
	app.StravaAthleteAPI.Sort(mlStrava)
	app.InfoLog.Printf("[CheckStravaMembers] Sorted club athlete list alphabetically by firstname")

	// Log output type and format as appropriate
	app.InfoLog.Printf("[CheckStravaMembers] Generating %s output of matches with %s \n\n", app.Config.Output, app.Config.DBfile)

	// Iterate over list of Strava club athletes and check if present in reference DB
	for _, mStrava := range mlStrava {

		// Populate a new member search struct with query criteria
		ms := models.MemberSVTC{
			FirstName: strings.ToLower(strings.TrimSpace(mStrava.FirstName)),
			LastName:  strings.ToLower(strings.TrimSpace(string(mStrava.LastName[0]))),
			Email:     string('_'),
			Status:    models.StatusMap[app.Config.Output],
			Expired:   app.Config.Expire,
		}

		// Query list of members from Sqlite3 DB using query criteria according to (firstname AND lastname%)
		ml, err := app.MemberSQL.ListMatch("strava", &ms)
		if err != nil {
			app.ErrorLog.Printf("[ListMembers SQL] %s", err)
			return err
		}

		// Query alias table for members using same criteria
		ma, err := app.MemberSQL.GetAlias(&ms)
		if err != nil {
			app.ErrorLog.Printf("[Alias SQL] %s", err)
			return err
		}

		// Append matches from alias table to result set
		ml = append(ml, ma...)

		// Sort results of comparison by expiration date
		app.sort(ml, "exp")

		// Determine output based on configuration settings and print results of comnparison
		switch app.Config.Output {

		case "NF":

			// Print record not found in reference
			if len(ml) == 0 {
				fmt.Printf("[%s %s] Not Found \n", mStrava.FirstName, mStrava.LastName)
				///fmt.Printf("insert into alias (memberid, firstname, lastname, email) values=(NNNN, '%s', '%s', '');\n", mStrava.FirstName, mStrava.LastName)
			}

		case "DUP":

			if len(ml) > 1 {
				// Print all records where there is more than one match, gouped by Strava Athlete record
				fmt.Printf("[%s %s] \n", mStrava.FirstName, mStrava.LastName)
				for _, m := range ml {
					fmt.Printf("\t[%s] %s %s (%s) - %s [%s] \n", m.Num, m.FirstName, m.LastName, m.Email, m.Status, m.Expired)
				}
			}

		case "EXP", "ACT", "TRI":

			// Print records that have a expired status. When email flag is set, print in RFC 5322 format
			if len(ml) > 0 {
				if !app.Config.Email {
					fmt.Printf("[%s %s] \n", mStrava.FirstName, mStrava.LastName)
				}
				for _, m := range ml {
					if app.Config.Email {
						fmt.Printf("%s %s <%s>,\n", m.FirstName, m.LastName, m.Email)
					} else {
						fmt.Printf("\t[%s] %s %s (%s) - %s [%s] \n", m.Num, m.FirstName, m.LastName, m.Email, m.Status, m.Expired)
					}
				}
			}

		default:

			// Print all records (incl. duplicates and not found) grouped by Strava Athlete record
			fmt.Printf("[%s %s] \n", mStrava.FirstName, mStrava.LastName)
			for _, m := range ml {
				fmt.Printf("\t[%s] %s %s (%s) - %s [%s] \n", m.Num, m.FirstName, m.LastName, m.Email, m.Status, m.Expired)
			}

		}

	}

	return nil

}

// --------------------------------------------------------------------------------------------

// Sorts a MemberSVTC slice by the Expired date field in descending order
func (app *Application) sort(ml []*models.MemberSVTC, sortby string) []*models.MemberSVTC {

	switch sortby {

	case "exp":
		sort.Slice(ml, func(i, j int) bool {
			return helpers.GetDate(ml[i].Expired).After(helpers.GetDate(ml[j].Expired))
		})

	case "lname":
		sort.Slice(ml, func(i, j int) bool {
			return ml[i].LastName < ml[j].LastName
		})

	}

	return ml
}

// --------------------------------------------------------------------------------------------
