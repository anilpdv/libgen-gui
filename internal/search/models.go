package search

import (
	"fmt"
	"strings"

	"github.com/dustin/go-humanize"
)

// Query encapsulates structured search filters and pagination parameters.
type Query struct {
	Text      string
	Language  string
	Format    string
	YearFrom  int
	YearTo    int
	Publisher string
	Author    string
	Page      int
	PageSize  int
	SortBy    string
	SortASC   bool
}

// Result represents a domain-level book search result.
type Result struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Authors     []string `json:"authors"`
	Publisher   string   `json:"publisher"`
	Year        int      `json:"year"`
	Language    string   `json:"language"`
	Format      string   `json:"format"`
	Size        int64    `json:"size"`
	Pages       string   `json:"pages"`
	Md5         string   `json:"md5"`
	SourceID    string   `json:"source_id"`
	DownloadURL string   `json:"download_url"`
	PageURL     string   `json:"page_url"`
	CoverURL    string   `json:"cover_url"`
}

// FormattedAuthors returns a readable comma-separated authors string.
func (r Result) FormattedAuthors() string {
	if len(r.Authors) == 0 {
		return "Unknown Author"
	}
	return strings.Join(r.Authors, ", ")
}

// FormattedSize returns a human-readable size string (e.g. "4.2 MB").
func (r Result) FormattedSize() string {
	if r.Size <= 0 {
		return "N/A"
	}
	return humanize.Bytes(uint64(r.Size))
}

// Details represents in-depth metadata for a specific work.
type Details struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Authors     []string `json:"authors"`
	Publisher   string   `json:"publisher"`
	Year        int      `json:"year"`
	Language    string   `json:"language"`
	Format      string   `json:"format"`
	Size        int64    `json:"size"`
	Pages       string   `json:"pages"`
	Edition     string   `json:"edition"`
	Description string   `json:"description"`
	CoverURL    string   `json:"cover_url"`
	DownloadURL string   `json:"download_url"`
	PageURL     string   `json:"page_url"`
	Md5         string   `json:"md5"`
}

// Key returns a unique deduplication and selection identifier for the search result.
func (r Result) Key() string {
	if r.Md5 != "" {
		return strings.ToLower(r.Md5)
	}
	if r.ID != "" {
		return r.ID
	}
	return fmt.Sprintf("%s_%s_%d_%s", r.Title, r.FormattedAuthors(), r.Year, r.Format)
}
