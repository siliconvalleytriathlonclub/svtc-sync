package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"svtc-sync/app"
	"svtc-sync/pkg/helpers"

	"svtc-sync/pkg/models/api"
	"svtc-sync/pkg/models/sqlite"

	_ "github.com/mattn/go-sqlite3"
)

func main() {

	// --------------------------------------------------------------------------------------------

	var cfg app.Configuration

	// Specify user supplied reference Sqlite3 DB file or use default
	flag.StringVar(&cfg.DBfile, "db", "./svtc-sync.db", "Reference sqlite3 DB file of past and current club members")

	// Filter output to show either based on Status (EXP, ACT, TRI) or Not Found (NF) or Duplicate (DUP)records only
	flag.StringVar(&cfg.Output, "out", "", "Apply output filters to show records of specific type")

	// Ignore records with an Expire date field that is before this specified date
	flag.StringVar(&cfg.Expire, "exp", "1963-11-04", "Ignore records with an expiration prior to this date")

	// Flag to output only emails of members in a format that is useful for c&p into an email client
	flag.BoolVar(&cfg.Email, "email", false, "Output email client friendly records of matched members")

	// Flag to cause an update sync of active member data from ClubExpress to the reference data store
	flag.BoolVar(&cfg.Actives, "actives", false, "Option to update active members from ClubExpress")

	// Flag to output out unprocessed JSON data of active members from ClubExpress
	flag.BoolVar(&cfg.Raw, "raw", false, "Option to output raw active member JSON data from ClubExpress")

	// Flag to output only result of active member sync, NOT commit updates to DB
	flag.BoolVar(&cfg.Preview, "pre", false, "Option to only preview results of active member sync")

	// Custom usage output, override standard flag.Usage function
	flag.Usage = func() {
		fmt.Printf("Usage: \n")
		fmt.Printf("  svtc-sync -h \n")
		fmt.Printf("  svtc-sync [-db file] -actives [-raw] [-pre] \n")
		fmt.Printf("  svtc-sync [-db file] (ref|alias) \n")
		fmt.Printf("  svtc-sync [-db file] [-out NF|DUP] (strava|slack) \n")
		fmt.Printf("  svtc-sync [-db file] [-out EXP|ACT|TRI] [-exp date] [-email] (strava|slack) \n")
	}

	flag.Parse()

	infoLog := log.New(os.Stdout, "INFO ", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stderr, "ERROR ", log.Ldate|log.Ltime)

	// Check if output option is in the list of supported options, print usage info and exit if not
	err := helpers.CheckArgs(&cfg.Output, cfg.Output, []string{"NF", "DUP", "EXP", "TRI", "ACT", ""})
	if err != nil {
		flag.Usage()
		os.Exit(0)
	}

	// Validate expire option to ensure it is in the expected date format (YYYY-MM-DD)
	// Exit and print usage info if not.
	if helpers.GetDate(cfg.Expire).IsZero() {
		flag.Usage()
		os.Exit(0)
	}

	// Assign last cli argument as operator to specify source of member data to validate (unless it is to update actives)
	// Exit and print usage info if not specified.
	// Check if operator is in list of supported platforms,exit and print usage info if not
	if len(os.Args) > 1 {
		if !cfg.Actives {
			err := helpers.CheckArgs(&cfg.Source, os.Args[len(os.Args)-1], []string{"strava", "slack", "alias", "ref"})
			if err != nil {
				flag.Usage()
				os.Exit(0)
			}
		}
	} else {
		flag.Usage()
		os.Exit(0)
	}

	// --------------------------------------------------------------------------------------------

	//
	// 1. Check if DB file exists at specified path
	// 2. Open Sqlite3 DB file, generate handler
	// 3. Test connection
	_, err = os.Stat(cfg.DBfile)
	if os.IsNotExist(err) {
		errorLog.Fatal(fmt.Errorf("no DB file found: %w", err))
	}
	db, err := sql.Open("sqlite3", cfg.DBfile)
	if err != nil {
		errorLog.Fatal(fmt.Errorf("could not open DB file: %w", err))
	}
	err = db.Ping()
	if err != nil {
		errorLog.Fatal(fmt.Errorf("unable to connect to DB: %w", err))
	}
	defer db.Close()

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

	svtc_sync := app.Application{
		ErrorLog:         errorLog,
		InfoLog:          infoLog,
		Config:           &cfg,
		Creds:            &api.CredsModel{Client: netClient},
		MemberSQL:        &sqlite.MemberModel{DB: db},
		ExpressMemberAPI: &api.ExpressMemberModel{Client: netClient},
		StravaAthleteAPI: &api.StravaAthleteModel{Client: netClient},
		SlackMemberAPI:   &api.SlackMemberModel{Client: netClient},
	}

	// --------------------------------------------------------------------------------------------

	if cfg.Actives {

		if cfg.Raw {

			// Retrieve JSON file from the ClubExpress API. Output header and data as is.

			err := svtc_sync.ActivesRaw()
			if err != nil {
				svtc_sync.ErrorLog.Printf("[ActivesRaw] cannot retrieve active records from ClubExpress API: %s", err)
				os.Exit(1)
			}

		} else {

			// Function call to sync an update of active members from the ClubExpress JSON file with the
			// local Sqlite3 reference database. ActiveSync can be done directly or in preview mode dependent on
			// selected configuration flags

			err := svtc_sync.ActivesSync()
			if err != nil {
				svtc_sync.ErrorLog.Printf("[ActivesSync] unable to sync active records from ClubExpress API to DB: %s", err)
				os.Exit(1)
			}

		}

		os.Exit(0)

	}

	// --------------------------------------------------------------------------------------------

	switch cfg.Source {

	case "slack":

		// Check Slack workspace users against the current reference Sqlite3 database and
		// output results in a format that is determined by configuration flags and options.

		err = svtc_sync.CheckSlackMembers()
		if err != nil {
			svtc_sync.ErrorLog.Printf("[CheckSlackMembers] cannot check Slack members against DB: %s", err)
			os.Exit(1)
		}

	case "strava":

		// Check Strava athletes against the current reference Sqlite3 database and
		// output results in a format that is determined by configuration flags and options.

		err = svtc_sync.CheckStravaMembers()
		if err != nil {
			svtc_sync.ErrorLog.Printf("[CheckStravaMembers] cannot check Strava members against DB: %s", err)
			os.Exit(1)
		}

	case "alias":

		// Output all records from the alias table with mappings to their member records. Multiple aliases may
		// exist that point to the same member record.

		err := svtc_sync.ListAlias()
		if err != nil {
			svtc_sync.ErrorLog.Printf("[ListAlias] unable to list alias records from Reference DB: %s", err)
			os.Exit(1)
		}

	case "ref":

		// Output of all valid members, ie with an active flag set.
		// Can be useful on cmd line to pipe for further processing using grep, awk, etc.

		err := svtc_sync.ListMembers()
		if err != nil {
			svtc_sync.ErrorLog.Printf("[ListMembers] unable to list member records from Reference DB: %s", err)
			os.Exit(1)
		}

	}

	os.Exit(0)

}

// --------------------------------------------------------------------------------------------
