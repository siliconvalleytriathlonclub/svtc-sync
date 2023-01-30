## NAME

svtc-sync - Synchronization of various platforms with a club membership reference

## SYNOPSIS

    svtc-sync [-h] [-out NF|NA] [-ref file] (strava|slack)

## DESCRIPTION

SVTC has a need to periodically check which individuals that are not paying club members might be subscribed or affilated with various platforms the club uses to provide member benefits. 
Such individuals should be identified and contacted to remind them of opportunities to become members. 

This cli tool will use public APIs of platforms such as Strava and Slack to fetch lists of users affiliated with SVTC and compare their user data to a reference file provided by the ClubExpress management platform. The data of this platform is regarded to be the source of truth for membership data, such as contact information and status.

svtc-sync will take the reference file as an optional parameteri. The default is ClubExpressMemberList.csv in the current working directory. This file is expected to be a valid comma separated value (CSV) format file with a header in the form of:

    firstname,middle,lastname,email,status,expiration
    Dave,TheMan,Scott,dave@gmail.com,Expired,12/31/22

The order of header/fields does not matter.

Validation of the specified CSV file is performed prior to further processing. It returns a list of parsing errros if encountered and will exit.

The tool will take an argument specifying the platform to verify users for. Currently values can be either "strava" or "slack". This command line argument is non optional.

Optionally, an output filter may be applied to determine whether to show Not Active (NA) ot Not Found (NF) records. Default being to show both.

Usage information can be obtained via the flag -h or -help.

## EXAMPLE USE

Specify a reference member data file and validate Strava SVTC club athletes against it.

    svtc-sync -ref /usr/local/etc/MemberData.csv strava

Use the default reference member csv file and validate Slack SVTC workspace users against it. Output all Slack users that are found in the reference file, but are listed as not active.

    svtc-sync -out NA slack

Get usage information (same as -h, --h or -help).

    svtc-sync --help


