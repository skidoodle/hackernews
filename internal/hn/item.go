package hn

import (
	"fmt"
	"net/url"
	"strings"
	"time"
)

type Item struct {
	ID          int     `json:"id"`
	Deleted     bool    `json:"deleted"`
	Type        string  `json:"type"`
	By          string  `json:"by"`
	Time        int64   `json:"time"`
	Text        string  `json:"text"`
	Dead        bool    `json:"dead"`
	Parent      int     `json:"parent"`
	Kids        []int   `json:"kids"`
	URL         string  `json:"url"`
	Score       int     `json:"score"`
	Title       string  `json:"title"`
	Descendants int     `json:"descendants"`
	Comments    []*Item `json:"-"`
}

func (item *Item) Host() string {
	if item.URL == "" {
		return ""
	}
	parsedURL, err := url.Parse(item.URL)
	if err != nil {
		return ""
	}
	return strings.TrimPrefix(parsedURL.Hostname(), "www.")
}

func (item *Item) TimeAgo() string {
	duration := time.Since(time.Unix(item.Time, 0))

	if duration.Hours() >= 24*2 {
		return fmt.Sprintf("%d days ago", int(duration.Hours()/24))
	}
	if duration.Hours() >= 24 {
		return "1 day ago"
	}
	if duration.Hours() >= 2 {
		return fmt.Sprintf("%d hours ago", int(duration.Hours()))
	}
	if duration.Hours() >= 1 {
		return "1 hour ago"
	}
	if duration.Minutes() >= 2 {
		return fmt.Sprintf("%d minutes ago", int(duration.Minutes()))
	}
	if duration.Minutes() >= 1 {
		return "1 minute ago"
	}
	return "just now"
}
