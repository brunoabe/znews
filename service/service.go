// Package service exposes functionality for the API through RESTful endpoints.
package service

import (
	"fmt"
	"log"
	"net/http"

	"../types"

	"github.com/gin-gonic/gin"
)

// Feeder describes the functionality needed to consume articles from a feeder.
type Feeder interface {
	Consume(feed *types.Feed) error
}

// ArticleStore describes the functionality needed to store and retrieve articles.
type ArticleStore interface {
	List(cursor string, pageSize int, feed string, categories ...string) ([]*types.Article, error)
	Get(ID string) (*types.Article, error)
}

// FeedStore describes the functionality needed to store and retrieve feeds.
type FeedStore interface {
	List() ([]*types.Feed, error)
	Create(feed *types.Feed) (*types.Feed, error)
	Get(ID string) (*types.Feed, error)
}

// Service represents a web service capable of acting on RESTful requests for getting articles.
type Service struct {
	feeder       Feeder
	articleStore ArticleStore
	feedStore    FeedStore
}

// NewService returns a new Service capable of exposing the required endpoints for the news app.
func NewService(feeder Feeder, feedStore FeedStore, articleStore ArticleStore) *Service {
	return &Service{
		feeder:       feeder,
		feedStore:    feedStore,
		articleStore: articleStore,
	}
}

// ServeForever sets up the service router and start serving until receiving a signal to exit.
func (s *Service) ServeForever(port uint) {
	r := s.setupServiceRouter()
	// Run http server
	if err := r.Run(fmt.Sprintf(":%d", port)); err != nil {
		log.Fatalf("could not run server: %v", err)
	}
}

func (s *Service) setupServiceRouter() *gin.Engine {
	r := gin.Default()

	r.PUT("/feeds", s.createFeed)
	r.GET("/feeds", s.listFeeds)
	r.GET("/feeds/:id", s.getFeed)
	r.POST("/feeds/load", s.loadFeed)

	r.GET("/articles", s.listArticles)
	r.GET("/articles/:id", s.getArticle)

	return r
}

// CreateFeedArgs represents the arguments in a create feed request.
type CreateFeedArgs struct {
	Provider string `json:"provider" binding:"required"`
	Category string `json:"category" binding:"required"`
	Address  string `json:"address" binding:"required"`
}

func (s *Service) createFeed(c *gin.Context) {
	var args CreateFeedArgs
	if c.BindJSON(&args) != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid arguments",
		})
		return
	}
	feed, err := s.feedStore.Create(&types.Feed{
		Provider: args.Provider,
		Category: args.Category,
		Address:  args.Address,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, feed)
}

// GetFeedArgs represents the arguments in a get feed request.
type GetFeedArgs struct {
	ID string `uri:"id" binding:"required"`
}

func (s *Service) getFeed(c *gin.Context) {
	var args GetFeedArgs
	if c.BindUri(&args) != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid arguments",
		})
		return
	}
	feed, err := s.feedStore.Get(args.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, feed)
}

func (s *Service) listFeeds(c *gin.Context) {
	feeds, err := s.feedStore.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, feeds)
}

// LoadFeedArgs represents the arguments in a load feed request.
type LoadFeedArgs struct {
	ID string `json:"id" binding:"required"`
}

func (s *Service) loadFeed(c *gin.Context) {
	var args LoadFeedArgs
	if c.BindJSON(&args) != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid arguments",
		})
		return
	}
	feed, err := s.feedStore.Get(args.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	err = s.feeder.Consume(feed)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
	}
	return
}

// GetArticleArgs represents the arguments in a get article request.
type GetArticleArgs struct {
	ID string `uri:"id" binding:"required"`
}

func (s *Service) getArticle(c *gin.Context) {
	var args GetArticleArgs
	if c.BindUri(&args) != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid arguments",
		})
		return
	}
	article, err := s.articleStore.Get(args.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, article)
}

// ListArgs represents the arguments accepted in a list articles request.
type ListArgs struct {
	Cursor     string   `form:"c"`
	PageSize   int      `form:"pageSize"`
	Feed       string   `form:"feed"`
	Categories []string `form:"cat"`
}

func (s *Service) listArticles(c *gin.Context) {
	var args ListArgs
	if c.BindQuery(&args) != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid arguments",
		})
		return
	}

	articles, err := s.articleStore.List(args.Cursor, args.PageSize, args.Feed, args.Categories...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, articles)
}
