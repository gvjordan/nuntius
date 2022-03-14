package main

import (
	"fmt"
	"regexp"
	"strings"
)

func reverseChannelMap(m map[string]interface{}) map[string]interface{} {
	r := make(map[string]interface{})
	for k, v := range m {
		r[v.(string)] = k
	}
	return r
}

func channelIsInMap(channel string) bool {
	// check if channel exists in mapToDiscord or mapToIRC
	if _, ok := mapToDiscord[channel]; ok {
		return true
	} else if _, ok := mapToIRC[channel]; ok {
		return true
	}
	return false
}

func formatter(data interface{}) string {
	user := data.(*msgObject).User
	message := data.(*msgObject).Message
	target := data.(*msgObject).Target

	formattedString := ""
	if target == "irc" {
		formattedString = "<" + colorWrapUser(user) + "> " + message
	}

	if target == "discord" {
		formattedString = "**<" + user + ">** " + stripColorCodes(message)
	}
	return formattedString
}

func getColorIndexByUsername(username string) int {
	firstChar := []rune(username)[0]
	colorIndex := (int(firstChar) + len(username)) % 19
	return colorIndex
}

func colorWrapUser(username string) string {
	colorIndex := getColorIndexByUsername(username)
	return fmt.Sprintf("\x03%d%s\x03", colorIndex, username)
}

func stripColorCodes(s string) string {
	CharBold := "\x02"

	colorRegex := regexp.MustCompile(`\x03(\d\d?)?(?:,(\d\d?))?`)
	replacer := strings.NewReplacer(
		string(CharBold), "",
	)
	newString := colorRegex.ReplaceAllString(s, "")
	newString = replacer.Replace(newString)
	return newString
}
