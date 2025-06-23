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

	years := int(duration.Hours() / (24 * 365))
	if years > 1 {
		return fmt.Sprintf("%d years ago", years)
	}
	if years == 1 {
		return "1 year ago"
	}

	months := int(duration.Hours() / (24 * 30))
	if months > 1 {
		return fmt.Sprintf("%d months ago", months)
	}
	if months == 1 {
		return "1 month ago"
	}

	days := int(duration.Hours() / 24)
	if days > 1 {
		return fmt.Sprintf("%d days ago", days)
	}
	if days == 1 {
		return "1 day ago"
	}

	hours := int(duration.Hours())
	if hours > 1 {
		return fmt.Sprintf("%d hours ago", hours)
	}
	if hours == 1 {
		return "1 hour ago"
	}

	minutes := int(duration.Minutes())
	if minutes > 1 {
		return fmt.Sprintf("%d minutes ago", minutes)
	}
	if minutes == 1 {
		return "1 minute ago"
	}

	return "just now"
}
