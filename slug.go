package chronicle

import (
	"regexp"
	"strings"
)

//Slugify return a human readable slug of a text
func Slugify(text string) string {
	wordsLowerCased := []string{}
	for _, word := range strings.Split(text, " ") {
		regex, _ := regexp.Compile("[a-zA-Z0-9]+")
		matchedString := regex.FindString(word)

		if len(matchedString) > 0 {
			wordsLowerCased = append(wordsLowerCased, strings.ToLower(matchedString))
		}
	}

	return strings.Join(wordsLowerCased, "-")
}
