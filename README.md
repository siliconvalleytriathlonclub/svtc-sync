## NAME

`svtc-sync` - Synchronization of various platforms with a club membership reference

## SYNOPSIS

    svtc-sync [-h] [-ref file] [-out NF|DUP|EXP] [-exp date] [-email] (strava|slack)

## DESCRIPTION

SVTC has a need to periodically check which individuals that are not paying club members might be subscribed or affilated with various platforms the club uses to provide member benefits. 
Such individuals should be identified and contacted to remind them of opportunities to become members. 

This cli tool will use public APIs of platforms such as Strava and Slack to fetch lists of users affiliated with SVTC and compare their user data to a reference file provided by the ClubExpress management platform. The data of this platform is regarded to be the source of truth for membership data, such as contact information and status.

svtc-sync will take the reference file as an optional parameter. The default is `ClubExpressMemberList.csv` in the current working directory. This file is expected to be a valid comma separated value (CSV) format file with a header in the form of:

    firstname,middle,lastname,email,status,expiration
    Dave,TheMan,Scott,dave@gmail.com,Expired,12/31/22

The listed fields represent the minimum required fields, but may contain more. The order of header/fields does not matter.

Validation of the specified CSV file is performed prior to further processing and returns a list of parsing errors if encountered and will exit.

The tool will take a mandatory argument specifying the platform to verify users for. Currently values can be either "strava" or "slack".

Optionally, an output filter may be applied to specify whether to only show Expired (EXP), Not Found (NF) or Duplicate (DUP) records. Default behavior is to output all records that match.

When the EXP output is used, a user may optionally specify a date prior to which records should be ignored. This date is compared to the `Expired` date in the source file that reflects when a membership has either expired or shall expire. The date remains ignored for other output options.

The tool will generate an aphabetically ordered list of matched reference records (multiple if applicable) in the following format; - unless the -email option is specified (see below).

    [source record]
        [num] name (email) - status [expired date]

To support simplified cut and paste of email addresses into an email client, the user may optionally specifiy the -email flag. This will output records in RFC 5322 conform format, e.g.

    Paula Newby-Frasure <queenofkona@gmail.com>,

This option is only available in combination with the EXP (Expired) output option. For other output types it remains ignored.

Usage information can be obtained via the flag -h or -help.

## EXAMPLE USE

Specify a reference member data file and validate Strava SVTC club athletes against it.

    svtc-sync -ref /usr/local/etc/MemberData.csv strava

Use the default reference member csv file and validate Slack SVTC workspace users against it. Output all Slack users that are found in the reference file, but are listed as not active.

    svtc-sync -out EXP slack

Output all member records from the default reference that match data coming from Strava and yield duplicate matches. (Note: output format includes the platform record and all matches)

    svtc-sync -out DUP strava

List all matching Strava records with an expired date after Jan-01-2020. List only non active members.

    svtc-sync -out EXP -exp 1/1/20 strava

List all non-active Slack users whose membership has expired after Jun-31-2021. Print in email client friendly format.

    svtc-sync -out EXP -exp 6/30/21 -email slack

Get usage information (same as -h, --h or -help).

    svtc-sync --help

