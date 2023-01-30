package main

import (
	"fmt"
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
