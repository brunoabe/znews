package converters

import (
	"testing"
	// "time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/ungerik/go-rss"
	// "../../types"
)

func TestRSSToNativeArticles(t *testing.T) {
	t.Run("empty slice returns empty results", func(t *testing.T) {
		r := require.New(t)
		a := assert.New(t)
		articles, err := RSSToNativeArticles([]rss.Item{})
		r.NoError(err)
		a.Nil(articles)
	})

	t.Run("convert enclosure correctly", func(t *testing.T) {
		r := require.New(t)
		a := assert.New(t)
		articles, err := RSSToNativeArticles([]rss.Item{
			rss.Item{
				PubDate: "Tue, 12 Jan 2021 00:05:18 GMT",
				Enclosure: []rss.ItemEnclosure{
					rss.ItemEnclosure{
						URL:  "url",
						Type: "type",
					},
				},
			},
		})
		r.NoError(err)
		r.Len(articles, 1, "unexpected number of articles")
		r.Len(articles[0].Enclosures, 1, "unexpected number of enclosures")
		a.Equal("url", articles[0].Enclosures[0].URL)
		a.Equal("type", articles[0].Enclosures[0].Type)
	})

	t.Run("errors for invalid publish date", func(t *testing.T) {
		r := require.New(t)
		a := assert.New(t)
		articles, err := RSSToNativeArticles([]rss.Item{rss.Item{}})

		r.Empty(articles)
		r.Error(err)
		a.Contains(err.Error(), "could not parse publish date")
	})

	t.Run("convert enclosures correctly", func(t *testing.T) {
		r := require.New(t)
		a := assert.New(t)
		articles, err := RSSToNativeArticles([]rss.Item{
			rss.Item{
				PubDate: "Tue, 12 Jan 2021 00:05:18 GMT",
				Enclosure: []rss.ItemEnclosure{
					rss.ItemEnclosure{
						URL:  "url",
						Type: "type",
					},
					rss.ItemEnclosure{
						URL:  "url2",
						Type: "type2",
					},
				},
			},
		})
		r.NoError(err)
		r.Len(articles, 1, "unexpected number of articles")
		r.Len(articles[0].Enclosures, 2, "unexpected number of enclosures")
		a.Equal("url", articles[0].Enclosures[0].URL)
		a.Equal("type", articles[0].Enclosures[0].Type)
		a.Equal("url2", articles[0].Enclosures[1].URL)
		a.Equal("type2", articles[0].Enclosures[1].Type)
	})

	t.Run("convert all fields correctly", func(t *testing.T) {
		r := require.New(t)
		a := assert.New(t)
		categories := []string{"cat_1", "cat_2"}
		articles, err := RSSToNativeArticles([]rss.Item{
			rss.Item{
				GUID:        "guid",
				Title:       "title",
				Link:        "link",
				Comments:    "comments",
				PubDate:     "Mon, 02 Jan 2006 15:04:05 MST",
				Category:    categories,
				Enclosure:   []rss.ItemEnclosure{},
				Description: "description",
				Author:      "author",
				Content:     "content",
				FullText:    "full_text",
			},
		})
		r.NoError(err)
		r.Len(articles, 1, "unexpected number of articles")
		r.Len(articles[0].Enclosures, 0, "unexpected number of enclosures")
		a.Equal("guid", articles[0].GUID)
		a.Equal("title", articles[0].Title)
		a.Equal("link", articles[0].Link)
		a.Equal("comments", articles[0].Comments)
		a.EqualValues(1136214245, articles[0].PublishDate.UTC().Unix())
		a.Equal(categories, articles[0].Categories)
		a.Equal("description", articles[0].Description)
		a.Equal("author", articles[0].Author)
		a.Equal("content", articles[0].Content)
		a.Equal("full_text", articles[0].FullText)
	})
}
