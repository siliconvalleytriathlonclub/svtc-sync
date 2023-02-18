package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type config struct {
	mfile         string // Master CSV file of current Club Members
	dbfile        string // SQL database reference file
	source        string // Source data to check against master Member reference
	output        string // NF (not Found), Expired status, Duplicates or nil
	expire        string // Date in the format M/D/YY to filter out earlier expire dates
	email         bool   // Emails only formatted with delimiter
	clubidexpress int    // ClubExpress Club ID for SVTC
	athleteid     int    // Strava user / athlete svtc-sync@svtriclub.org
	clubid        int    // Strava Club ID for Silicon Valley Triathlon Club
	ucfilestrava  string // Strava User API credentials JSON file
	bcfileslack   string // Slack Bot API credentials JSON file
	ccfile        string // Strava and Slack Client credentials JSON file
}

type application struct {
	creds   *CredsModel   // API Credentials
	clubCSV *ClubCSVModel // ClubExpress CSV member refrence data
	// clubSQL *ClubSQLModel // ClubExpress based SQL DB reference data
	// clubAPI   *ClubAPIModel   // ClubExpress API member data
	stravaAPI *StravaAPIModel // Strava Club API object
	slackAPI  *SlackAPIModel  // Slack Web API workspace data
}

func main() {

	// --------------------------------------------------------------------------------------------

	var cfg config

	// Assign user supplied reference file or use default
	flag.StringVar(&cfg.mfile, "ref", "./ClubExpressMemberList.csv", "Reference CSV file of current Club Members")

	// Filter output to show either Not Found (NF), Expired (EXP) or Duplicate (DUP)records only
	flag.StringVar(&cfg.output, "out", "", "Apply output filters to show records of specific type")

	// Ignore records with an Expire date field that is before this specified date
	flag.StringVar(&cfg.expire, "exp", "1/1/01", "Ignore records with an expiration prior to this date")

	// Flag to output only emails of members in a format that is useful for c&p into an email client
	flag.BoolVar(&cfg.email, "email", false, "Output email client friendly records of matched members")

	// Custom usage output, override standard flag.Usage function
	flag.Usage = func() {
		fmt.Printf("Usage: svtc-sync [-h] [-out NF|DUP|EXP] [-email] [-exp date] [-ref file] (strava|slack) \n")
	}

	flag.Parse()

	// Check if output option is in the list of supported options, print usage info and exit if not
	err := CheckArgs(&cfg.output, cfg.output, []string{"NF", "EXP", "DUP", ""})
	if err != nil {
		flag.Usage()
		os.Exit(0)
	}

	// Validate expire option to ensure it is in the expected date format (M/D/YY)
	// Exit and print usage info if not.
	if GetDate(cfg.expire).IsZero() {
		flag.Usage()
		// log.Printf("%v \n", cfg)
		os.Exit(0)
	}

	// Assign last cli argument as operator to specify source of member data to validate.
	// Exit and print usage info if not specified. Check if operator is in list of supported platforms,
	// exit and print usage info if not
	if len(os.Args) > 1 {
		err := CheckArgs(&cfg.source, os.Args[len(os.Args)-1], []string{"strava", "slack"})
		if err != nil {
			flag.Usage()
			os.Exit(0)
		}
	} else {
		flag.Usage()
		os.Exit(0)
	}

	cfg.dbfile = "./svtc-sync.db" // sqlite3 database for reference data

	cfg.clubidexpress = 325779 //SVTC Club ID for ClubExpress api requests

	cfg.athleteid = 112729399 // ID of Strava user under who this app is registered
	cfg.clubid = 449951       // Strava Club ID for SVTC

	// Credential files for various APIs.
	// Part of distribution, but not checked in to repo (.gitignore)
	cfg.ucfilestrava = "./.secret/user_creds_strava.json"
	cfg.bcfileslack = "./.secret/bot_creds_slack.json"
	cfg.ccfile = "./.secret/api_creds.json"

	// --------------------------------------------------------------------------------------------

	// Open and generate file handler for reference data file. Exit on failure.
	file, err := os.Open(cfg.mfile)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("[main] File not found: %s \n", cfg.mfile)
			os.Exit(1)
		} else {
			log.Fatal(err) // log print and exit(1)
		}
	}
	defer file.Close()

	//
	// Open Sqlite3 DB file, generate handler and test connect. Exit on failure.
	/*
		db, err := sql.Open("sqlite3", cfg.dbfile)
		if err != nil {
			log.Fatal(err)
		}
		err = db.Ping()
		if err != nil {
			log.Fatal(err)
		}
		defer db.Close()
	*/

	//
	// Create custom http client to implement timeout handling for TCP connect (Dial), TLS
	// handshake and overall end-to-end connection duration.
	netClient := &http.Client{
		Transport: &http.Transport{
			// MaxIdleConns:        1000,
			// MaxIdleConnsPerHost: 1000,
			Dial: (&net.Dialer{
				Timeout: 5 * time.Second,
			}).Dial,
			TLSHandshakeTimeout: 5 * time.Second,
		},
		Timeout: 10 * time.Second,
	}

	// --------------------------------------------------------------------------------------------

	app := application{
		creds:   &CredsModel{Client: netClient},
		clubCSV: &ClubCSVModel{File: file},
		// clubSQL: &ClubSQLModel{DB: db},
		// clubAPI   *ClubAPIModel{Client: netClient}
		stravaAPI: &StravaAPIModel{Client: netClient, ClubID: cfg.clubid},
		slackAPI:  &SlackAPIModel{Client: netClient},
	}

	// --------------------------------------------------------------------------------------------

	// Validate CSV file against RFC4180 specs. Returns list of parsing errors with line numbers.
	ErrList, err := app.clubCSV.Validate()
	if err != nil {
		log.Printf("[Validate] Error processing file for validation: %s", err)
		return
	}

	// List CSV validation parsing errors encountered (if any) and exit program
	if len(ErrList) > 0 {
		log.Printf("[main] Parsing errors encountered validating CSV file")
		for _, e := range ErrList {
			fmt.Printf("[%d] %s \n", e.Line, e.Err.Error())
		}
		return
	}

	// Get list of members from CSV file
	mlCSV, err := app.clubCSV.MemberList()
	if err != nil {
		log.Printf("[ListMembers] %s", err)
		return
	}
	log.Printf("[main] Read list of %d club members from %s", len(mlCSV), cfg.mfile)

	// --------------------------------------------------------------------------------------------

	switch cfg.source {

	case "slack":

		// Read Slack Bot Access Token from credentials file.
		slack_access_token, err := app.creds.GetSlackAccess(cfg.bcfileslack)
		if err != nil {
			log.Printf("[ReadSlackAccess] Unable to read Slack bot credentials %s", err)
			return
		}
		log.Printf("[main] Retrieved Slack api access token from file \n")

		// Get list of Slack team members of workspace that app is installed in
		mlSlack, err := app.slackAPI.UserList(slack_access_token)
		if err != nil {
			log.Printf("[ListUsers] %s", err)
			return
		}
		log.Printf("[main] Request list of %d workspace users from Slack web api", len(mlSlack))

		// Sort workspace user list by first name (ignore upper / lowercase)
		app.slackAPI.Sort(mlSlack)
		log.Printf("[main] Sorted workspace user list alphabetically by firstname")

		// Log output type and format as appropriate
		if cfg.email && (cfg.output == "NA") {
			log.Printf("[main] Email flag set: generating output in Email address format")
		}
		log.Printf("[main] Generating %s output of matches with %s \n\n", cfg.output, cfg.mfile)

		// Iterate over list of Slack workspace users/members and check against reference member list
		for _, mSlack := range mlSlack {

			// Ignore bot or app records (this flag is set to false for those in Slack)
			if !mSlack.Is_Email_Confirmed {
				continue
			}

			// Check if record is member, based on criteria implemented in this function
			ml := app.clubCSV.CheckMember(mlCSV, cfg, string("slack"), mSlack)

			// Determine output based on configuration settings
			switch cfg.output {

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
						fmt.Printf("\t[%d] %s %s (%s) - %s [%s] \n", m.Num, m.FirstName, m.LastName, m.Email, m.Status, m.Expired)
					}
				}

			case "EXP":

				// Print records that have a non-active status. When email flag is set, print in RFC 5322 format
				if len(ml) > 0 {
					if !cfg.email {
						fmt.Printf("[%s %s (%s)] \n", mSlack.Profile.FirstName, mSlack.Profile.LastName, mSlack.Profile.Email)
					}
					for _, m := range ml {
						if cfg.email {
							fmt.Printf("%s %s <%s>,\n", m.FirstName, m.LastName, m.Email)
						} else {
							fmt.Printf("\t[%d] %s %s (%s) - %s [%s] \n", m.Num, m.FirstName, m.LastName, m.Email, m.Status, m.Expired)
						}
					}
				}

			default:

				// Print all records (incl. duplicates and not found) grouped by Slack User record
				fmt.Printf("[%s %s (%s)] \n", mSlack.Profile.FirstName, mSlack.Profile.LastName, mSlack.Profile.Email)
				for _, m := range ml {
					fmt.Printf("\t[%d] %s %s (%s) - %s [%s] \n", m.Num, m.FirstName, m.LastName, m.Email, m.Status, m.Expired)
				}

			}

		}

	case "strava":

		// Check Strava Access Token expiration date. Request new one if expired.
		log.Printf("[main] Check for expiration of Strava api access token \n")
		strava_access_token, err := app.creds.CheckStravaExp(cfg.athleteid, cfg.ucfilestrava, cfg.ccfile)
		if err != nil {
			log.Printf("[CheckStravaExp] refresh Strava authorization failed %s", err)
			return
		}

		// Get club information from Strava to obtain member count
		cStrava, err := app.stravaAPI.Club(strava_access_token)
		if err != nil {
			log.Printf("[Get] %s", err)
			return
		}
		log.Printf("[main] Request data from Strava api for %s \n", cStrava.Name)

		// Get list of athletes (club members) of Strava club
		mlStrava, err := app.stravaAPI.AthleteList(cStrava.MemberCount, strava_access_token)
		if err != nil {
			log.Printf("[ListAthletes] %s", err)
			return
		}
		log.Printf("[main] Request list of %d club athletes from Strava api", len(mlStrava))

		// Sort athlete list (club members) by first name (ignore upper/lower case)
		app.stravaAPI.Sort(mlStrava)
		log.Printf("[main] Sorted club athlete list alphabetically by firstname")

		// Log output type and format as appropriate
		if cfg.email && (cfg.output == "NA") {
			log.Printf("[main] Email flag set: generating output in Email address format")
		}
		log.Printf("[main] Generating %s output of matches with %s \n\n", cfg.output, cfg.mfile)

		// Iterate over list of Strava club athletes and check if present in reference member list
		for _, mStrava := range mlStrava {

			// Check if record is member, based on criteria implemented in this function
			ml := app.clubCSV.CheckMember(mlCSV, cfg, string("strava"), mStrava)

			// Determine output based on configuration settings
			switch cfg.output {

			case "NF":

				// Print record not found in reference
				if len(ml) == 0 {
					fmt.Printf("[%s %s] Not Found \n", mStrava.FirstName, mStrava.LastName)
				}

			case "DUP":

				if len(ml) > 1 {
					// Print all records where there is more than one match, gouped by Strava Athlete record
					fmt.Printf("[%s %s] \n", mStrava.FirstName, mStrava.LastName)
					for _, m := range ml {
						fmt.Printf("\t[%d] %s %s (%s) - %s [%s] \n", m.Num, m.FirstName, m.LastName, m.Email, m.Status, m.Expired)
					}
				}

			case "EXP":

				// Print records that have a expired status. When email flag is set, print in RFC 5322 format
				if len(ml) > 0 {
					if !cfg.email {
						fmt.Printf("[%s %s] \n", mStrava.FirstName, mStrava.LastName)
					}
					for _, m := range ml {
						if cfg.email {
							fmt.Printf("%s %s <%s>,\n", m.FirstName, m.LastName, m.Email)
						} else {
							fmt.Printf("\t[%d] %s %s (%s) - %s [%s] \n", m.Num, m.FirstName, m.LastName, m.Email, m.Status, m.Expired)
						}
					}
				}

			default:

				// Print all records (incl. duplicates and not found) grouped by Strava Athlete record
				fmt.Printf("[%s %s] \n", mStrava.FirstName, mStrava.LastName)
				for _, m := range ml {
					fmt.Printf("\t[%d] %s %s (%s) - %s [%s] \n", m.Num, m.FirstName, m.LastName, m.Email, m.Status, m.Expired)
				}

			}

		}

	}

	/*
		err = Email("Alles Klar!")
		if err != nil {
			log.Printf("[Email] %s", err)
			return
		}
	*/

	os.Exit(0)

}

// --------------------------------------------------------------------------------------------
