package constants

import "regexp"

var (
	PhotoPath       = "/share/Photo"
	DuplicateMarker = "DUP"
	DuplicateRegExp = regexp.MustCompile(DuplicateMarker + "_([0-9]+)_")
)
