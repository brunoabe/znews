// Package types holds all information about the common structures that are used throughout the API,
// describing the common interfaces for sharing information between the parts of the service.
package types

import (
	"time"
)

// Feed holds information about a feed address.
type Feed struct {
	ID       string
	Provider string
	Category string
	Address  string
}

//Enclosure struct for each Item Enclosure
type Enclosure struct {
	URL  string
	Type string
}

// Article holds the information gathered for each article from the feeds.
type Article struct {
	FeedID      string
	ID          string
	GUID        string
	Title       string
	Link        string
	Comments    string
	PublishDate time.Time
	Categories  []string
	Enclosures  []*Enclosure
	Description string
	Author      string
	Content     string
	FullText    string
}
