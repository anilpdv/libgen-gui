package ui

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	"libgen-gui/pkg/libgen"
)

type MirrorState string

const (
	StateOnline   MirrorState = "Online"
	StateSlow     MirrorState = "Slow"
	StateOffline  MirrorState = "Offline"
	StateChecking MirrorState = "Checking"
)

type MirrorRole string

const (
	RoleSearchAndDownload MirrorRole = "Search & Download"
	RoleIPFSGateway       MirrorRole = "IPFS Gateway"
)

// MirrorInfo holds the real-time health and diagnostics of a single mirror.
type MirrorInfo struct {
	Host        string
	ProbeURL    string
	Role        MirrorRole
	State       MirrorState
	StatusCode  int
	Latency     time.Duration
	LastChecked time.Time
	ErrorMsg    string
}

// MirrorHealthManager monitors the health and reachability of LibGen mirrors and IPFS gateways.
type MirrorHealthManager struct {
	mu          sync.RWMutex
	mirrors     []MirrorInfo
	subscribers []func(activeSearch, totalSearch, activeAll, totalAll int)
	httpClient  *http.Client
	stopBg      chan struct{}
}

var (
	defaultHealthManager     *MirrorHealthManager
	defaultHealthManagerOnce sync.Once
)

// GetMirrorHealthManager returns the application-wide singleton health manager.
func GetMirrorHealthManager() *MirrorHealthManager {
	defaultHealthManagerOnce.Do(func() {
		defaultHealthManager = NewMirrorHealthManager()
	})
	return defaultHealthManager
}

// NewMirrorHealthManager initializes a new health manager with standard mirrors.
func NewMirrorHealthManager() *MirrorHealthManager {
	transport := &http.Transport{
		TLSClientConfig:       &tls.Config{InsecureSkipVerify: true},
		DisableKeepAlives:     true,
		ResponseHeaderTimeout: 4 * time.Second,
	}
	client := &http.Client{
		Transport: transport,
		Timeout:   4 * time.Second,
	}

	m := &MirrorHealthManager{
		httpClient: client,
		stopBg:     make(chan struct{}),
	}

	// Initialize default candidate list from SearchMirrors + IPFS gateways
	m.initDefaults()
	return m
}

func (m *MirrorHealthManager) initDefaults() {
	m.mu.Lock()
	defer m.mu.Unlock()

	var list []MirrorInfo
	// Primary Search & Download mirrors from libgen-cli
	for _, u := range libgen.SearchMirrors {
		probe := u.String()
		list = append(list, MirrorInfo{
			Host:        u.Host,
			ProbeURL:    probe,
			Role:        RoleSearchAndDownload,
			State:       StateChecking,
			StatusCode:  0,
			Latency:     0,
			LastChecked: time.Time{},
		})
	}

	// Known IPFS Gateways
	ipfsGateways := []string{
		"https://cloudflare-ipfs.com/ipfs/",
		"https://ipfs.io/ipfs/",
	}
	for _, raw := range ipfsGateways {
		if parsed, err := url.Parse(raw); err == nil {
			list = append(list, MirrorInfo{
				Host:        parsed.Host,
				ProbeURL:    raw,
				Role:        RoleIPFSGateway,
				State:       StateChecking,
				StatusCode:  0,
				Latency:     0,
				LastChecked: time.Time{},
			})
		}
	}

	m.mirrors = list
}

// SetCustomMirrors allows injecting mock or alternative mirrors (used by tests).
func (m *MirrorHealthManager) SetCustomMirrors(mirrors []MirrorInfo) {
	m.mu.Lock()
	m.mirrors = make([]MirrorInfo, len(mirrors))
	copy(m.mirrors, mirrors)
	m.mu.Unlock()
	m.notifySubscribers()
}

// GetMirrors returns a thread-safe copy of all mirror health records.
func (m *MirrorHealthManager) GetMirrors() []MirrorInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]MirrorInfo, len(m.mirrors))
	copy(out, m.mirrors)
	return out
}

// GetActiveSearchCount returns active search mirrors vs total search mirrors.
func (m *MirrorHealthManager) GetActiveSearchCount() (active int, total int) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, item := range m.mirrors {
		if item.Role == RoleSearchAndDownload {
			total++
			if item.State == StateOnline || item.State == StateSlow {
				active++
			}
		}
	}
	return active, total
}

// GetActiveAllCount returns active vs total across all mirrors and gateways.
func (m *MirrorHealthManager) GetActiveAllCount() (active int, total int) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	total = len(m.mirrors)
	for _, item := range m.mirrors {
		if item.State == StateOnline || item.State == StateSlow {
			active++
		}
	}
	return active, total
}

// ProbeAll concurrently probes all mirrors and notifies subscribers of updated statuses.
func (m *MirrorHealthManager) ProbeAll(ctx context.Context) {
	m.mu.Lock()
	for i := range m.mirrors {
		m.mirrors[i].State = StateChecking
	}
	mirrorsCopy := make([]MirrorInfo, len(m.mirrors))
	copy(mirrorsCopy, m.mirrors)
	m.mu.Unlock()
	m.notifySubscribers()

	var wg sync.WaitGroup
	results := make([]MirrorInfo, len(mirrorsCopy))

	for i, item := range mirrorsCopy {
		wg.Add(1)
		go func(idx int, info MirrorInfo) {
			defer wg.Done()
			results[idx] = m.probeSingle(ctx, info)
		}(i, item)
	}

	wg.Wait()

	m.mu.Lock()
	m.mirrors = results
	m.mu.Unlock()

	m.notifySubscribers()
}

// ProbeSingle probes a specific mirror host.
func (m *MirrorHealthManager) ProbeSingle(ctx context.Context, host string) MirrorInfo {
	m.mu.RLock()
	var target MirrorInfo
	found := false
	for _, item := range m.mirrors {
		if item.Host == host {
			target = item
			found = true
			break
		}
	}
	m.mu.RUnlock()

	if !found {
		return MirrorInfo{Host: host, State: StateOffline, ErrorMsg: "unknown mirror"}
	}

	res := m.probeSingle(ctx, target)

	m.mu.Lock()
	for i := range m.mirrors {
		if m.mirrors[i].Host == host {
			m.mirrors[i] = res
			break
		}
	}
	m.mu.Unlock()

	m.notifySubscribers()
	return res
}

func (m *MirrorHealthManager) probeSingle(ctx context.Context, info MirrorInfo) MirrorInfo {
	start := time.Now()
	probeURL := info.ProbeURL

	req, err := http.NewRequestWithContext(ctx, http.MethodHead, probeURL, nil)
	if err != nil {
		req, _ = http.NewRequestWithContext(ctx, http.MethodGet, probeURL, nil)
	}
	if req != nil {
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	}

	res := info
	res.LastChecked = time.Now()

	resp, err := m.httpClient.Do(req)
	latency := time.Since(start)
	res.Latency = latency

	if err != nil {
		// Try fallback GET in case HEAD was rejected with 405 Method Not Allowed
		getReq, getErr := http.NewRequestWithContext(ctx, http.MethodGet, probeURL, nil)
		if getErr == nil {
			getReq.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
			if getResp, err2 := m.httpClient.Do(getReq); err2 == nil {
				defer getResp.Body.Close()
				res.StatusCode = getResp.StatusCode
				if getResp.StatusCode >= 200 && getResp.StatusCode < 400 {
					res.State = StateOnline
					if latency > 1500*time.Millisecond {
						res.State = StateSlow
					}
					res.ErrorMsg = ""
					return res
				}
			}
		}

		res.State = StateOffline
		res.StatusCode = 0
		res.ErrorMsg = err.Error()
		return res
	}
	defer resp.Body.Close()

	res.StatusCode = resp.StatusCode
	if resp.StatusCode >= 200 && resp.StatusCode < 400 {
		res.State = StateOnline
		if latency > 1500*time.Millisecond {
			res.State = StateSlow
		}
		res.ErrorMsg = ""
	} else {
		res.State = StateOffline
		res.ErrorMsg = fmt.Sprintf("HTTP %d %s", resp.StatusCode, http.StatusText(resp.StatusCode))
	}

	return res
}

// Subscribe registers a listener invoked whenever mirror statuses are updated.
func (m *MirrorHealthManager) Subscribe(fn func(activeSearch, totalSearch, activeAll, totalAll int)) func() {
	m.mu.Lock()
	m.subscribers = append(m.subscribers, fn)
	m.mu.Unlock()

	// Immediately trigger with current stats
	activeSearch, totalSearch := m.GetActiveSearchCount()
	activeAll, totalAll := m.GetActiveAllCount()
	fn(activeSearch, totalSearch, activeAll, totalAll)

	// Unsubscribe closure
	return func() {
		m.mu.Lock()
		defer m.mu.Unlock()
		for i, sub := range m.subscribers {
			// Compare function pointers
			if fmt.Sprintf("%p", sub) == fmt.Sprintf("%p", fn) {
				m.subscribers = append(m.subscribers[:i], m.subscribers[i+1:]...)
				break
			}
		}
	}
}

func (m *MirrorHealthManager) notifySubscribers() {
	m.mu.RLock()
	subs := make([]func(int, int, int, int), len(m.subscribers))
	copy(subs, m.subscribers)
	m.mu.RUnlock()

	activeSearch, totalSearch := m.GetActiveSearchCount()
	activeAll, totalAll := m.GetActiveAllCount()

	for _, sub := range subs {
		if sub != nil {
			go sub(activeSearch, totalSearch, activeAll, totalAll)
		}
	}
}
