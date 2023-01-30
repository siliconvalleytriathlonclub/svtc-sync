package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
)

type config struct {
	mfile        string // Master CSV file of current Club Members
	source       string // Source data to check against master Member reference
	output       string // NF (not Found), Active status or nil
	athleteid    int    // Strava user / athlete sctc-sync@svtriclub.org
	clubid       int    // Strava Club Silicon Valley Triathlon Club
	ucfilestrava string // Strava User API credentials JSON file
	bcfileslack  string // Slack Bot API credentials JSON file
	ccfile       string // Strava and Slack Client credentials JSON file
}

type application struct {
	creds     *CredsModel     // API Credentials
	clubAPI   *ClubAPIModel   // Strava Club API object
	clubCSV   *ClubCSVModel   // ClubExpress CSV member data
	workspace *WorkspaceModel // Slack Web API workspace data
}

func main() {

	// --------------------------------------------------------------------------------------------

	var cfg config

	// Assign user supplied reference file or use default
	flag.StringVar(&cfg.mfile, "ref", "./ClubExpressMember.csv", "Reference CSV file of current Club Members")

	// Filter output to show either Not Found (NF) or Not Active (NA) records only
	// Custom function to allow to validate multiple option values for a flag
	flag.StringVar(&cfg.output, "out", "", "Output only Not Found or Not Active records")

	// Custom usage output, override standard flag.Usage function
	flag.Usage = func() {
		fmt.Printf("Usage: svtc-sync [-h] [-out NF|NA] [-ref file] (strava|slack) \n")
	}

	flag.Parse()

	// Check if output option is in the list of supported options, print usage info and exit if not
	err := CheckArgs(&cfg.output, cfg.output, []string{"NF", "NA", ""})
	if err != nil {
		flag.Usage()
		os.Exit(0)
	}

	// Assign last cli argument as operator to specify source of member data to validate.
	// Exit and print usage inflo if not specified. Check if operator is in list of supported platforms,
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

	cfg.athleteid = 112729399
	cfg.clubid = 449951
	cfg.ucfilestrava = "./.secret/user_creds_strava.json"
	cfg.bcfileslack = "./.secret/bot_creds_slack.json"
	cfg.ccfile = "./.secret/api_creds.json"

	// --------------------------------------------------------------------------------------------

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

	netClient := &http.Client{}

	// --------------------------------------------------------------------------------------------

	app := application{
		creds:     &CredsModel{Client: netClient},
		clubAPI:   &ClubAPIModel{Client: netClient, ClubID: cfg.clubid},
		clubCSV:   &ClubCSVModel{File: file},
		workspace: &WorkspaceModel{Client: netClient},
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
		mlSlack, err := app.workspace.UserList(slack_access_token)
		if err != nil {
			log.Printf("[ListUsers] %s", err)
			return
		}
		log.Printf("[main] Request list of %d workspace users from Slack web api", len(mlSlack))

		// Iterate over list of Slack workspace users/members and check if present in reference member list
		for _, mSlack := range mlSlack {

			// Ignore bot or app records (this flag is set to false for those in Slack)
			if !mSlack.Is_Email_Confirmed {
				continue
			}

			// Check if record is member, based on criteria implemented in this function
			m := app.clubCSV.CheckMember(mlCSV, string("slack"), mSlack)

			// Determine output based on configuration settings
			if m == nil {
				if cfg.output == "NF" || cfg.output == "" {
					fmt.Printf("%s %s (%s) - Not Found \n", mSlack.Profile.FirstName, mSlack.Profile.LastName, mSlack.Profile.Email)
				}
			} else {
				if m.Status != "Active" && (cfg.output == "NA" || cfg.output == "") {
					fmt.Printf("%s %s (%s) - %s on %s\n", m.FirstName, m.LastName, m.Email, m.Status, m.Expired)
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
		cStrava, err := app.clubAPI.Club(strava_access_token)
		if err != nil {
			log.Printf("[Get] %s", err)
			return
		}
		log.Printf("[main] Request data from Strava api for %s \n", cStrava.Name)

		// Get list of athletes (club members) of Stava club
		mlStrava, err := app.clubAPI.AthleteList(cStrava.MemberCount, strava_access_token)
		if err != nil {
			log.Printf("[ListAthletes] %s", err)
			return
		}
		log.Printf("[main] Request list of %d club athletes from Strava api", len(mlStrava))

		// Iterate over list of Strava club athletes and check if present in reference member list
		for _, mStrava := range mlStrava {

			// Check if record is member, based on criteria implemented in this function
			m := app.clubCSV.CheckMember(mlCSV, string("strava"), mStrava)

			// Determine output based on configuration settings
			if m == nil {
				if cfg.output == "NF" || cfg.output == "" {
					fmt.Printf("%s %s - Not Found \n", mStrava.FirstName, mStrava.LastName)
				}
			} else {
				if m.Status != "Active" && (cfg.output == "NA" || cfg.output == "") {
					fmt.Printf("%s %s (%s) - %s on %s \n", m.FirstName, m.LastName, m.Email, m.Status, m.Expired)
				}
			}

		}

	}

	os.Exit(0)

}
