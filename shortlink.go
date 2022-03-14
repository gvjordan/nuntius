package main

import (
	"encoding/json"
	"net/http"
)

func getShortLink(url string) string {
	// fetch the shortlink and parse the response
	// {
	// "shortID": "cd6f2",
	// "URL": "test.com",
	// "status": "success"
	// }
	resp, err := http.Get("https://llll.ink/new?url=" + url)
	if err != nil {
		return "Error: " + err.Error()
	}

	var shortLink struct {
		ShortID string `json:"shortID"`
		URL     string `json:"URL"`
		Status  string `json:"status"`
	}

	err = json.NewDecoder(resp.Body).Decode(&shortLink)
	if err != nil {
		return "Error: " + err.Error()
	}

	return "https://llll.ink/" + shortLink.ShortID
}
