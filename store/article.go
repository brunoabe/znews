package store

import (
	"errors"
	"sync"

	"github.com/google/uuid"

	"../types"
)

// ArticleStore provides storage functionality for articles.
type ArticleStore struct {
	mu            sync.RWMutex
	a             []*types.Article
	m             map[string]*types.Article
	uuidNamespace uuid.UUID
}

// NewArticleStore returns a new Article Store.
func NewArticleStore() *ArticleStore {
	return &ArticleStore{
		a:             []*types.Article{},
		m:             map[string]*types.Article{},
		uuidNamespace: uuid.MustParse(uuidNamespace),
	}
}

// Reset clears the store to its initial state.
func (as *ArticleStore) Reset() {
	as.mu.Lock()
	defer as.mu.Unlock()
	as.a = []*types.Article{}
	as.m = map[string]*types.Article{}
}

// Create stores the provided article in the store in the correct order by publish date and returns
// the saved item. If the GUID is already present in the store, it will just return the existing
// item, discarding the provided value.
func (as *ArticleStore) Create(article *types.Article) (*types.Article, error) {
	if article == nil {
		return nil, nil
	}
	generatedID := uuid.NewSHA1(as.uuidNamespace, []byte(article.GUID)).String()
	if a, ok := as.m[generatedID]; ok {
		return a, nil
	}
	article.ID = generatedID
	as.mu.Lock()
	defer as.mu.Unlock()
	// To maintain the order of the slice, whenever a new element is added, it is injected in order.
	// This is an expensive operation for writes, but is optimal for reading.

	// If it is already the newer item, append it to the end.
	if len(as.a) == 0 || !as.a[len(as.a)-1].PublishDate.After(article.PublishDate) {
		as.a = append(as.a, article)
		as.m[generatedID] = article
		return article, nil
	}

	// If the article is the oldest one, append to the beginning.
	if !article.PublishDate.After(as.a[0].PublishDate) {
		as.a = append([]*types.Article{article}, as.a...)
		as.m[generatedID] = article
		return article, nil
	}

	// The check is done in backwards because it is likely that new articles will have newer publish
	// dates.
	for i := len(as.a) - 2; i >= 0; i-- {
		if article.PublishDate.After(as.a[i].PublishDate) || i == 0 {
			as.a = append(as.a[:i+1], as.a[i:]...)
			as.a[i+1] = article
			as.m[generatedID] = article
			break
		}
	}
	return article, nil
}

// List reads articles from the store and returns the requested number of articles starting from the
// provided cursor. Since the store will be ordered by publish date, if a newer article is added in
// between calls, it might not be returned unless a new call to the endpoint is made with an earlier
// cursor. If no categories are provided, no filter will be applied. If categories are provided, the
// filtering will bypass any news for any category provided.
func (as *ArticleStore) List(cursor string, pageSize int, feed string, categories ...string) ([]*types.Article, error) {
	// Create a hashmap for filtering.
	cat := make(map[string]struct{}, len(categories))
	for _, c := range categories {
		cat[c] = struct{}{}
	}

	as.mu.RLock()
	defer as.mu.RUnlock()

	firstReturnIndex, ok := as.findArticleCursorIndex(cursor)
	if !ok {
		return nil, errors.New("could not find provided cursor")
	}
	found := 0
	var res []*types.Article
	for i := firstReturnIndex; i < len(as.a); i++ {
		current := as.a[i]
		if len(cat) > 0 {
			// Must do some filtering on categories.
			if len(current.Categories) == 0 {
				continue
			}
			hasCategory := false
			for _, c := range current.Categories {
				if _, ok := cat[c]; ok {
					hasCategory = true
					break
				}
			}
			if !hasCategory {
				continue
			}
		}
		if feed != "" {
			// Must do filtering on feed.
			if current.FeedID != feed {
				continue
			}
		}
		res = append(res, current)
		found++
		if found == pageSize {
			break
		}
	}

	return res, nil
}

// Get returns an article from the store based on its GUID if it exists. Returns an error otherwise.
func (as *ArticleStore) Get(ID string) (*types.Article, error) {
	if ID == "" {
		return nil, errors.New("invalid ID provided")
	}
	as.mu.RLock()
	defer as.mu.RUnlock()
	if _, ok := as.m[ID]; !ok {
		return nil, errors.New("resource not found")
	}
	return as.m[ID], nil
}

// findArticleCursorIndex returns the slice index for the article that has the cursor as its ID. If
// it fails to find the article, the second return argument will be false.
func (as *ArticleStore) findArticleCursorIndex(cursor string) (int, bool) {
	if cursor == "" {
		return 0, true
	}
	for i, a := range as.a {
		if a.ID == cursor {
			return i + 1, true
		}
	}
	return 0, false
}
