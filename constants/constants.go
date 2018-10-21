package constants

import (
	"regexp"
	"strings"
)

var (
	PhotoPath       = "/share/Photo"
	DuplicateMarker = "DUP"
	DuplicateRegExp = regexp.MustCompile(DuplicateMarker + "_([0-9]+)_")
	PhotoExtensions = map[string]bool{
		"jpg":  true,
		"jpeg": true,
		"png":  true,
		"gif":  true,
	}
	FilteredTerms = []string{"thumb", "Recycle", "Thumb"}
)

func FilterPAth(path string) bool {
	for _, filter := range FilteredTerms {
		if strings.Contains(path, filter) {
			return false
		}
	}
	return true
}
