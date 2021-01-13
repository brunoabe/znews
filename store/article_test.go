package store

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"../types"
)

func TestArticleStoreCreate(t *testing.T) {
	t.Run("nil article returns nil and no error", func(t *testing.T) {
		store := NewArticleStore()
		r := require.New(t)
		a := assert.New(t)

		article, err := store.Create(nil)
		r.NoError(err)
		a.Nil(article)
	})

	t.Run("generate the correct UUID ID", func(t *testing.T) {
		store := NewArticleStore()
		r := require.New(t)
		a := assert.New(t)

		article, err := store.Create(&types.Article{
			GUID: "test_guid",
		})
		r.NoError(err)
		a.Equal("dbefb2be-dfe0-5513-b23a-cc04c551221e", article.ID)
	})

	t.Run("all fields are stored and returned correctly", func(t *testing.T) {
		store := NewArticleStore()
		r := require.New(t)
		a := assert.New(t)

		enclosures := []*types.Enclosure{
			&types.Enclosure{
				URL:  "url_1",
				Type: "type_1",
			},
			&types.Enclosure{
				URL:  "url_2",
				Type: "type_2",
			},
		}

		article, err := store.Create(&types.Article{
			FeedID:      "feed_id",
			GUID:        "test_guid",
			Title:       "title",
			Link:        "link",
			Comments:    "comments",
			PublishDate: time.Unix(0, 1).UTC(),
			Categories:  []string{"c1", "c2"},
			Enclosures:  enclosures,
			Description: "description",
			Author:      "author",
			Content:     "content",
			FullText:    "full_text",
		})
		r.NoError(err)
		a.Equal("feed_id", article.FeedID)
		a.Equal("test_guid", article.GUID)
		a.Equal("title", article.Title)
		a.Equal("link", article.Link)
		a.Equal("comments", article.Comments)
		a.Equal(time.Unix(0, 1).UTC(), article.PublishDate)
		a.Equal([]string{"c1", "c2"}, article.Categories)
		a.Equal(enclosures, article.Enclosures)
		a.Equal("description", article.Description)
		a.Equal("author", article.Author)
		a.Equal("content", article.Content)
		a.Equal("full_text", article.FullText)
	})

	t.Run("existent return article and do not duplicate record", func(t *testing.T) {
		store := NewArticleStore()
		r := require.New(t)
		a := assert.New(t)

		_, err := store.Create(&types.Article{
			GUID: "test_guid",
		})
		r.NoError(err)

		_, err = store.Create(&types.Article{
			GUID: "test_guid",
		})
		r.NoError(err)

		articles, err := store.List("", 2, "")
		r.Len(articles, 1, "unexpected number of articles")
		a.Equal("dbefb2be-dfe0-5513-b23a-cc04c551221e", articles[0].ID)
	})

	t.Run("newer article is appended to the end", func(t *testing.T) {
		store := NewArticleStore()
		r := require.New(t)
		a := assert.New(t)

		_, err := store.Create(&types.Article{
			GUID:        "first",
			PublishDate: time.Unix(0, 1).UTC(),
		})
		r.NoError(err)

		_, err = store.Create(&types.Article{
			GUID:        "second",
			PublishDate: time.Unix(0, 2).UTC(),
		})
		r.NoError(err)

		articles, err := store.List("", 3, "")
		r.Len(articles, 2, "unexpected number of articles")
		a.Equal("first", articles[0].GUID)
		a.Equal("second", articles[1].GUID)
	})

	t.Run("older article is appended before", func(t *testing.T) {
		store := NewArticleStore()
		a := assert.New(t)
		r := require.New(t)

		_, err := store.Create(&types.Article{
			GUID:        "second",
			PublishDate: time.Unix(0, 2).UTC(),
		})
		r.NoError(err)

		_, err = store.Create(&types.Article{
			GUID:        "first",
			PublishDate: time.Unix(0, 1).UTC(),
		})
		r.NoError(err)

		articles, err := store.List("", 3, "")
		r.Len(articles, 2, "unexpected number of articles")
		a.Equal("first", articles[0].GUID)
		a.Equal("second", articles[1].GUID)
	})

	t.Run("three articles appended in the correct order", func(t *testing.T) {
		store := NewArticleStore()
		r := require.New(t)
		a := assert.New(t)

		_, err := store.Create(&types.Article{
			GUID:        "first",
			PublishDate: time.Unix(0, 1).UTC(),
		})
		r.NoError(err)

		_, err = store.Create(&types.Article{
			GUID:        "third",
			PublishDate: time.Unix(0, 3).UTC(),
		})
		r.NoError(err)

		_, err = store.Create(&types.Article{
			GUID:        "second",
			PublishDate: time.Unix(0, 2).UTC(),
		})
		r.NoError(err)

		articles, err := store.List("", 4, "")
		r.Len(articles, 3, "unexpected number of articles")
		a.Equal("first", articles[0].GUID)
		a.Equal("second", articles[1].GUID)
		a.Equal("third", articles[2].GUID)
	})

	t.Run("five articles appended in the correct order", func(t *testing.T) {
		store := NewArticleStore()
		r := require.New(t)
		a := assert.New(t)

		_, err := store.Create(&types.Article{
			GUID:        "first",
			PublishDate: time.Unix(0, 1).UTC(),
		})
		r.NoError(err)

		_, err = store.Create(&types.Article{
			GUID:        "third",
			PublishDate: time.Unix(0, 3).UTC(),
		})
		r.NoError(err)

		_, err = store.Create(&types.Article{
			GUID:        "fourth",
			PublishDate: time.Unix(0, 4).UTC(),
		})
		r.NoError(err)

		_, err = store.Create(&types.Article{
			GUID:        "fifth",
			PublishDate: time.Unix(0, 5).UTC(),
		})
		r.NoError(err)

		_, err = store.Create(&types.Article{
			GUID:        "second",
			PublishDate: time.Unix(0, 2).UTC(),
		})
		r.NoError(err)

		articles, err := store.List("", 6, "")
		r.Len(articles, 5, "unexpected number of articles")
		a.Equal("first", articles[0].GUID)
		a.Equal("second", articles[1].GUID)
		a.Equal("third", articles[2].GUID)
		a.Equal("fourth", articles[3].GUID)
		a.Equal("fifth", articles[4].GUID)
	})
}

func TestArticleStoreist(t *testing.T) {
	store := NewArticleStore()
	r := require.New(t)
	_, err := store.Create(&types.Article{
		FeedID:      "feed_id",
		GUID:        "first",
		PublishDate: time.Unix(0, 1).UTC(),
		Categories:  []string{"cat_1", "cat_4"},
	})
	r.NoError(err)

	_, err = store.Create(&types.Article{
		FeedID:      "feed_id",
		GUID:        "second",
		PublishDate: time.Unix(0, 2).UTC(),
		Categories:  []string{"cat_1", "cat_3"},
	})
	r.NoError(err)

	_, err = store.Create(&types.Article{
		FeedID:      "feed_id2",
		GUID:        "third",
		PublishDate: time.Unix(0, 3).UTC(),
		Categories:  []string{"cat_1"},
	})
	r.NoError(err)

	_, err = store.Create(&types.Article{
		FeedID:      "feed_id",
		GUID:        "fourth",
		PublishDate: time.Unix(0, 4).UTC(),
		Categories:  []string{},
	})
	r.NoError(err)

	_, err = store.Create(&types.Article{
		FeedID:      "feed_id2",
		GUID:        "fifth",
		PublishDate: time.Unix(0, 5).UTC(),
		Categories:  []string{"cat_3"},
	})
	r.NoError(err)

	t.Run("empty cursor returns first page", func(t *testing.T) {
		r := require.New(t)
		a := assert.New(t)
		articles, err := store.List("", 2, "")
		r.NoError(err)
		r.Len(articles, 2, "unexpected number of articles")
		a.Equal("first", articles[0].GUID)
		a.Equal("second", articles[1].GUID)
	})

	t.Run("cursor returns next page", func(t *testing.T) {
		r := require.New(t)
		a := assert.New(t)
		// The generated ID of the second item is 461b4f1d-0d71-5a3c-96e8-a2654b90d1ea.
		articles, err := store.List("461b4f1d-0d71-5a3c-96e8-a2654b90d1ea", 2, "")
		r.NoError(err)
		r.Len(articles, 2, "unexpected number of articles")
		a.Equal("third", articles[0].GUID)
		a.Equal("fourth", articles[1].GUID)
	})

	t.Run("last page shows only the remaining items", func(t *testing.T) {
		r := require.New(t)
		a := assert.New(t)
		// The generated ID of the fourth item is 1d852aa8-2ce9-58fa-b8e5-46c3cdd4a098.
		articles, err := store.List("1d852aa8-2ce9-58fa-b8e5-46c3cdd4a098", 2, "")
		r.NoError(err)
		r.Len(articles, 1, "unexpected number of articles")
		a.Equal("fifth", articles[0].GUID)
	})

	t.Run("cursor on last item returns empty result", func(t *testing.T) {
		r := require.New(t)
		a := assert.New(t)
		// The generated ID of the fifth item is a651761e-8285-5539-81de-db51820bda65.
		articles, err := store.List("a651761e-8285-5539-81de-db51820bda65", 2, "")
		r.NoError(err)
		a.Len(articles, 0, "unexpected number of articles")
	})

	t.Run("empty category filter returns everything", func(t *testing.T) {
		r := require.New(t)
		a := assert.New(t)
		articles, err := store.List("", 6, "")
		r.NoError(err)
		a.Len(articles, 5, "unexpected number of articles")
	})

	t.Run("can filter feed", func(t *testing.T) {
		r := require.New(t)
		a := assert.New(t)
		articles, err := store.List("", 6, "feed_id")
		r.NoError(err)
		a.Len(articles, 3, "unexpected number of articles")
		a.Equal("first", articles[0].GUID)
		a.Equal("second", articles[1].GUID)
		a.Equal("fourth", articles[2].GUID)
	})

	t.Run("can filter category for one value", func(t *testing.T) {
		r := require.New(t)
		a := assert.New(t)
		articles, err := store.List("", 6, "", "cat_1")
		r.NoError(err)
		a.Len(articles, 3, "unexpected number of articles")
		a.Equal("first", articles[0].GUID)
		a.Equal("second", articles[1].GUID)
		a.Equal("third", articles[2].GUID)
	})

	t.Run("can filter category for two values one overlaps", func(t *testing.T) {
		r := require.New(t)
		a := assert.New(t)
		articles, err := store.List("", 6, "", "cat_1", "cat_4")
		r.NoError(err)
		a.Len(articles, 3, "unexpected number of articles")
		a.Equal("first", articles[0].GUID)
		a.Equal("second", articles[1].GUID)
		a.Equal("third", articles[2].GUID)
	})

	t.Run("can filter category for two values one semi-overlaps", func(t *testing.T) {
		r := require.New(t)
		a := assert.New(t)
		articles, err := store.List("", 6, "", "cat_1", "cat_3")
		r.NoError(err)
		a.Len(articles, 4, "unexpected number of articles")
		a.Equal("first", articles[0].GUID)
		a.Equal("second", articles[1].GUID)
		a.Equal("third", articles[2].GUID)
		a.Equal("fifth", articles[3].GUID)
	})

	t.Run("can filter category for two values no overlap", func(t *testing.T) {
		r := require.New(t)
		a := assert.New(t)
		articles, err := store.List("", 6, "", "cat_4", "cat_3")
		r.NoError(err)
		a.Len(articles, 3, "unexpected number of articles")
		a.Equal("first", articles[0].GUID)
		a.Equal("second", articles[1].GUID)
		a.Equal("fifth", articles[2].GUID)
	})

	t.Run("return empty if category filter removes all", func(t *testing.T) {
		r := require.New(t)
		a := assert.New(t)
		articles, err := store.List("", 6, "", "cat_invalid")
		r.NoError(err)
		a.Len(articles, 0, "unexpected number of articles")
	})

	t.Run("errors if cursor not found", func(t *testing.T) {
		r := require.New(t)
		a := assert.New(t)
		articles, err := store.List("invalid_cursor", 2, "")
		r.Nil(articles)
		r.Error(err)
		a.Contains(err.Error(), "could not find provided cursor")
	})

}

func TestArticleStoreGet(t *testing.T) {
	store := NewArticleStore()
	r := require.New(t)
	_, err := store.Create(&types.Article{
		GUID:        "first",
		PublishDate: time.Unix(0, 1).UTC(),
		Categories:  []string{"cat_1", "cat_4"},
	})
	r.NoError(err)

	_, err = store.Create(&types.Article{
		GUID:        "second",
		PublishDate: time.Unix(0, 2).UTC(),
		Categories:  []string{"cat_1", "cat_3"},
	})
	r.NoError(err)

	t.Run("errors if ID is empty", func(t *testing.T) {
		r := require.New(t)
		a := assert.New(t)
		article, err := store.Get("")
		r.Nil(article)
		r.Error(err)
		a.Contains(err.Error(), "invalid ID provided")
	})

	t.Run("errors if ID not found", func(t *testing.T) {
		r := require.New(t)
		a := assert.New(t)
		article, err := store.Get("invalid_id")
		r.Nil(article)
		r.Error(err)
		a.Contains(err.Error(), "resource not found")
	})

	t.Run("return correct result", func(t *testing.T) {
		r := require.New(t)
		a := assert.New(t)
		article, err := store.Get("461b4f1d-0d71-5a3c-96e8-a2654b90d1ea")
		r.NoError(err)
		r.NotNil(article)
		a.Equal("second", article.GUID)
	})
}

func TestArticleStoreReset(t *testing.T) {
	t.Run("clears all existing articles", func(t *testing.T) {
		store := NewArticleStore()
		r := require.New(t)
		a := assert.New(t)

		article, err := store.Create(&types.Article{
			FeedID: "feed_id",
			GUID:   "test_guid",
		})
		r.NoError(err)
		a.Equal("dbefb2be-dfe0-5513-b23a-cc04c551221e", article.ID)

		store.Reset()

		articles, err := store.List("", 2, "")
		r.NoError(err)
		a.Len(articles, 0)

		art, err := store.Get("dbefb2be-dfe0-5513-b23a-cc04c551221e")
		r.Error(err)
		a.Contains(err.Error(), "resource not found")
		a.Nil(art)
	})
}
