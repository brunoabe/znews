package store

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"../types"
)

func TestFeedStoreCreate(t *testing.T) {
	t.Run("nil feed returns nil and no error", func(t *testing.T) {
		store := NewFeedStore()
		r := require.New(t)
		a := assert.New(t)

		feed, err := store.Create(nil)
		r.NoError(err)
		a.Nil(feed)
	})

	t.Run("generate the correct UUID ID", func(t *testing.T) {
		store := NewFeedStore()
		r := require.New(t)
		a := assert.New(t)

		feed, err := store.Create(&types.Feed{
			Address: "test_guid",
		})
		r.NoError(err)
		a.Equal("dbefb2be-dfe0-5513-b23a-cc04c551221e", feed.ID)
	})

	t.Run("all fields are stored correctly", func(t *testing.T) {
		store := NewFeedStore()
		r := require.New(t)
		a := assert.New(t)

		feed, err := store.Create(&types.Feed{
			Provider: "provider",
			Category: "category",
			Address:  "test_guid",
		})
		r.NoError(err)
		a.Equal("provider", feed.Provider)
		a.Equal("category", feed.Category)
		a.Equal("test_guid", feed.Address)
	})

	t.Run("existent return article and do not duplicate record", func(t *testing.T) {
		store := NewFeedStore()
		r := require.New(t)
		a := assert.New(t)

		_, err := store.Create(&types.Feed{
			Address: "test_guid",
		})
		r.NoError(err)

		_, err = store.Create(&types.Feed{
			Address: "test_guid",
		})
		r.NoError(err)

		feeds, err := store.List()
		r.Len(feeds, 1, "unexpected number of feeds")
		a.Equal("dbefb2be-dfe0-5513-b23a-cc04c551221e", feeds[0].ID)
	})
}

func TestFeedStoreList(t *testing.T) {
	store := NewFeedStore()
	r := require.New(t)

	_, err := store.Create(&types.Feed{
		Address: "test_guid",
	})
	r.NoError(err)
	_, err = store.Create(&types.Feed{
		Address: "test_guid_2",
	})
	r.NoError(err)

	t.Run("list all available values", func(t *testing.T) {
		r := require.New(t)
		a := assert.New(t)
		feeds, err := store.List()
		r.NoError(err)
		r.Len(feeds, 2, "unexpected number of feeds")
		addresses := []string{feeds[0].Address, feeds[1].Address}
		a.ElementsMatch([]string{"test_guid", "test_guid_2"}, addresses)
	})
}

func TestFeedStoreGet(t *testing.T) {
	store := NewFeedStore()
	r := require.New(t)

	_, err := store.Create(&types.Feed{
		Address: "test_guid",
	})
	r.NoError(err)
	_, err = store.Create(&types.Feed{
		Address: "test_guid_2",
	})
	r.NoError(err)

	t.Run("errors if ID is empty", func(t *testing.T) {
		r := require.New(t)
		a := assert.New(t)
		feed, err := store.Get("")
		r.Nil(feed)
		r.Error(err)
		a.Contains(err.Error(), "invalid ID provided")
	})

	t.Run("errors if ID not found", func(t *testing.T) {
		r := require.New(t)
		a := assert.New(t)
		feed, err := store.Get("invalid_id")
		r.Nil(feed)
		r.Error(err)
		a.Contains(err.Error(), "resource not found")
	})

	t.Run("return correct result", func(t *testing.T) {
		r := require.New(t)
		a := assert.New(t)
		feed, err := store.Get("dbefb2be-dfe0-5513-b23a-cc04c551221e")
		r.NoError(err)
		r.NotNil(feed)
		a.Equal("test_guid", feed.Address)
	})
}

func TestFeedStoreReset(t *testing.T) {
	t.Run("clears all existing feeds", func(t *testing.T) {
		store := NewFeedStore()
		r := require.New(t)
		a := assert.New(t)

		feed, err := store.Create(&types.Feed{
			Address: "test_guid",
		})
		r.NoError(err)
		a.Equal("dbefb2be-dfe0-5513-b23a-cc04c551221e", feed.ID)

		store.Reset()

		feeds, err := store.List()
		r.NoError(err)
		a.Len(feeds, 0, "unexpected number of feeds")
	})
}
