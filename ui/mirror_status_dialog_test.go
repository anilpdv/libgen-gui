package ui

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"fyne.io/fyne/v2/test"
)

func TestMirrorStatusFormatting(t *testing.T) {
	online := MirrorInfo{State: StateOnline, StatusCode: 200, Latency: 42 * time.Millisecond}
	if got := formatMirrorStatus(online); got != "● Online • HTTP 200" {
		t.Fatalf("unexpected online status: %q", got)
	}
	if got := formatMirrorLatency(online); got != "42 ms" {
		t.Fatalf("unexpected latency: %q", got)
	}

	checking := MirrorInfo{State: StateChecking}
	if got := formatMirrorLatency(checking); got != "—" {
		t.Fatalf("checking latency should be unavailable, got %q", got)
	}

	offline := MirrorInfo{State: StateOffline, ErrorMsg: "connection refused"}
	if got := formatMirrorStatus(offline); !strings.Contains(got, "connection refused") {
		t.Fatalf("offline status should include the error, got %q", got)
	}
}

func TestMirrorStatusDialogRendersAndRefreshes(t *testing.T) {
	fyneApp := test.NewApp()
	defer fyneApp.Quit()
	window := fyneApp.NewWindow("Mirror Status Test")
	defer window.Close()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	serverURL, err := url.Parse(server.URL)
	if err != nil {
		t.Fatal(err)
	}

	manager := NewMirrorHealthManager()
	manager.SetCustomMirrors([]MirrorInfo{{
		Host:     serverURL.Host,
		ProbeURL: server.URL,
		Role:     RoleSearchAndDownload,
		State:    StateChecking,
	}})

	view := NewMirrorStatusDialog(NewApp(window), manager)
	if view == nil || view.Dialog == nil {
		t.Fatal("expected a dedicated mirror status dialog")
	}
	defer view.Dialog.Hide()

	if len(view.MirrorList.Objects) == 0 {
		t.Fatal("expected configured mirrors to be rendered")
	}
	if !strings.Contains(view.SummaryLabel.Text, "0/1") {
		t.Fatalf("expected initial availability in summary, got %q", view.SummaryLabel.Text)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	manager.ProbeAll(ctx)
	view.render()

	if !strings.Contains(view.SummaryLabel.Text, "1/1") {
		t.Fatalf("expected refreshed availability in summary, got %q", view.SummaryLabel.Text)
	}
	if view.LastUpdated.Text == "Not checked yet" {
		t.Fatal("expected last checked time after refresh")
	}
}

func TestShowMirrorStatusDialogHandlesInvalidApp(t *testing.T) {
	ShowMirrorStatusDialog(nil)
	if got := NewMirrorStatusDialog(nil, nil); got != nil {
		t.Fatal("expected nil dialog for nil app")
	}
}
