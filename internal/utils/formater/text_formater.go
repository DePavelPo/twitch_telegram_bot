package formater

import (
	"fmt"
	"regexp"
	"strings"
	"twitch_telegram_bot/internal/models"
)

func ToLower(text string) string {
	return strings.ToLower(text)
}

// change all twitch tags to hyperlinks with twitch channel's address for telegram
func TwitchTagToTelegram(text string) string {
	var re = regexp.MustCompile(`@[^\s.,!?]+`)
	matches := re.FindAllString(text, -1)
	for _, match := range matches {
		text = strings.ReplaceAll(text, match, fmt.Sprintf("[%s](%s/%s)", match, models.TwitchWWWSchemeHost, match[1:]))
	}

	return text
}

// clear all @ symbols in tag subtrings because we can interpret it wrong
func ClearTags(text string) string {
	var re = regexp.MustCompile(`@[^\s.,!?]+`)
	matches := re.FindAllString(text, -1)
	for _, match := range matches {
		text = strings.ReplaceAll(text, match, match[1:])
	}

	return text
}
