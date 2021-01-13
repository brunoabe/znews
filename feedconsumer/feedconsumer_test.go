package feedconsumer

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"../types"
)

type MockFeed struct {
	mock.Mock
}

func (mf *MockFeed) Load(address string) ([]*types.Article, error) {
	args := mf.Called(address)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*types.Article), args.Error(1)
}

type MockArticleStore struct {
	mock.Mock
}

func (mas *MockArticleStore) Create(article *types.Article) (*types.Article, error) {
	args := mas.Called(article)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.Article), args.Error(1)
}

func TestConsume(t *testing.T) {
	t.Run("bypasses feed loading error", func(t *testing.T) {
		r := require.New(t)
		a := assert.New(t)
		mockFeed := &MockFeed{}
		mockFeed.On("Load", "address").Return(nil, errors.New("random error"))
		feedConsumer := NewFeedConsumer(mockFeed, nil)
		err := feedConsumer.Consume(&types.Feed{Address: "address"})
		r.Error(err)
		a.Contains(err.Error(), "random error")
		mockFeed.AssertExpectations(t)
	})

	t.Run("return nil if no articles are fetched", func(t *testing.T) {
		r := require.New(t)
		mockFeed := &MockFeed{}
		mockFeed.On("Load", "address").Return(nil, nil)
		mockArticleStore := &MockArticleStore{}
		feedConsumer := NewFeedConsumer(mockFeed, mockArticleStore)
		err := feedConsumer.Consume(&types.Feed{Address: "address"})
		r.Nil(err)
		mockFeed.AssertExpectations(t)
	})

	t.Run("bypass store inserting error", func(t *testing.T) {
		r := require.New(t)
		a := assert.New(t)
		mockFeed := &MockFeed{}
		articlesToReturn := []*types.Article{
			&types.Article{},
		}
		mockFeed.On("Load", "address").Return(articlesToReturn, nil)
		mockArticleStore := &MockArticleStore{}
		mockArticleStore.On("Create", articlesToReturn[0]).Return(nil, errors.New("random error"))
		feedConsumer := NewFeedConsumer(mockFeed, mockArticleStore)
		err := feedConsumer.Consume(&types.Feed{Address: "address"})
		r.Error(err)
		a.Contains(err.Error(), "random error")
		mockFeed.AssertExpectations(t)
	})

	t.Run("consume and stores returned article", func(t *testing.T) {
		r := require.New(t)
		mockFeed := &MockFeed{}
		articlesToReturn := []*types.Article{
			&types.Article{GUID: "test_guid"},
		}
		savedArticle := &types.Article{
			ID:   "generated_uuid",
			GUID: "test_guid",
		}
		mockFeed.On("Load", "address").Return(articlesToReturn, nil)
		mockArticleStore := &MockArticleStore{}
		mockArticleStore.On("Create", articlesToReturn[0]).Return(savedArticle, nil)
		feedConsumer := NewFeedConsumer(mockFeed, mockArticleStore)
		err := feedConsumer.Consume(&types.Feed{Address: "address"})
		r.NoError(err)
		mockFeed.AssertExpectations(t)
		mockArticleStore.AssertExpectations(t)
	})

	t.Run("consume and stores returned articles", func(t *testing.T) {
		r := require.New(t)
		mockFeed := &MockFeed{}
		articlesToReturn := []*types.Article{
			&types.Article{GUID: "test_guid"},
			&types.Article{GUID: "test_guid_2"},
		}
		savedArticle := &types.Article{
			ID:   "generated_uuid",
			GUID: "test_guid",
		}
		mockFeed.On("Load", "address").Return(articlesToReturn, nil)
		mockArticleStore := &MockArticleStore{}
		mockArticleStore.On("Create", articlesToReturn[0]).Return(savedArticle, nil)
		mockArticleStore.On("Create", articlesToReturn[1]).Return(savedArticle, nil)
		feedConsumer := NewFeedConsumer(mockFeed, mockArticleStore)
		err := feedConsumer.Consume(&types.Feed{Address: "address"})
		r.NoError(err)
		mockFeed.AssertExpectations(t)
		mockArticleStore.AssertExpectations(t)
	})

}
