package station

import (
	"errors"
	"fmt"
	"regexp"
)

var (
	rx                = regexp.MustCompile("https://(.+.youtube.com|youtu.be)/(watch\\?v=(\\w+)|(\\w+))")
	ErrNotYoutubeLink = errors.New("not a youtube link")
)

// ExtractYoutubeID from message
func ExtractYoutubeID(message string) (string, error) {
	if rx.MatchString(message) {

		sub := rx.FindStringSubmatch(message)
		if len(sub) > 3 {

			// whenever link is youtu.be return sub 4
			if sub[1] == "youtu.be" && len(sub[4]) > 0 {
				return sub[4], nil
			}

			// otherwise, assume link is youtube.com, take the id from ?v=... and check the len

			if len(sub[3]) > 0 {
				return sub[3], nil
			}
		}
	}

	// otherwise assume that not a valid youtube link
	return "", fmt.Errorf("%s %w", message, ErrNotYoutubeLink)
}
