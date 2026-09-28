package network

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestClient_RedirectHeaders(t *testing.T) {
	var requestedReferer string
	var requestedRange string

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/redirect" {
			http.Redirect(w, r, "/target", http.StatusFound)
			return
		}
		if r.URL.Path == "/target" {
			requestedReferer = r.Header.Get("Referer")
			requestedRange = r.Header.Get("Range")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("ok"))
			return
		}
	}))
	defer ts.Close()

	client := NewClient(ClientConfig{Timeout: 5 * time.Second})

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, ts.URL+"/redirect", nil)
	if err != nil {
		t.Fatalf("unexpected request error: %v", err)
	}
	req.Header.Set("Range", "bytes=100-")

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("unexpected do error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
	if requestedReferer == "" {
		t.Errorf("expected referer to be populated on redirect")
	}
	if requestedRange != "bytes=100-" {
		t.Errorf("expected Range header to be preserved on redirect, got %q", requestedRange)
	}
}

func TestRetry_ExponentialBackoff(t *testing.T) {
	d0 := Backoff(0)
	d1 := Backoff(1)
	d2 := Backoff(2)

	if d0 > d1 || d1 > d2 {
		t.Errorf("expected increasing backoff delays, got %v, %v, %v", d0, d1, d2)
	}
}

func TestRetry_AfterServerFailure(t *testing.T) {
	var attempts atomic.Int32

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := attempts.Add(1)
		if n < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("success"))
	}))
	defer ts.Close()

	client := NewClient()

	err := Retry(context.Background(), 4, func(ctx context.Context, attempt int) error {
		req, _ := http.NewRequestWithContext(ctx, http.MethodGet, ts.URL, nil)
		resp, doErr := client.Do(req)
		if doErr != nil {
			return ClassifyError(doErr, ts.URL, 0)
		}
		defer resp.Body.Close()
		if resp.StatusCode >= 400 {
			return ClassifyError(errors.New(resp.Status), ts.URL, resp.StatusCode)
		}
		return nil
	})

	if err != nil {
		t.Fatalf("expected retry to succeed on 3rd attempt, got error: %v", err)
	}
	if attempts.Load() != 3 {
		t.Errorf("expected exactly 3 attempts, got %d", attempts.Load())
	}
}

func TestRetry_NonRetryableError(t *testing.T) {
	var attempts atomic.Int32

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts.Add(1)
		w.WriteHeader(http.StatusNotFound) // 404 is not retryable
	}))
	defer ts.Close()

	client := NewClient()

	err := Retry(context.Background(), 5, func(ctx context.Context, attempt int) error {
		req, _ := http.NewRequestWithContext(ctx, http.MethodGet, ts.URL, nil)
		resp, doErr := client.Do(req)
		if doErr != nil {
			return ClassifyError(doErr, ts.URL, 0)
		}
		defer resp.Body.Close()
		return ClassifyError(nil, ts.URL, resp.StatusCode)
	})

	if err == nil {
		t.Fatalf("expected 404 error, got nil")
	}
	if attempts.Load() != 1 {
		t.Errorf("expected exactly 1 attempt for non-retryable 404 error, got %d", attempts.Load())
	}
}

func TestMirrorManager_CooldownAndBest(t *testing.T) {
	m1Server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(50 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer m1Server.Close()

	m2Server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(10 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer m2Server.Close()

	mgr := NewMirrorManager(NewClient(), []string{m1Server.URL, m2Server.URL})

	mgr.ProbeAll(context.Background())

	best, err := mgr.Best()
	if err != nil {
		t.Fatalf("unexpected Best() error: %v", err)
	}

	if best.URL != m2Server.URL {
		t.Errorf("expected faster mirror %s to be chosen as best, got %s", m2Server.URL, best.URL)
	}

	// Fail m2 repeatedly to trigger cooldown
	m2 := mgr.mirrors[1]
	for i := 0; i < 5; i++ {
		m2.RecordFailure()
	}

	snap := m2.Snapshot()
	if !snap.InCooldown {
		t.Errorf("expected m2 to be in cooldown after 5 consecutive failures")
	}

	// Now m1 should be best
	bestAfterFail, err := mgr.Best()
	if err != nil {
		t.Fatalf("unexpected Best() error: %v", err)
	}
	if bestAfterFail.URL != m1Server.URL {
		t.Errorf("expected m1 to be chosen while m2 is in cooldown, got %s", bestAfterFail.URL)
	}
}
