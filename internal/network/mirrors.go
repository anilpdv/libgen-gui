package network

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/url"
	"sort"
	"sync"
	"time"
)

var (
	ErrNoHealthyMirror = errors.New("no healthy mirrors available")
)

// DefaultSearchMirrors contains official standard search mirrors.
var DefaultSearchMirrors = []string{
	"https://libgen.li/index.php",
	"https://libgen.vg/index.php",
}

// DefaultDownloadMirrors contains official standard download mirrors.
var DefaultDownloadMirrors = []string{
	"https://libgen.li/ads.php",
	"https://libgen.vg/ads.php",
}

// Mirror tracks live health, latency, and failure metrics for a single mirror.
type Mirror struct {
	rawURL string
	parsed url.URL

	mu                  sync.RWMutex
	latency             time.Duration
	consecutiveFailures int
	lastSuccess         time.Time
	lastFailure         time.Time
	lastProbe           time.Time
	healthy             bool
}

// NewMirror creates a new initialized Mirror.
func NewMirror(rawURL string) *Mirror {
	parsed, _ := url.Parse(rawURL)
	return &Mirror{
		rawURL:  rawURL,
		parsed:  *parsed,
		healthy: true, // Optimistically healthy until probed
	}
}

// MirrorSnapshot is an immutable value object representing the state of a mirror at a point in time.
type MirrorSnapshot struct {
	URL                 string
	Host                string
	Latency             time.Duration
	ConsecutiveFailures int
	Healthy             bool
	InCooldown          bool
	CooldownRemaining   time.Duration
	LastSuccess         time.Time
	LastFailure         time.Time
}

func (m *Mirror) Snapshot() MirrorSnapshot {
	m.mu.RLock()
	defer m.mu.RUnlock()

	now := time.Now()
	cd := m.cooldownDurationUnderLock()
	inCD := false
	var remaining time.Duration

	if m.consecutiveFailures > 0 && !m.lastFailure.IsZero() {
		elapsed := now.Sub(m.lastFailure)
		if elapsed < cd {
			inCD = true
			remaining = cd - elapsed
		}
	}

	return MirrorSnapshot{
		URL:                 m.rawURL,
		Host:                m.parsed.Host,
		Latency:             m.latency,
		ConsecutiveFailures: m.consecutiveFailures,
		Healthy:             m.healthy && !inCD,
		InCooldown:          inCD,
		CooldownRemaining:   remaining,
		LastSuccess:         m.lastSuccess,
		LastFailure:         m.lastFailure,
	}
}

func (m *Mirror) cooldownDurationUnderLock() time.Duration {
	switch {
	case m.consecutiveFailures >= 5:
		return 5 * time.Minute
	case m.consecutiveFailures >= 3:
		return 1 * time.Minute
	case m.consecutiveFailures >= 1:
		return 10 * time.Second
	default:
		return 0
	}
}

func (m *Mirror) RecordSuccess(latency time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.latency = latency
	m.consecutiveFailures = 0
	m.lastSuccess = time.Now()
	m.healthy = true
}

func (m *Mirror) RecordFailure() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.consecutiveFailures++
	m.lastFailure = time.Now()
	m.healthy = false
}

// MirrorListener is notified when mirror statuses change.
type MirrorListener func(snapshots []MirrorSnapshot)

// MirrorManager manages mirror selection, probing, cooldowns, and health monitoring.
type MirrorManager struct {
	httpClient *http.Client
	mirrors    []*Mirror
	mu         sync.RWMutex
	listeners  []MirrorListener
}

// NewMirrorManager constructs a MirrorManager with candidate mirror URLs.
func NewMirrorManager(httpClient *http.Client, mirrorURLs []string) *MirrorManager {
	if httpClient == nil {
		httpClient = NewClient()
	}

	mgr := &MirrorManager{
		httpClient: httpClient,
		mirrors:    make([]*Mirror, 0, len(mirrorURLs)),
	}

	for _, u := range mirrorURLs {
		mgr.mirrors = append(mgr.mirrors, NewMirror(u))
	}

	return mgr
}

// AddMirror registers a new mirror URL.
func (m *MirrorManager) AddMirror(rawURL string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, existing := range m.mirrors {
		if existing.rawURL == rawURL {
			return
		}
	}

	m.mirrors = append(m.mirrors, NewMirror(rawURL))
}

// Snapshots returns current immutable status snapshots for all managed mirrors.
func (m *MirrorManager) Snapshots() []MirrorSnapshot {
	m.mu.RLock()
	mirrors := append([]*Mirror(nil), m.mirrors...)
	m.mu.RUnlock()

	snapshots := make([]MirrorSnapshot, len(mirrors))
	for i, mir := range mirrors {
		snapshots[i] = mir.Snapshot()
	}
	return snapshots
}

// Healthy returns snapshots of all currently healthy, non-cooldown mirrors.
func (m *MirrorManager) Healthy() []MirrorSnapshot {
	snapshots := m.Snapshots()
	healthy := make([]MirrorSnapshot, 0, len(snapshots))
	for _, s := range snapshots {
		if s.Healthy {
			healthy = append(healthy, s)
		}
	}
	return healthy
}

// Best returns the lowest-latency healthy mirror, or ErrNoHealthyMirror if none are ready.
func (m *MirrorManager) Best() (MirrorSnapshot, error) {
	healthy := m.Healthy()
	if len(healthy) == 0 {
		// If none are healthy, fall back to least-failed mirror
		all := m.Snapshots()
		if len(all) == 0 {
			return MirrorSnapshot{}, ErrNoHealthyMirror
		}
		sort.Slice(all, func(i, j int) bool {
			return all[i].ConsecutiveFailures < all[j].ConsecutiveFailures
		})
		return all[0], nil
	}

	sort.Slice(healthy, func(i, j int) bool {
		if healthy[i].Latency == 0 {
			return false
		}
		if healthy[j].Latency == 0 {
			return true
		}
		return healthy[i].Latency < healthy[j].Latency
	})

	return healthy[0], nil
}

// Subscribe registers a listener callback called on health changes.
func (m *MirrorManager) Subscribe(fn MirrorListener) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.listeners = append(m.listeners, fn)
}

func (m *MirrorManager) notifyListeners() {
	m.mu.RLock()
	listeners := append([]MirrorListener(nil), m.listeners...)
	m.mu.RUnlock()

	snapshots := m.Snapshots()
	for _, fn := range listeners {
		fn(snapshots)
	}
}

// Probe probes a single mirror by URL and updates its latency and health metrics without holding the manager lock.
func (m *MirrorManager) Probe(ctx context.Context, target *Mirror) error {
	start := time.Now()
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, target.rawURL, nil)
	if err != nil {
		target.RecordFailure()
		return err
	}
	req.Header.Set("User-Agent", DefaultUserAgent)

	resp, err := m.httpClient.Do(req)
	latency := time.Since(start)

	if err != nil {
		target.RecordFailure()
		slog.Debug("mirror probe failed", "url", target.rawURL, "error", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 400 {
		target.RecordSuccess(latency)
		slog.Debug("mirror probe succeeded", "url", target.rawURL, "latency_ms", latency.Milliseconds())
		return nil
	}

	target.RecordFailure()
	slog.Debug("mirror probe returned error status", "url", target.rawURL, "status", resp.StatusCode)
	return &NetworkError{Kind: ErrorHTTP, StatusCode: resp.StatusCode, URL: target.rawURL, Err: errors.New(resp.Status)}
}

// ProbeAll probes all mirrors concurrently and notifies subscribers.
func (m *MirrorManager) ProbeAll(ctx context.Context) {
	m.mu.RLock()
	mirrors := append([]*Mirror(nil), m.mirrors...)
	m.mu.RUnlock()

	var wg sync.WaitGroup
	for _, mir := range mirrors {
		snap := mir.Snapshot()
		if snap.InCooldown {
			continue
		}

		wg.Add(1)
		go func(target *Mirror) {
			defer wg.Done()
			probeCtx, cancel := context.WithTimeout(ctx, 4*time.Second)
			defer cancel()
			_ = m.Probe(probeCtx, target)
		}(mir)
	}

	wg.Wait()
	m.notifyListeners()
}
