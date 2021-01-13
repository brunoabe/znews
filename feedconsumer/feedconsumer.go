// Package feedconsumer provides functionality to fetch articles from a Feed and save them in an
// ArticleStore.
package feedconsumer

import (
	"fmt"

	"../types"
)

// Feed describes the functionality required to load data from a feed.
type Feed interface {
	Load(address string) ([]*types.Article, error)
}

// ArticleStore describes the functionality needed to store articles.
type ArticleStore interface {
	Create(article *types.Article) (*types.Article, error)
}

// FeedConsumer is a consumer that fetches articles from a feed and stores them in a store.
type FeedConsumer struct {
	feed  Feed
	store ArticleStore
}

// Consume fetches news from the provided feed and saves them in the provided store.
func (c *FeedConsumer) Consume(feed *types.Feed) error {
	articles, err := c.feed.Load(feed.Address)
	if err != nil {
		return fmt.Errorf("could not load articles from the feed: %v", err)
	}
	if len(articles) == 0 {
		return nil
	}
	for _, article := range articles {
		article.FeedID = feed.ID
		_, err := c.store.Create(article)
		if err != nil {
			return err
		}
	}
	return nil
}

// NewFeedConsumer returns a new FeedConsumer providing functionality to gather news/articles from
// the provided feed and saving them in the provided store.
func NewFeedConsumer(feed Feed, store ArticleStore) *FeedConsumer {
	return &FeedConsumer{
		feed:  feed,
		store: store,
	}
}
