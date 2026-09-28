package search

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"libgen-gui/internal/network"
)

type mockSearchService struct {
	results []Result
	err     error
	calls   atomic.Int32
}

func (m *mockSearchService) Search(ctx context.Context, query Query) ([]Result, error) {
	m.calls.Add(1)
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	return m.results, m.err
}

func (m *mockSearchService) GetDetails(ctx context.Context, idOrHash string) (*Details, error) {
	if len(m.results) > 0 {
		r := m.results[0]
		return &Details{
			ID:      r.ID,
			Title:   r.Title,
			Authors: r.Authors,
			Format:  r.Format,
			Size:    r.Size,
			Md5:     r.Md5,
		}, nil
	}
	return nil, nil
}

func TestSearchController_DebounceAndCancelStale(t *testing.T) {
	mockSvc := &mockSearchService{
		results: []Result{
			{Title: "Concurrent Go", Authors: []string{"Alan"}, Year: 2024, Format: "pdf", Size: 1048576, Md5: "md5_concurrent"},
		},
	}

	ctrl := NewSearchController(mockSvc, 100*time.Millisecond)

	var mu sync.Mutex
	var lastResults []Result
	var lastErr error

	// Rapid typing simulation (G -> Go -> Gol -> Golang)
	for _, q := range []string{"G", "Go", "Gol", "Golang"} {
		ctrl.Search(Query{Text: q}, func(results []Result, err error) {
			mu.Lock()
			lastResults = results
			lastErr = err
			mu.Unlock()
		})
		time.Sleep(30 * time.Millisecond) // Faster than 100ms debounce
	}

	// Wait for debounce timer to fire once
	time.Sleep(200 * time.Millisecond)

	if calls := mockSvc.calls.Load(); calls != 1 {
		t.Errorf("expected exactly 1 debounced search call, got %d", calls)
	}

	mu.Lock()
	defer mu.Unlock()
	if lastErr != nil {
		t.Fatalf("unexpected search error: %v", lastErr)
	}
	if len(lastResults) != 1 || lastResults[0].Title != "Concurrent Go" {
		t.Errorf("unexpected results: %+v", lastResults)
	}
}

func TestSearchService_RealHTMLEngineParsing(t *testing.T) {
	htmlTable := `
	<html><body>
	<table id="tablelibgen">
		<tr>
			<td><a href="/edition.php?id=123">Programming in Go</a></td>
			<td>Author Name</td>
			<td>Tech Press</td>
			<td>2023</td>
			<td>English</td>
			<td>350</td>
			<td>5 MB</td>
			<td>pdf</td>
			<td><a href="/ads.php?md5=abcdef1234567890abcdef1234567890">GET</a></td>
		</tr>
	</table>
	</body></html>`

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(htmlTable))
	}))
	defer ts.Close()

	client := network.NewClient()
	mirrorMgr := network.NewMirrorManager(client, []string{ts.URL})
	svc := NewService(mirrorMgr, client)

	results, err := svc.Search(context.Background(), Query{Text: "Programming"})
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(results) == 0 {
		t.Fatalf("expected search results, got 0")
	}

	r0 := results[0]
	if r0.Title != "Programming in Go" {
		t.Errorf("expected title 'Programming in Go', got %q", r0.Title)
	}
	if r0.Format != "pdf" {
		t.Errorf("expected format 'pdf', got %q", r0.Format)
	}
}
