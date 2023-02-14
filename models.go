package main

type Creds struct {
	Strava StravaCreds `json:"strava"`
	Slack  SlackCreds  `json:"slack"`
}

// ------------------------------------------------------------------------------------------------
//
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

// ------------------------------------------------------------------------------------------------

// Strava Athlete data structure is used to hold response data from the Strava List Club Members (getClubMembersById) API call
// which returns an array of limited relevant Strava Athlete data.
// Currently only first name and the initial of last name are returned.
type Athlete struct {
	FirstName string `json:"firstname"`
	LastName  string `json:"lastname"`
}

// Club data structure is used to hold response data from the Strava Get Club (getClubById) API call
// which returns a DetailedClub object.
type Club struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	City        string `json:"city"`
	State       string `json:"state"`
	Country     string `json:"country"`
	Private     bool   `json:"private"`
	MemberCount int    `json:"member_count"`
	URL         string `json:"url"`
	Admin       bool   `json:"admin"`
	Description string `jason:"description"`
}

// ------------------------------------------------------------------------------------------------

type MemberSVTC struct {
	ID        int    `csv:"-"`         // unused currently
	FirstName string `csv:"firstname"` // First name of member
	Middle    string `csv:"middle"`    // Middle initial or name
	LastName  string `csv:"lastname"`  // Last name of member
	Email     string `csv:"email"`     // Member Email
	Status    string `csv:"status"`    // Member status (Expired, Active, Dropped)
	Joined    string `csv:"joined"`    // Date when joined M/D/YY
	Expired   string `csv:"expired"`   // Date when expired / will expire M/D/YY
	Num       int    `csv:"num"`       // Member number
}

// ------------------------------------------------------------------------------------------------

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
