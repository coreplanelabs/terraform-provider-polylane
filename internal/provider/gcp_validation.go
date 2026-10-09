package provider

import "regexp"

var (
	gcpProjectIDPattern     = regexp.MustCompile(`^[a-z][a-z0-9-]{4,28}[a-z0-9]$`)
	gcpProjectNumberPattern = regexp.MustCompile(`^[1-9][0-9]{0,19}$`)
)
