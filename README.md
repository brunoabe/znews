# znews

A service capable of reading rss news feeds and storing the articles, making them available throughout a RESTful API.

## Usage

ZNews is a service that provides an API capable of storing rss feeds, allowing the provided addresses to be loaded via a custom endpoint call. Two examples are provided to create the service:

### Running the program locally

To use the system, the following dependencies should be installed:

```
go get github.com/stretchr/testify
go get github.com/ungerik/go-rss
go get github.com/gin-gonic/gin
go get github.com/google/uuid
```

With the dependencies available, one can go to the root directory of the repo and run:

```
go run main.go
```

### Running the program in a Docker container

The service has capabilities of being built in a distroless container and run locally provided one has got Docker installed. To do so, the following command should be run once to build the image:

```
docker build . -t znews
```

Once the image is already present locally, one should run the image and expose the service port locally using the following command:

```
docker run -p 8052:8052 znews
```

## Testing

This repo contains a set of tests that can be run, most of them are unit tests, but some end to end tests are provided considering a hypothetical use of the API, explaining some happy path scenarios.

All tests should be run using the Golang test framework. To do that, one needs to navigate to any folder containing a file in the format "<package>_test" and run the following command:

```
go test -v
```

_*Note*: Running the above command under `znews/e2e` will run the end to end test scenarios. To do that, a new service will be spin-up locally and automatic calls will be made to verify the behaviour.

# Documentation

## News Feeds

The API allows storing news feed addresses, whereby a custom endpoint allow loading news from such feed. The following endpoints are provided:

### CreateFeed

Allows the storage of a news feeded by providing the news provider, the category and the rss feed address.

*Example*
```
curl -v -X PUT \
  "http://localhost:8052/feeds" \
  -H 'content-type: application/json' \
  -d '{ "provider": "BBC News", "category": "UK", "address": "http://feeds.bbci.co.uk/news/uk/rss.xml" }'
```

### ListFeeds

Lists all feeds available in the system. It shows all feed information and could be used by the consumer to get which feeds are for which providers or even of a given category.

*Example*
```
curl -v -X GET \
  "http://localhost:8052/feeds"
```

### GetFeed

Return a single fees stored by its ID.

*Example*
```
curl -v -X GET \
  "http://localhost:8052/feeds/0792cd43-d8f3-5a38-9739-c797bd08c6fa"
```

### LoadFeed

Fetches information from the rss feed that was previously created in the system by its respective ID. Loading data multiple times are going to be additive operations where new articles are going to be stored and existing ones disregarded. The API will consider the field GUID from the feed to be unique globally and will use it to generate a hash for being the ID of each article.

*Example*
```
curl -v -X POST \
  "http://localhost:8052/feeds/load" \
  -H 'content-type: application/json' \
  -d '{ "id": "0792cd43-d8f3-5a38-9739-c797bd08c6fa" }'
```

_Note*: Because the ID of the feed is a hash of its address, the above example should work for the inserted feed above._

## Articles

Once a Feed has been added to the system and news from it are loaded, articles are going to be available for consumption.

### ListArticles

Articles can be retrieved from the system using the `List` endpoint. It returns all data unless `pageSize` is informed. If there is a page size, the API will paginate the results giving the first set of articles in the first call. It uses cursor based pagination, so to retrieve the next pages, the last ID retrieved in the previous call must be informed. The respose of this endpoint is ordered by publish date.

The same endpoint also allow for filtering on categories. The category might be available or not in the news feed, if there are no matches for the category informed, the API will return an empty response. Multiple categories are allowed and the API will return any article containing any of the informed categories.

*Example*
```
curl -v -X GET \ 
  "http://localhost:8052/articles"
```

_*Note*: If the query parameter for pageSize is not informed, the API will return all available data._

```
curl -v -X GET \
  "http://localhost:8052/articles?pageSize=5"
```

_*Note*: If the query parameter for pageSize is informed and no cursor is infored, the API will return the first page of articles._

```
curl -v -X GET \
  "http://localhost:8052/articles?pageSize=5&c=c77397a6-163a-56df-9e22-8e29ea7a62b5"
```

_*Note*: If the query parameter for pageSize is informed and a cursor is infored, the API will return the next page of articles starting from the next article from the informed cursor._

```
curl -v -X GET \
  "http://localhost:8052/articles?pageSize=5&feed=0792cd43-d8f3-5a38-9739-c797bd08c6fa"
```

_*Note*: If the query parameter for feed is informed with a feed ID, the API will filter only articles for such feed to be returned._

```
curl -v -X GET \
  "http://localhost:8052/articles?pageSize=5&cat=Technology&cat=UK"
```

_*Note*: If the query parameter for categories is informed, the API will return filtered data based on the category field of the rss feed. If the field doesn't support that and any category is informed, the API will return an empty response._

### GetArticle

If the intention is to get a single resource by its ID, the GetArticle endpoint is the right choice, it returns all information of a single article provided an ID is informed.

*Example*

```
curl -v -X GET \
  "http://localhost:8052/articles/7b485edd-4f46-56c9-8c08-1db5dda37624"  
```
