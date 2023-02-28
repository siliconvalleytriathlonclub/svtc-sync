package helpers

import (
	"fmt"
	"time"
)

// --------------------------------------------------------------------------------------------

// Function to check a value against a list of options and assign it to a variable if a match is found.
func CheckArgs(argVar *string, argVal string, optionList []string) error {

	for _, s := range optionList {
		if argVal == s {
			*argVar = argVal
			return nil
		}
	}

	return fmt.Errorf("%s", "")

}

// --------------------------------------------------------------------------------------------

// Convert a date string to a golang time.Time object. Returns the object or the zero time (0001-01-01 00:00:00 +0000 UTC) on error.
// The zero time can be checked in the calling function via Time.IsZero()
func GetDate(dstr string) time.Time {

	const layout = "2006-01-02"

	t, err := time.Parse(layout, dstr)
	if err != nil {
		return time.Time{}
	}

	return t
}
