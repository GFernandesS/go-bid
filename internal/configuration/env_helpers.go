package configuration

import (
	"os"
	"strconv"
)

func ShouldUseCSRFToken() bool {
	value, err := strconv.ParseBool(os.Getenv("GOBID_USE_CSRF_TOKEN"))

	if err != nil {
		return false
	}

	return value
}
