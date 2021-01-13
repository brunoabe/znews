// Package main creates a znews service and listens for RESTful requests.
package main

import (
	"./feedconsumer"
	"./rssreader"
	"./service"
	"./store"
)

const servicePort = 8052

func main() {
	feedStore := store.NewFeedStore()
	articleStore := store.NewArticleStore()
	feed := rssreader.NewFeed()
	consumer := feedconsumer.NewFeedConsumer(feed, articleStore)

	s := service.NewService(consumer, feedStore, articleStore)
	s.ServeForever(servicePort)
}
