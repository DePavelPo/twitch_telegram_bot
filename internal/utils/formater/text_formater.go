package formater

import (
	"fmt"
	"regexp"
	"strings"
)

func ToLower(text string) string {
	return strings.ToLower(text)
}

// change all twitch tags to hyperlinks with twitch channel's address for telegram
func TwitchTagToTelegram(text string) string {
	var re = regexp.MustCompile(`@[^\s.,!?]+`)
	matches := re.FindAllString(text, -1)
	for _, match := range matches {
		text = strings.ReplaceAll(text, match, fmt.Sprintf("[%s](https://www.twitch.tv/%s)", match, match[1:]))
	}

	return text
}
