// Package e2e holds all end to end test cases that guarantee that the API is working as expected
// with all happy-path scenarios being executed. Running this test package spins up a local web
// service and allow test cases to make calls to it, validating the required features are working.
package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"../feedconsumer"
	"../rssreader"
	"../service"
	"../store"
	"../types"
)

const (
	testPort               = 8787
	testRssFeed            = "https://www.nytimes.com/svc/collections/v1/publish/https://www.nytimes.com/section/world/rss.xml"
	testRssFeedID          = "7ebf8f47-3aef-509e-a3fe-1bfe55e317a8"
	testSecondaryRssFeed   = "http://feeds.bbci.co.uk/news/uk/rss.xml"
	testSecondaryRssFeedID = "0792cd43-d8f3-5a38-9739-c797bd08c6fa"
)

// ResponseError reads the error message returned in a JSON response.
type ResponseError struct {
	Error string
}

// TestSuite represents a test suite capable of running all e2e test cases for the service.
type TestSuite struct {
	httpServerExitDone *sync.WaitGroup
	suite.Suite
	feedStore    *store.FeedStore
	articleStore *store.ArticleStore
	feed         *rssreader.Feed
	service      *service.Service
}

// SetupTest runs whenever a new test starts. In this case, it resets the stores to allow the test
// initial state to be the same for all test cases.
func (s *TestSuite) SetupTest() {
	s.feedStore.Reset()
	s.articleStore.Reset()
}

// SetupSuite runs whenever the whole test suite starts. In this case it loads new components and
// inject them into a local web service that exposes the API functionality so that the tests can
// call the API endpoints through HTTP requests.
func (s *TestSuite) SetupSuite() {
	s.feedStore = store.NewFeedStore()
	s.articleStore = store.NewArticleStore()
	s.feed = rssreader.NewFeed()
	consumer := feedconsumer.NewFeedConsumer(s.feed, s.articleStore)
	s.service = service.NewService(consumer, s.feedStore, s.articleStore)
	go s.service.ServeForever(testPort)
}

// TestCanReadArticlesFromANewsFeed tests that calling the correct API endpoints will store a new
// feed and be able to request data from it. It has a dependency on the rss provider actually having
// sensible results. A static page could be published to make the test results more predictable.
func (s *TestSuite) TestCanReadArticlesFromANewsFeed() {
	t := s.T()
	a := assert.New(t)
	r := require.New(t)

	// Create a new feed.
	feed := s.createFeed("p", "c", testRssFeed)
	a.Equal("c", feed.Category)
	a.Equal("p", feed.Provider)
	a.Equal(testRssFeed, feed.Address)
	r.Equal(testRssFeedID, feed.ID)

	// Load the feed.
	s.loadFeed(feed.ID)

	// Check if results were loaded.
	articles := s.listArticles("", 1, "")
	a.Len(articles, 1)
}

// TestCanReadRequiredNewsFeed tests that all provided rss feeds as examples are able to be loaded
// by the API.
func (s *TestSuite) TestCanReadRequiredNewsFeed() {
	t := s.T()
	a := assert.New(t)
	r := require.New(t)

	exampleFeeds := map[string]string{
		"0792cd43-d8f3-5a38-9739-c797bd08c6fa": "http://feeds.bbci.co.uk/news/uk/rss.xml",
		"eb486396-2226-5371-a2b3-15f9dcfa6235": "http://feeds.bbci.co.uk/news/technology/rss.xml",
		"d9641b79-1061-590e-b904-9ed34a852b47": "http://feeds.skynews.com/feeds/rss/uk.xml",
		"101371f8-b500-5a45-8eef-14c6101ab650": "http://feeds.skynews.com/feeds/rss/technology.xml",
	}

	// Create new feeds and load articles.
	for id, feedAddress := range exampleFeeds {
		feed := s.createFeed(id+"_p", id+"_c", feedAddress)
		a.Equal(id+"_c", feed.Category)
		a.Equal(id+"_p", feed.Provider)
		a.Equal(feedAddress, feed.Address)
		r.Equal(id, feed.ID)
		s.loadFeed(feed.ID)
	}

	// Getting the first two articles to check pagination.
	articles := s.listArticles("", 100, "")
	feedCountMap := map[string]int{}

	for _, a := range articles {
		if _, ok := feedCountMap[a.FeedID]; !ok {
			feedCountMap[a.FeedID] = 1
			continue
		}
		feedCountMap[a.FeedID]++
	}

	r.Equal(4, len(feedCountMap), "unexpected number of feeds")

	for c, count := range feedCountMap {
		// Checking all found categories for the length of articles.
		firstPageArticles := s.listArticles("", 100, c)
		r.Len(firstPageArticles, count)
	}
}

// TestCanLazyLoadArticles tests that the API can provide a good manner of scrolling throughout the
// articles without having to make a huge request, being able to load as required.
func (s *TestSuite) TestCanLazyLoadArticles() {
	t := s.T()
	a := assert.New(t)
	r := require.New(t)

	// Create a new feed.
	feed := s.createFeed("p", "c", testRssFeed)
	a.Equal("c", feed.Category)
	a.Equal("p", feed.Provider)
	a.Equal(testRssFeed, feed.Address)
	r.Equal(testRssFeedID, feed.ID)

	// Load the feed.
	s.loadFeed(feed.ID)

	// Getting the first two articles to check pagination.
	articles := s.listArticles("", 2, "")
	a.Len(articles, 2)
	firstID := articles[0].ID
	secondID := articles[1].ID

	// Getting the first article paginated.
	firstPageArticles := s.listArticles("", 1, "")
	a.Len(firstPageArticles, 1)
	a.Equal(firstPageArticles[0].ID, firstID)

	// Getting the second article and checks pagination.
	secondPageArticles := s.listArticles(firstID, 1, "")
	a.Len(secondPageArticles, 1)
	a.Equal(secondPageArticles[0].ID, secondID)
}

// TestCanFilterArticleCategories tests that the API can return articles filtered by categories.
func (s *TestSuite) TestCanFilterArticlesByCategories() {
	t := s.T()
	a := assert.New(t)
	r := require.New(t)

	// Create a new feed.
	feed := s.createFeed("p", "c", testRssFeed)
	a.Equal("c", feed.Category)
	a.Equal("p", feed.Provider)
	a.Equal(testRssFeed, feed.Address)
	r.Equal(testRssFeedID, feed.ID)

	// Load the feed.
	s.loadFeed(feed.ID)

	// Getting all articles to check how many articles are found per category.
	articles := s.listArticles("", 0, "")
	categoriesCountMap := map[string]int{}

	// Stores the count of articles per category to assert filtering.
	for _, a := range articles {
		if len(a.Categories) > 0 {
			for _, c := range a.Categories {
				if _, ok := categoriesCountMap[c]; !ok {
					categoriesCountMap[c] = 1
					continue
				}
				categoriesCountMap[c]++
			}
		}
	}

	// If the provided feed returns only articles with no categories, this test is not valid.
	r.NotEqual(0, len(categoriesCountMap), "invalid test, no articles have categories")

	for c, count := range categoriesCountMap {
		// Checking all found categories for the length of articles.
		firstPageArticles := s.listArticles("", 0, "", c)
		r.Len(firstPageArticles, count)
	}
}

// TestCanFilterArticlesByFeed tests that the API can return articles filtered by feed, fulfilling
// the requirement of selecting a feed based on a provider or a category.
func (s *TestSuite) TestCanFilterArticlesByFeed() {
	t := s.T()
	a := assert.New(t)
	r := require.New(t)

	// Create a new feed.
	feed := s.createFeed("p1", "c1", testRssFeed)
	a.Equal("c1", feed.Category)
	a.Equal("p1", feed.Provider)
	a.Equal(testRssFeed, feed.Address)
	r.Equal(testRssFeedID, feed.ID)

	// Create a secondary feed to have two stored feeds and articles in the store.
	feed2 := s.createFeed("p2", "c2", testSecondaryRssFeed)
	a.Equal("c2", feed2.Category)
	a.Equal("p2", feed2.Provider)
	a.Equal(testSecondaryRssFeed, feed2.Address)
	r.Equal(testSecondaryRssFeedID, feed2.ID)

	// Load the feeds.
	s.loadFeed(feed.ID)
	s.loadFeed(feed2.ID)

	// Getting all articles to check how many articles are found per category.
	articles := s.listArticles("", 0, "")
	feedCountMap := map[string]int{}

	// Stores the count of articles per feed to assert filtering.
	for _, a := range articles {
		if _, ok := feedCountMap[a.FeedID]; !ok {
			feedCountMap[a.FeedID] = 1
			continue
		}
		feedCountMap[a.FeedID]++
	}

	r.Equal(2, len(feedCountMap), "unexpected number of feeds")

	for c, count := range feedCountMap {
		// Checking all found feeds for the length of articles.
		firstPageArticles := s.listArticles("", 0, c)
		r.Len(firstPageArticles, count)
	}
}

// TestCanGetSingleArticleByID tests that the API can return a single article based on its ID, which
// would allow an article to be seen on screen in full. It also provides enough information to be
// shared via social network or e-mail.
func (s *TestSuite) TestCanGetSingleArticleByID() {
	t := s.T()
	a := assert.New(t)
	r := require.New(t)

	// Create a new feed.
	feed := s.createFeed("p", "c", testRssFeed)
	a.Equal("c", feed.Category)
	a.Equal("p", feed.Provider)
	a.Equal(testRssFeed, feed.Address)
	r.Equal(testRssFeedID, feed.ID)

	// Load the feed.
	s.loadFeed(feed.ID)

	// Getting the first four articles and checking the ID of the latest one returned.
	articles := s.listArticles("", 4, "")
	r.NotEqual(0, len(articles), "invalid test, no articles returned")
	selectedArticle := articles[len(articles)-1]
	articleID := selectedArticle.ID

	// Get the article by the selected ID and compare the information.
	article := s.getArticle(articleID)
	r.NotNil(article)
	a.Equal(articleID, article.ID)
}

// TestArticlesAreRetunedOrderedByPublishDate tests that when the API is called to list articles,
// they are returned ordered by publish date.
func (s *TestSuite) TestArticlesAreRetunedOrderedByPublishDate() {
	t := s.T()
	a := assert.New(t)
	r := require.New(t)

	// Create a new feed.
	feed := s.createFeed("p", "c", testRssFeed)
	a.Equal("c", feed.Category)
	a.Equal("p", feed.Provider)
	a.Equal(testRssFeed, feed.Address)
	r.Equal(testRssFeedID, feed.ID)

	// Load the feed.
	s.loadFeed(feed.ID)

	// Check if results were loaded.
	articles := s.listArticles("", 0, "")
	a.True(len(articles) > 1, "not enough articles to check the condition")

	// Checking order.
	for i := 0; i < len(articles)-1; i++ {
		r.True(!articles[i].PublishDate.After(articles[i+1].PublishDate), "found unordered article")
	}
}

func (s *TestSuite) createFeed(provider string, category string, address string) *types.Feed {
	r := require.New(s.T())
	feedData := map[string]string{
		"provider": provider,
		"category": category,
		"address":  address,
	}
	jsonData, _ := json.Marshal(feedData)

	req, err := http.NewRequest(http.MethodPut, getAPIUrl("feeds"), bytes.NewReader(jsonData))
	r.NoError(err)
	client := &http.Client{}
	res, err := client.Do(req)
	r.NoError(err)
	defer res.Body.Close()
	var feed types.Feed
	err = json.NewDecoder(res.Body).Decode(&feed)
	r.NoError(err)
	return &feed
}

func (s *TestSuite) loadFeed(ID string) {
	r := require.New(s.T())
	feedData := map[string]string{
		"id": ID,
	}
	jsonData, _ := json.Marshal(feedData)

	req, err := http.NewRequest(http.MethodPost, getAPIUrl("feeds/load"), bytes.NewReader(jsonData))
	r.NoError(err)
	client := &http.Client{}
	res, err := client.Do(req)
	r.NoError(err)
	defer res.Body.Close()
	var responseError ResponseError
	_ = json.NewDecoder(res.Body).Decode(&responseError)
	r.Empty(responseError.Error)
}

func (s *TestSuite) listArticles(cursor string, pageSize uint, feed string, categories ...string) []*types.Article {
	r := require.New(s.T())
	req, err := http.NewRequest(http.MethodGet, getAPIUrl("articles"), nil)
	r.NoError(err)
	q := req.URL.Query()
	q.Add("c", cursor)
	q.Add("pageSize", fmt.Sprintf("%d", pageSize))
	q.Add("feed", feed)
	for _, c := range categories {
		q.Add("cat", c)
	}
	req.URL.RawQuery = q.Encode()
	client := &http.Client{}
	res, err := client.Do(req)
	r.NoError(err)
	defer res.Body.Close()
	var articles []*types.Article
	err = json.NewDecoder(res.Body).Decode(&articles)
	r.NoError(err)
	return articles
}

func (s *TestSuite) getArticle(ID string) *types.Article {
	r := require.New(s.T())
	req, err := http.NewRequest(http.MethodGet, getAPIUrl("articles", ID), nil)
	r.NoError(err)
	client := &http.Client{}
	res, err := client.Do(req)
	r.NoError(err)
	defer res.Body.Close()
	var article types.Article
	err = json.NewDecoder(res.Body).Decode(&article)
	r.NoError(err)
	return &article
}

func getAPIUrl(action string, args ...string) string {
	r := fmt.Sprintf("http://localhost:%d/%s", testPort, action)
	for _, a := range args {
		r = fmt.Sprintf("%s/%s", r, a)
	}
	return r
}

// TestTestSuite runs the whole end to end test suite.
func TestTestSuit(t *testing.T) {
	suite.Run(t, new(TestSuite))
}
