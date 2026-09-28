package storage

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"../../evil.pdf", "evil.pdf"},
		{"Complex: Title / Subtitle * [2024]?.epub", "Complex - Title _ Subtitle _ [2024]_.epub"},
		{"\x00malicious\n\tfile.pdf", "malicious file.pdf"},
		{"", "download.bin"},
		{strings.Repeat("a", 300) + ".pdf", strings.Repeat("a", 176) + ".pdf"},
	}

	for _, tc := range tests {
		got := SanitizeFilename(tc.input)
		if got != tc.expected {
			t.Errorf("SanitizeFilename(%q) = %q, expected %q", tc.input, got, tc.expected)
		}
	}
}

func TestFileSystemStorage_CRUDAndAtomicCommit(t *testing.T) {
	tmpDir := t.TempDir()
	store, err := NewFileSystemStorage(tmpDir)
	if err != nil {
		t.Fatalf("unexpected NewFileSystemStorage error: %v", err)
	}

	ctx := context.Background()

	// 1. Create .part file and write data
	partName := "book1.pdf.part"
	finalName := "book1.pdf"

	w, err := store.Create(ctx, partName)
	if err != nil {
		t.Fatalf("store.Create failed: %v", err)
	}

	payload := "test book contents payload"
	if _, err := io.WriteString(w, payload); err != nil {
		t.Fatalf("failed writing to .part: %v", err)
	}
	_ = w.Close()

	// 2. Verify .part exists and size matches
	exists, err := store.Exists(ctx, partName)
	if err != nil || !exists {
		t.Fatalf("expected .part file to exist: %v", err)
	}

	sz, err := store.Size(ctx, partName)
	if err != nil || sz != int64(len(payload)) {
		t.Errorf("expected size %d, got %d", len(payload), sz)
	}

	// 3. Test Append Resume
	appW, existingBytes, err := store.OpenAppend(ctx, partName)
	if err != nil {
		t.Fatalf("store.OpenAppend failed: %v", err)
	}
	if existingBytes != int64(len(payload)) {
		t.Errorf("expected existing bytes %d, got %d", len(payload), existingBytes)
	}
	appPayload := " appended"
	if _, err := io.WriteString(appW, appPayload); err != nil {
		t.Fatalf("failed writing append: %v", err)
	}
	_ = appW.Close()

	// 4. Atomic Commit (.part -> final)
	if err := store.CommitPart(ctx, partName, finalName); err != nil {
		t.Fatalf("store.CommitPart failed: %v", err)
	}

	// 5. Verify final exists and .part is gone
	partExists, _ := store.Exists(ctx, partName)
	if partExists {
		t.Errorf("expected .part file to no longer exist after commit")
	}

	finalExists, _ := store.Exists(ctx, finalName)
	if !finalExists {
		t.Errorf("expected final file to exist after commit")
	}

	finalData, _ := os.ReadFile(filepath.Join(tmpDir, finalName))
	if string(finalData) != payload+appPayload {
		t.Errorf("final content mismatch: got %q", string(finalData))
	}

	// 6. Delete
	if err := store.Delete(ctx, finalName); err != nil {
		t.Fatalf("store.Delete failed: %v", err)
	}
	finalExistsAfterDelete, _ := store.Exists(ctx, finalName)
	if finalExistsAfterDelete {
		t.Errorf("expected final file to be deleted")
	}
}

func TestSAFStorage_Creation(t *testing.T) {
	tmpDir := t.TempDir()
	store, err := NewStorage(tmpDir, true)
	if err != nil {
		t.Fatalf("NewStorage failed: %v", err)
	}

	ctx := context.Background()
	w, err := store.Create(ctx, "saf_test.epub")
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	_ = w.Close()

	exists, err := store.Exists(ctx, "saf_test.epub")
	if err != nil || !exists {
		t.Errorf("expected saf_test.epub to exist")
	}
}
