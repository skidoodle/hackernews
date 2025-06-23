package hn

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"sync"

	"hackernews/internal/cache"
	"hackernews/internal/config"
)

type Client struct {
	httpClient  *http.Client
	itemCache   *cache.Cache[*Item]
	userCache   *cache.Cache[*User]
	idListCache *cache.Cache[[]int]
	logger      *slog.Logger
	cfg         *config.HackerNewsAPIConfig
}

func (c *Client) GetUser(ctx context.Context, id string) (*User, error) {
	cacheKey := fmt.Sprintf("user:%s", id)
	if cached, found := c.userCache.Get(cacheKey); found {
		return cached, nil
	}

	url := fmt.Sprintf("%s/user/%s.json", c.cfg.BaseURL, id)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request for user %s: %w", id, err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user %s: %w", id, err)
	}
	defer resp.Body.Close()

	var user User
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, fmt.Errorf("failed to decode user %s: %w", id, err)
	}

	c.userCache.Set(cacheKey, &user)
	return &user, nil
}

func (c *Client) GetItemsByIDs(ctx context.Context, ids []int) ([]*Item, error) {
	items := make([]*Item, len(ids))
	jobs := make(chan int, len(ids))
	results := make(chan *Item, len(ids))

	var wg sync.WaitGroup
	for i := 0; i < c.cfg.WorkerCount; i++ {
		wg.Add(1)
		go c.storyWorker(ctx, &wg, jobs, results)
	}

	for _, id := range ids {
		jobs <- id
	}
	close(jobs)

	wg.Wait()
	close(results)

	idToItemMap := make(map[int]*Item, len(ids))
	for item := range results {
		if item != nil {
			idToItemMap[item.ID] = item
		}
	}

	for i, id := range ids {
		items[i] = idToItemMap[id]
	}

	return items, nil
}

func (c *Client) GetStoriesForPage(ctx context.Context, storyType string, page int) ([]*Item, error) {
	allStoryIDs, err := c.GetStoryIDs(ctx, storyType)
	if err != nil {
		return nil, err
	}

	start := (page - 1) * c.cfg.ItemsPerPage
	end := start + c.cfg.ItemsPerPage

	if start >= len(allStoryIDs) {
		return []*Item{}, nil
	}
	if end > len(allStoryIDs) {
		end = len(allStoryIDs)
	}

	return c.GetItemsByIDs(ctx, allStoryIDs[start:end])
}

func NewClient(logger *slog.Logger, cfg *config.Config) *Client {
	return &Client{
		httpClient:  &http.Client{},
		itemCache:   cache.New[*Item](cfg.Cache.ItemTTL * 2),
		userCache:   cache.New[*User](cfg.Cache.ItemTTL * 2),
		idListCache: cache.New[[]int](cfg.Cache.ItemTTL),
		logger:      logger,
		cfg:         &cfg.HackerNewsAPI,
	}
}

func (c *Client) GetItem(ctx context.Context, id int) (*Item, error) {
	item, err := c.fetchItem(ctx, id)
	if err != nil {
		return nil, err
	}

	if len(item.Kids) > 0 {
		item.Comments = c.fetchComments(ctx, item.Kids)
	}

	return item, nil
}

func (c *Client) GetStoryIDs(ctx context.Context, storyType string) ([]int, error) {
	cacheKey := fmt.Sprintf("idlist:%s", storyType)
	if cached, found := c.idListCache.Get(cacheKey); found {
		c.logger.Info("serving story ID list from cache", "type", storyType)
		return cached, nil
	}

	c.logger.Info("fetching story ID list from API", "type", storyType)
	url := fmt.Sprintf("%s/%sstories.json", c.cfg.BaseURL, storyType)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request for story IDs: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch story IDs: %w", err)
	}
	defer resp.Body.Close()

	var ids []int
	if err := json.NewDecoder(resp.Body).Decode(&ids); err != nil {
		return nil, fmt.Errorf("failed to decode story IDs: %w", err)
	}

	c.idListCache.Set(cacheKey, ids)
	return ids, nil
}

func (c *Client) fetchItem(ctx context.Context, id int) (*Item, error) {
	if cachedItem, found := c.itemCache.Get(fmt.Sprintf("item:%d", id)); found {
		return cachedItem, nil
	}

	url := fmt.Sprintf("%s/item/%d.json", c.cfg.BaseURL, id)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request for item %d: %w", id, err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch item %d: %w", id, err)
	}
	defer resp.Body.Close()

	var item Item
	if err := json.NewDecoder(resp.Body).Decode(&item); err != nil {
		return nil, fmt.Errorf("failed to decode item %d: %w", id, err)
	}

	if !item.Deleted && !item.Dead {
		c.itemCache.Set(fmt.Sprintf("item:%d", id), &item)
	}

	return &item, nil
}

func (c *Client) storyWorker(ctx context.Context, wg *sync.WaitGroup, jobs <-chan int, results chan<- *Item) {
	defer wg.Done()
	for id := range jobs {
		item, err := c.fetchItem(ctx, id)
		if err != nil {
			c.logger.Error("failed to fetch story item", "id", id, "error", err)
			continue
		}
		results <- item
	}
}

func (c *Client) fetchComments(ctx context.Context, ids []int) []*Item {
	comments := make([]*Item, 0, len(ids))
	for _, id := range ids {
		comment, err := c.GetItem(ctx, id)
		if err != nil {
			c.logger.Error("failed to fetch comment", "id", id, "error", err)
			continue
		}
		if comment != nil && !comment.Deleted && !comment.Dead {
			comments = append(comments, comment)
		}
	}
	return comments
}
