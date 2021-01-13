// Package converters provides functionality to convert articles from the rss library into the
// internal representation of articles.
package converters

import (
	"fmt"

	"../../types"

	"github.com/ungerik/go-rss"
)

const dateFormat = "Mon, 02 Jan 2006 15:04:05 MST"

// RSSToNativeArticles converts a slice of items provided by the rss library into the internal
// representation of an article.
func RSSToNativeArticles(is []rss.Item) ([]*types.Article, error) {
	if len(is) == 0 {
		return nil, nil
	}
	articles := make([]*types.Article, 0, len(is))
	for _, i := range is {
		a, err := rssToNativeArticle(i)
		if err != nil {
			return nil, err // The error returned here will have some format already.
		}
		articles = append(articles, a)
	}
	return articles, nil
}

func rssToNativeArticle(i rss.Item) (*types.Article, error) {
	publishDate, err := i.PubDate.ParseWithFormat(dateFormat)
	if err != nil {
		return nil, fmt.Errorf("could not parse publish date: %v", err)
	}
	return &types.Article{
		GUID:        i.GUID,
		Title:       i.Title,
		Link:        i.Link,
		Comments:    i.Comments,
		PublishDate: publishDate,
		Categories:  i.Category,
		Enclosures:  rssToNativeEnclosures(i.Enclosure),
		Description: i.Description,
		Author:      i.Author,
		Content:     i.Content,
		FullText:    i.FullText,
	}, nil
}

func rssToNativeEnclosure(ie rss.ItemEnclosure) *types.Enclosure {
	return &types.Enclosure{
		URL:  ie.URL,
		Type: ie.Type,
	}
}

func rssToNativeEnclosures(ies []rss.ItemEnclosure) []*types.Enclosure {
	enclosures := make([]*types.Enclosure, 0, len(ies))
	for _, ie := range ies {
		enclosures = append(enclosures, rssToNativeEnclosure(ie))
	}
	return enclosures
}
