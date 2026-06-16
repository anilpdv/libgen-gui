package ui

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync/atomic"
	"testing"
	"time"
)

func TestMirrorHealthManager_ProbingAndSubscription(t *testing.T) {
	// Setup mock servers
	okServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("libgen mock ok"))
	}))
	defer okServer.Close()

	failServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte("503 Service Unavailable"))
	}))
	defer failServer.Close()

	okURL, _ := url.Parse(okServer.URL)
	failURL, _ := url.Parse(failServer.URL)

	mgr := NewMirrorHealthManager()
	customMirrors := []MirrorInfo{
		{
			Host:     okURL.Host,
			ProbeURL: okServer.URL,
			Role:     RoleSearchAndDownload,
			State:    StateChecking,
		},
		{
			Host:     failURL.Host,
			ProbeURL: failServer.URL,
			Role:     RoleSearchAndDownload,
			State:    StateChecking,
		},
		{
			Host:     "offline.invalid.domain",
			ProbeURL: "https://offline.invalid.domain/index.php",
			Role:     RoleIPFSGateway,
			State:    StateChecking,
		},
	}
	mgr.SetCustomMirrors(customMirrors)

	var subFired int32
	unsub := mgr.Subscribe(func(activeSearch, totalSearch, activeAll, totalAll int) {
		atomic.AddInt32(&subFired, 1)
	})
	defer unsub()

	// Probe all mirrors
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	mgr.ProbeAll(ctx)

	mirrors := mgr.GetMirrors()
	if len(mirrors) != 3 {
		t.Fatalf("expected 3 mirrors, got %d", len(mirrors))
	}

	// 1. Check OK server
	if mirrors[0].State != StateOnline {
		t.Errorf("expected mirrors[0] StateOnline, got %v", mirrors[0].State)
	}
	if mirrors[0].StatusCode != 200 {
		t.Errorf("expected mirrors[0] StatusCode 200, got %d", mirrors[0].StatusCode)
	}
	if mirrors[0].Latency <= 0 {
		t.Errorf("expected mirrors[0] positive latency, got %v", mirrors[0].Latency)
	}

	// 2. Check 503 server
	if mirrors[1].State != StateOffline {
		t.Errorf("expected mirrors[1] StateOffline, got %v", mirrors[1].State)
	}
	if mirrors[1].StatusCode != 503 {
		t.Errorf("expected mirrors[1] StatusCode 503, got %d", mirrors[1].StatusCode)
	}

	// 3. Check invalid domain
	if mirrors[2].State != StateOffline {
		t.Errorf("expected mirrors[2] StateOffline, got %v", mirrors[2].State)
	}

	// 4. Check active counts
	activeSearch, totalSearch := mgr.GetActiveSearchCount()
	if activeSearch != 1 || totalSearch != 2 {
		t.Errorf("expected activeSearch 1/2, got %d/%d", activeSearch, totalSearch)
	}

	activeAll, totalAll := mgr.GetActiveAllCount()
	if activeAll != 1 || totalAll != 3 {
		t.Errorf("expected activeAll 1/3, got %d/%d", activeAll, totalAll)
	}

	if atomic.LoadInt32(&subFired) == 0 {
		t.Errorf("expected subscriber callback to have fired at least once")
	}

	// 5. Test ProbeSingle
	res := mgr.ProbeSingle(ctx, okURL.Host)
	if res.State != StateOnline || res.StatusCode != 200 {
		t.Errorf("ProbeSingle failed: %+v", res)
	}
}
