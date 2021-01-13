// Package rssreader provides functionality for reading an rss feed and returning the articles found
// in a standard format.
package rssreader

import (
	"../types"
	"./converters"

	"github.com/ungerik/go-rss"
)

// Feed provides the functionality required for consuming articles from RSS feeds.
type Feed struct{}

// NewFeed returns a new feed for the provided RSS feed address.
func NewFeed() *Feed {
	return &Feed{}
}

// Load reads the feed configured on instantiation and returns a slice of articles.
func (rssf *Feed) Load(address string) ([]*types.Article, error) {
	res, err := rss.Read(address, false)
	if err != nil {
		return nil, err
	}

	channel, err := rss.Regular(res)
	if err != nil {
		return nil, err
	}

	articles, err := converters.RSSToNativeArticles(channel.Item)
	if err != nil {
		return nil, err
	}

	return articles, nil
}
