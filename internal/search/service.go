package search

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"libgen-gui/internal/network"
	"libgen-gui/pkg/libgen"
)

// SearchService defines the interface between UI components and backend search logic.
type SearchService interface {
	Search(ctx context.Context, query Query) ([]Result, error)
	GetDetails(ctx context.Context, idOrHash string) (*Details, error)
}

// Service implements SearchService using mirror failover, connection pooling, and context cancellation.
type Service struct {
	mirrorManager *network.MirrorManager
	httpClient    *http.Client
}

// NewService constructs a new search Service.
func NewService(mirrorManager *network.MirrorManager, httpClient *http.Client) *Service {
	if httpClient == nil {
		httpClient = network.NewClient()
	}
	if mirrorManager == nil {
		mirrorManager = network.NewMirrorManager(httpClient, network.DefaultSearchMirrors)
	}
	return &Service{
		mirrorManager: mirrorManager,
		httpClient:    httpClient,
	}
}

// Search queries the fastest healthy mirror with context propagation and domain model mapping.
func (s *Service) Search(ctx context.Context, query Query) ([]Result, error) {
	if strings.TrimSpace(query.Text) == "" {
		return nil, nil
	}

	bestMirror, err := s.mirrorManager.Best()
	if err != nil {
		return nil, err
	}

	mirrorURL, err := url.Parse(bestMirror.URL)
	if err != nil {
		mirrorURL, _ = url.Parse("https://libgen.li/index.php")
	}

	opts := &libgen.SearchOptions{
		Query:        query.Text,
		SearchMirror: *mirrorURL,
		Results:      query.PageSize,
		Page:         query.Page,
		Context:      ctx,
		SortBy:       query.SortBy,
		SortASC:      query.SortASC,
		Publisher:    query.Publisher,
		Language:     query.Language,
	}

	if opts.Results <= 0 {
		opts.Results = 25
	}
	if opts.Page <= 0 {
		opts.Page = 1
	}

	if query.Format != "" && !strings.EqualFold(query.Format, "any") {
		opts.Extension = []string{query.Format}
		opts.AutoPaging = true
		opts.MaxPages = 5
	}
	if query.YearFrom > 0 && query.YearTo > 0 && query.YearFrom == query.YearTo {
		opts.Year = query.YearFrom
	}

	books, err := libgen.Search(opts)
	if err != nil {
		return nil, network.ClassifyError(err, mirrorURL.String(), 0)
	}

	results := make([]Result, 0, len(books))
	for _, b := range books {
		if b == nil {
			continue
		}
		var authors []string
		if b.Author != "" {
			for _, a := range strings.Split(b.Author, ",") {
				if trimmed := strings.TrimSpace(a); trimmed != "" {
					authors = append(authors, trimmed)
				}
			}
		}

		yr, _ := strconv.Atoi(b.Year)
		sz, _ := strconv.ParseInt(b.Filesize, 10, 64)

		results = append(results, Result{
			ID:          b.ID,
			Title:       b.Title,
			Authors:     authors,
			Publisher:   b.Publisher,
			Year:        yr,
			Language:    b.Language,
			Format:      strings.ToLower(b.Extension),
			Size:        sz,
			Pages:       b.Pages,
			Md5:         strings.ToLower(b.Md5),
			SourceID:    b.ID,
			DownloadURL: b.DownloadURL,
			PageURL:     b.PageURL,
			CoverURL:    b.CoverURL,
		})
	}

	return results, nil
}

// GetDetails resolves deep item metadata given an MD5 hash or unique ID.
func (s *Service) GetDetails(ctx context.Context, idOrHash string) (*Details, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	bestMirror, _ := s.mirrorManager.Best()
	mirrorURL, _ := url.Parse(bestMirror.URL)
	if mirrorURL == nil || mirrorURL.Host == "" {
		mirrorURL, _ = url.Parse("https://libgen.li/index.php")
	}

	books, err := libgen.GetDetails(&libgen.GetDetailsOptions{
		Hashes:       []string{idOrHash},
		SearchMirror: *mirrorURL,
	})
	if err != nil || len(books) == 0 {
		if err == nil {
			err = errors.New("item not found")
		}
		return nil, network.ClassifyError(err, mirrorURL.String(), 0)
	}

	b := books[0]
	var authors []string
	if b.Author != "" {
		for _, a := range strings.Split(b.Author, ",") {
			if trimmed := strings.TrimSpace(a); trimmed != "" {
				authors = append(authors, trimmed)
			}
		}
	}

	yr, _ := strconv.Atoi(b.Year)
	sz, _ := strconv.ParseInt(b.Filesize, 10, 64)

	return &Details{
		ID:          b.ID,
		Title:       b.Title,
		Authors:     authors,
		Publisher:   b.Publisher,
		Year:        yr,
		Language:    b.Language,
		Format:      strings.ToLower(b.Extension),
		Size:        sz,
		Pages:       b.Pages,
		Edition:     b.Edition,
		CoverURL:    b.CoverURL,
		DownloadURL: b.DownloadURL,
		PageURL:     b.PageURL,
		Md5:         strings.ToLower(b.Md5),
	}, nil
}

// SearchController coordinates debounced live searches and cancels stale in-flight requests.
type SearchController struct {
	service  SearchService
	debounce time.Duration

	mu     sync.Mutex
	cancel context.CancelFunc
	timer  *time.Timer
}

// NewSearchController creates a SearchController with debouncing.
func NewSearchController(service SearchService, debounce time.Duration) *SearchController {
	if debounce <= 0 {
		debounce = 400 * time.Millisecond
	}
	return &SearchController{
		service:  service,
		debounce: debounce,
	}
}

// Search cancels any existing in-flight search and executes query after debounce.
func (c *SearchController) Search(query Query, onResult func(results []Result, err error)) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.timer != nil {
		c.timer.Stop()
	}
	if c.cancel != nil {
		c.cancel()
	}

	ctx, cancel := context.WithCancel(context.Background())
	c.cancel = cancel

	c.timer = time.AfterFunc(c.debounce, func() {
		results, err := c.service.Search(ctx, query)
		if ctx.Err() != nil {
			return // Ignore results if cancelled by subsequent keystrokes
		}
		if onResult != nil {
			onResult(results, err)
		}
	})
}

// Cancel immediately halts any scheduled or running query.
func (c *SearchController) Cancel() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.timer != nil {
		c.timer.Stop()
	}
	if c.cancel != nil {
		c.cancel()
		c.cancel = nil
	}
}
