package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
)

type config struct {
	mfile           string // Master CSV file of current Club Members
	source          string // Source data to check against master Member reference
	athleteid       int    // Strava user / athlete sctc-sync@svtriclub.org
	clubid          int    // Strava Club Silicon Valley Triathlon Club
	teamid          string // Slack Team / Workspace for Silicon Valley Triathlon Club
	channelid       string // "general" Channel in SVTC Workspace
	usercredsfile   string // Strava User API credentials JSON file
	clientcredsfile string // Strava and Slack Client credentials JSON file
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
	flag.StringVar(&cfg.mfile, "ref", "./ClubExpressMemberList.csv", "Reference CSV file of current Club Members")

	// Custom usage output, override standard usage function
	flag.Usage = func() {
		fmt.Printf("Usage: svtc-sync [-ref file] [-h] (strava|slack) \n")
	}

	flag.Parse()

	// Assign last cli argument as operator to specify source of member data to compare to reference
	// If value is nil or not a valid operator (strava|slack) it is handled later in main - Fix this to be done here.
	if len(os.Args) > 1 {
		cfg.source = os.Args[len(os.Args)-1]
	}

	cfg.athleteid = 112729399
	cfg.clubid = 449951
	cfg.teamid = "T02UAHW031S"
	cfg.channelid = "C02TTL455EK"
	cfg.usercredsfile = "./.secret/user_credentials.json"
	cfg.clientcredsfile = "./.secret/api_credentials.json"

	const slack_access_token = "xoxb-2962608003060-4657504055463-7b2qlw829v2a8qHz2ffIZJYm"

	// --------------------------------------------------------------------------------------------

	file, err := os.Open(cfg.mfile)
	if err != nil {
		log.Fatal(err)
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
	// Need to do some file validation here:
	// - Is the file a CSV? Does it have proper headers? Are the columns formatted right?

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

		log.Printf("[main] Matching Slack Workspace Members \n")

		// Get list of Slack team members of workspace that app is installed in
		mlSlack, err := app.workspace.UserList(slack_access_token)
		if err != nil {
			log.Printf("[ListUsers] %s", err)
			return
		}
		log.Printf("[main] Request list of %d workspace users from Slack web api", len(mlSlack))

		// Iterate over list of Slack workspace users/members and check if present in reference member list
		for _, mSlack := range mlSlack {
			if !app.clubCSV.IsMember(mlCSV, string("slack"), mSlack) {
				fmt.Printf("%s %s (%s) \n", mSlack.Profile.FirstName, mSlack.Profile.LastName, mSlack.Profile.Email)
			}
		}

	case "strava":

		log.Printf("[main] Matching Strava Club Members \n")

		// Check Strava Access Token expiration date. Request new one if expired.
		log.Printf("[main] Check for expiration of Strava api access token \n")
		strava_access_token, err := app.creds.CheckStravaExp(cfg.athleteid, cfg.usercredsfile, cfg.clientcredsfile)
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
			if !app.clubCSV.IsMember(mlCSV, string("strava"), mStrava) {
				fmt.Printf("%s %s \n", mStrava.FirstName, mStrava.LastName)
			}
		}

	default:

		log.Printf("[main] No supported source provided. Must be (strava|slack) \n")

	}

	// --------------------------------------------------------------------------------------------

}
