## NAME

`svtc-sync` - Synchronization of various platforms with a club membership reference

## SYNOPSIS

    svtc-sync [-h]
    svtc-sync [-db file] -actives [-raw] [-pre]
    svtc-sync [-db file] [-out NF|DUP] (strava|slack)
    svtc-sync [-db file] [-out EXP] [-exp date] [-email] (strava|slack)

## DESCRIPTION

SVTC has a need to periodically check which individuals that are not paying club members might be subscribed or affilated with various platforms the club uses to provide member benefits. Such individuals should be identified and contacted to remind them of opportunities to become members. 

This cli tool will use public APIs of platforms such as Strava and Slack to fetch lists of users affiliated with SVTC and compare their user data to a reference data set provided by the ClubExpress management platform and maintained locally as a Sqlite3 DB. The data of this platform is regarded to be the source of truth for membership data, such as contact information and status.

svtc-sync will take the specification of a Sqlite3 database file as an optional parameter. The default is `svtc-sync.db` in the current working directory. (Note: the DB is seeded via a csv export from ClubExpress in a process that is outside the scope of this tool)

There are two major functional areas in the tool:

- Check Members - Compare source member data to the reference DB and display results in specific output formats
- Sync Actives - Retrieve daily refreshed data on active members from the ClubExpress platform on demand and update the DB

Usage information can be obtained via the flag -h or -help.

### Check Members

The tool will take a mandatory argument specifying the platform to verify users for. Currently values can be either "strava" or "slack".

It will generate an aphabetically ordered list of `matched` reference records (multiple if applicable) in the following format; - unless the -email option is specified (see below).

    [source record]
        [num] name (email) - status [expired date]

Optionally, an output specifier may be applied to only show Expired (EXP), Not Found (NF) or Duplicate (DUP) records. Default behavior is to output all source records and all matches from the reference DB.

A user may optionally specify a date via the -exp flag, in the (ISO 8601) form `YYYY-MM-DD`,  prior to which records should be ignored. This date is compared to the `Expired` date in the reference data, that reflects when a membership has either expired or shall expire.

To support simplified cut and paste of email addresses into an email client, the user may optionally specifiy the -email flag. This will output records in RFC 5322 conform format, e.g.

    Paula Newby-Frasure <queenofkona@gmail.com>,

This option is only available in combination with the EXP (Expired) output option. For other output types it remains ignored.

### Sync Actives

ClubExpress regularly posts an updated export of `active` member data as a JSON format file that may be accessed via a defined URL. The svtc-sync tool will retrieve and process this file to update the current state of active members in the reference DB via the command flag `-actives`.

In order to preview the updates of active records WITHOUT comitting them to the DB, a user may specify the `-pre` flag. The preview option will produce output as follows.

    [num] name (old status) -> (new status) new exp date

With `old status` coming from the DB and `new status` from the JSON file. The `new exp date` will be set to the last day of the current year.

Optionally, to test access and the validity of the of the JSON file, a raw dump of the http header and query result in JSON format can be output via the `-raw` flag.

## EXAMPLE USE

Specify a reference Sqlite3 member data DB file and check Strava SVTC club athletes against it.

    svtc-sync -db /usr/local/etc/data.db strava

Use the default reference member DB and check Slack SVTC workspace users against it. Output all Slack users that are found in the reference file with `Expired` status.

    svtc-sync -out EXP slack

Output all member records from the default reference that match data coming from Strava and yield duplicate matches. 

    svtc-sync -out DUP strava

List only matching Strava records with with status `Expired` and an expired date after Jan-01-2022.

    svtc-sync -out EXP -exp 2022-01-01 strava

List all expired Slack users whose membership has expired after Jun-30-2021. Print in email client friendly format.

    svtc-sync -out EXP -exp 2021-06-30 -email slack

Preview updates on `Active` members coming from the ClubExpress JSON export file WITHOUT writing to DB

    svtc-sync -actives -pre

Commit updated active member records to DB

    svtc-sync -actives

Dump the entire JSON update file on active members as it is received via http

    svtc-sync -actives -raw

Get usage information (same as -h, --h or -help).

    svtc-sync --help

