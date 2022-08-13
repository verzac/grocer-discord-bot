package utils

import "fmt"

// GenericErrorMessage produces an general error message for when 5xx errors occur
func GenericErrorMessage(err error) string {
	return fmt.Sprintf(":helmet_with_cross: Oops, something broke! Give it a day or so and it'll be fixed by the team (or you can follow up this issue with us at our Discord server!). Error:\n```\n%s\n```", err.Error())
}
