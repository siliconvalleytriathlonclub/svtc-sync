package models

// ------------------------------------------------------------------------------------------------
type Creds struct {
	Strava  StravaCreds  `json:"strava"`
	Slack   SlackCreds   `json:"slack"`
	Express ExpressCreds `json:"clubexpress"`
}

// ------------------------------------------------------------------------------------------------

// The Strava credentials data struct holds the data that is returned by the Strava http://www.strava.com/oauth/token request.
// Note the grant_type of "refresh_token" among other POST key/value pairs.
type StravaCreds struct {
	Client_ID     int    `json:"client_id"`     // Application ID, read from api_credentials.json file
	Client_Secret string `json:"client_secret"` // Application Secret, read from api_credentials.json file
	Refresh_Token string `json:"refresh_token"` // Used to refresh Access Tokem
	Access_Token  string `json:"access_token"`  // Transient Access Token
	Expires_At    int    `json:"expires_at"`    // Access Token expiration date/time
	Expires_In    int    `json:"expires_in"`    // Access Token expiration in epoch from time of grant
}

type SlackCreds struct {
	App_ID         string `json:"app_id"`
	Client_ID      string `json:"client_id"`      // Application ID, read from api_credentials.json file
	Client_Secret  string `json:"client_secret"`  // Application Secret, read from api_credentials.json file
	Signing_Secret string `json:"signing_secret"` // Secret to verify api responses
	Access_Token   string `json:"access_token"`   // Permanent Access Token
}

type ExpressCreds struct {
	Club_ID    string `json:"club_id"`    // SVTC Club ID
	Access_Key string `json:"access_key"` // Assigned Access Key
}

// ------------------------------------------------------------------------------------------------

// Strava Athlete data structure is used to hold response data from the Strava List Club Members (getClubMembersById) API call
// which returns an array of limited relevant Strava Athlete data.
// Currently only first name and the initial of last name are needed (among other data that is not used here).
type Athlete struct {
	FirstName string `json:"firstname"`
	LastName  string `json:"lastname"`
}

// Strava Club data structure is used to hold response data from the Strava Get Club (getClubById) API call
// which returns a DetailedClub object.
// Note only a subset of retiurned data is used/needed for this application
type Club struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	MemberCount int    `json:"member_count"`
}

// ------------------------------------------------------------------------------------------------

var StatusMap = map[string]string{
	"EXP": "Expired",
	"ACT": "Active",
	"TRI": "Trial",
}

// Core member structure, modeled after the ClubExpress csv export and JSON active member file. It is mainly used
// to create unique queries for lookup and match purposes. Non relevant fields for the time being are address and
// phone related; they are stored on new record insert, but remain unused.
type MemberSVTC struct {
	ID        int    `json:"id"`            // sql: id INTEGER
	Num       string `json:"memberNumber"`  // sql: num INTEGER
	Active    bool   `json:"active"`        // sql: active INTEGER
	Login     string `json:"loginName"`     // sql: login TEXT
	FirstName string `json:"firstName"`     // sql: firstname TEXT
	Middle    string `json:"middleInitial"` // sql: middle TEXT
	LastName  string `json:"lastName"`      // sql: lastname TEXT
	Email     string `json:"email"`         // sql: email TEXT
	Status    string `json:"status"`        // sql: status TEXT
	Joined    string `json:"joined"`        // sql: joined TEXT
	Expired   string `json:"expired"`       // sql: expired TEXT
	Address   string `json:"address1"`      // sql: address TEXT
	AddrExt   string `json:"address2"`      // sql: addr_ext TEXT
	City      string `json:"city"`          // sql: city TEXT
	State     string `json:"state"`         // sql: state TEXT
	Zip       string `json:"zip"`           // sql: zip INTEGER
	Mobile    string `json:"cellPhone"`     // sql: mobile TEXT
	Phone     string `json:"phone"`         // sql: phone TEXT
	// MemberID  string `json:"profileLink"`   // sql: clubexpress_id INTEGER
}

// Structure to support the use and mapping of name aliases for SVTC Members
// Aliases can be specified for First- Lastname and Email and are tied to a specific Member ID
type MemberAlias struct {
	ID        int    // sql: id
	MemberID  int    // sql: member_id INTEGER
	FirstName string // sql: firstname TEXT
	LastName  string // sql: lastname TEXT
	Email     string // sq;: email TEXT
}

// ------------------------------------------------------------------------------------------------

// Slack Workspace Member / User data structures as returned by the user.list request to their web api.
// These fields represent a subset of the data returned, as required for this application.
type ResponseMember struct {
	Ok      bool     `json:"ok"`
	Error   string   `json:"error"`   // Resonse body contains only this when !ok
	Members []Member `json:"members"` // Response body when ok
}

type Member struct {
	ID                 string  `json:"id"`
	Team_ID            string  `json:"team_id"`
	Name               string  `json:"name"`
	Profile            Profile `json:"profile"`
	Is_Email_Confirmed bool    `json:"is_email_confirmed"`
}

type Profile struct {
	FirstName string `json:"first_name"` // First name of team member
	LastName  string `json:"last_name"`  // Last name of team member
	Email     string `json:"email"`      // Team Member Email
}
