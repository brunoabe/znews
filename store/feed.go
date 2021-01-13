package store

import (
	"errors"
	"sync"

	"../types"

	"github.com/google/uuid"
)

// FeedStore stores information about feeds.
type FeedStore struct {
	mu            sync.RWMutex
	m             map[string]*types.Feed
	uuidNamespace uuid.UUID
}

// NewFeedStore returns a new Feed Store.
func NewFeedStore() *FeedStore {
	return &FeedStore{
		m:             map[string]*types.Feed{},
		uuidNamespace: uuid.MustParse(uuidNamespace),
	}
}

// Reset clears the store to its initial state.
func (fs *FeedStore) Reset() {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	fs.m = map[string]*types.Feed{}
}

// Create stores a new feed.
func (fs *FeedStore) Create(feed *types.Feed) (*types.Feed, error) {
	if feed == nil {
		return nil, nil
	}
	generatedID := uuid.NewSHA1(fs.uuidNamespace, []byte(feed.Address)).String()
	if a, ok := fs.m[generatedID]; ok {
		return a, nil
	}
	feed.ID = generatedID
	fs.mu.Lock()
	defer fs.mu.Unlock()
	fs.m[generatedID] = feed
	return feed, nil
}

// List reads feeds from the store and returns all available feeds. The order of the results is not
// guaranteed between calls.
func (fs *FeedStore) List() ([]*types.Feed, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	var res []*types.Feed
	for _, feed := range fs.m {
		res = append(res, feed)
	}

	return res, nil
}

// Get returns a feed from the store based on its GUID if it exists. Returns an error otherwise.
func (fs *FeedStore) Get(ID string) (*types.Feed, error) {
	if ID == "" {
		return nil, errors.New("invalid ID provided")
	}
	fs.mu.RLock()
	defer fs.mu.RUnlock()
	if _, ok := fs.m[ID]; !ok {
		return nil, errors.New("resource not found")
	}
	return fs.m[ID], nil
}
