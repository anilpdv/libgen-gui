// Copyright © 2019 Antoine Chiny <antoine.chiny@inria.fr>
// Copyright © 2019 Ryan Ciehanski <ryan@ciehanski.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package libgen

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/fatih/color"
)

// Book is the struct of resources on Library Genesis.
type Book struct {
	ID          string
	Title       string
	Author      string
	Filesize    string
	Extension   string
	Md5         string
	Year        string
	Language    string
	Pages       string
	Publisher   string
	Edition     string
	CoverURL    string
	DownloadURL string
	PageURL     string
}

// SearchOptions are the optional parameters available for the Search
// function.
type SearchOptions struct {
	Query         string
	SearchMirror  url.URL
	Results       int
	Print         bool
	RequireAuthor bool
	Extension     []string
	Year          int
	Publisher     string
	Language      string
	SortBy        string
	SortASC       bool
	Page          int             // 1-based page number (defaults to 1)
	MaxPages      int             // Maximum pages to scan during auto-paging (defaults to 5)
	AutoPaging    bool            // Automatically query subsequent pages until Results matching books are gathered
	Context       context.Context // Optional context for timeouts and cancellation
	WorkingMirror *url.URL        // Optional pointer to receive the mirror URL that succeeded
}

// GetDetailsOptions are the optional parameters available for the GetDetails
// function.
type GetDetailsOptions struct {
	Hashes        []string
	SearchMirror  url.URL
	Print         bool
	RequireAuthor bool
	Extension     []string
	Year          int
	Publisher     string
	Language      string
	SortBy        string
}

func isNetworkOrDNSError(err error) bool {
	if err == nil {
		return false
	}
	s := strings.ToLower(err.Error())
	return strings.Contains(s, "no such host") ||
		strings.Contains(s, "i/o timeout") ||
		strings.Contains(s, "timeout") ||
		strings.Contains(s, "connection refused") ||
		strings.Contains(s, "network is unreachable") ||
		strings.Contains(s, "no route to host") ||
		strings.Contains(s, "deadline exceeded") ||
		strings.Contains(s, "reset by peer") ||
		strings.Contains(s, "lookup ") ||
		strings.Contains(s, "broken pipe") ||
		strings.Contains(s, "handshake failure") ||
		strings.Contains(s, "unable to reach mirror")
}

// Search sends a query to the search or index page hosted by LibGen mirrors,
// parses results from the new engine table (if available) or classic HTML table,
// filters and sorts the books, and returns them.
func Search(options *SearchOptions) ([]*Book, error) {
	var mirrorsToTry []url.URL
	if options.SearchMirror.Host != "" {
		mirrorsToTry = append(mirrorsToTry, options.SearchMirror)
	}
	for _, m := range SearchMirrors {
		if options.SearchMirror.Host == "" || m.Host != options.SearchMirror.Host {
			mirrorsToTry = append(mirrorsToTry, m)
		}
	}

	desiredResults := options.Results
	if desiredResults <= 0 {
		desiredResults = 10
	}
	startPage := options.Page
	if startPage <= 0 {
		startPage = 1
	}

	// Normalize and clean extensions
	var cleanExtensions []string
	for _, ext := range options.Extension {
		ext = strings.TrimSpace(strings.ToLower(strings.TrimPrefix(ext, ".")))
		if ext != "" {
			cleanExtensions = append(cleanExtensions, ext)
		}
	}
	options.Extension = cleanExtensions

	hasFilters := len(options.Extension) > 0 || options.Year != 0 || options.Publisher != "" || options.Language != "" || options.RequireAuthor
	maxPages := options.MaxPages
	if maxPages <= 0 {
		if hasFilters || options.AutoPaging {
			maxPages = 5
		} else {
			maxPages = 1
		}
	}

	ctx := options.Context
	if ctx == nil {
		ctx = context.Background()
	}

	var lastErr error
	var atLeastOneSuccess bool
	for _, mirror := range mirrorsToTry {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		var allBooks []*Book
		seenMD5 := make(map[string]bool)

		for pageNum := startPage; pageNum < startPage+maxPages; pageNum++ {
			if ctx.Err() != nil {
				return nil, ctx.Err()
			}

			searchURL := mirror
			q := searchURL.Query()
			q.Set("req", options.Query)
			q.Set("lg_topic", "libgen")
			q.Set("open", "0")
			q.Set("view", "simple")
			q.Set("res", "100")
			q.Set("phrase", "1")
			q.Set("column", "def")
			if pageNum > 1 {
				q.Set("page", fmt.Sprint(pageNum))
			}
			switch options.SortBy {
			case "id":
				q.Set("sort", "id")
				setSortASC(q, options.SortASC)
			case "title":
				q.Set("sort", "title")
				setSortASC(q, options.SortASC)
			case "author":
				q.Set("sort", "author")
				setSortASC(q, options.SortASC)
			case "pub":
				q.Set("sort", "publisher")
				setSortASC(q, options.SortASC)
			case "ext":
				q.Set("sort", "extension")
				setSortASC(q, options.SortASC)
			case "year":
				q.Set("sort", "year")
				setSortASC(q, options.SortASC)
			case "size":
				q.Set("sort", "filesize")
				setSortASC(q, options.SortASC)
			case "lang":
				q.Set("sort", "language")
				setSortASC(q, options.SortASC)
			}
			searchURL.RawQuery = q.Encode()

			referer := fmt.Sprintf("https://%s/index.php", mirror.Host)
			b, err := GetBodyWithRefererContext(ctx, searchURL.String(), referer)
			if err != nil {
				lastErr = err
				if pageNum == startPage && !isNetworkOrDNSError(err) {
					// If index.php failed with a non-network error on the first page, try search.php (or vice-versa)
					altURL := searchURL
					if strings.HasSuffix(altURL.Path, "index.php") {
						altURL.Path = "search.php"
					} else {
						altURL.Path = "index.php"
					}
					b, err = GetBodyWithRefererContext(ctx, altURL.String(), referer)
					if err != nil {
						lastErr = err
						break
					}
					searchURL = altURL
				} else {
					break
				}
			}

			atLeastOneSuccess = true

			// 1. Try parsing new engine table
			if strings.Contains(string(b), "tablelibgen") || strings.Contains(string(b), "table-striped") {
				parsedBooks := parseNewEngineTable(b, mirror.Host, 0)
				if len(parsedBooks) == 0 {
					break
				}
				filtered := filterBooks(parsedBooks, options)
				for _, bk := range filtered {
					if !seenMD5[bk.Md5] {
						seenMD5[bk.Md5] = true
						allBooks = append(allBooks, bk)
					}
				}
				if len(allBooks) >= desiredResults || (!hasFilters && !options.AutoPaging) {
					break
				}
				continue
			}

			// 2. Fallback to classic engine
			hashes := parseHashes(b, 100)
			if len(hashes) == 0 {
				break
			}
			books, err := GetDetails(&GetDetailsOptions{
				Hashes:        hashes,
				SearchMirror:  searchURL,
				Print:         false,
				RequireAuthor: options.RequireAuthor,
				Extension:     options.Extension,
				Year:          options.Year,
				Publisher:     options.Publisher,
				Language:      options.Language,
				SortBy:        options.SortBy,
			})
			if err == nil {
				for _, bk := range books {
					if !seenMD5[bk.Md5] {
						seenMD5[bk.Md5] = true
						allBooks = append(allBooks, bk)
					}
				}
			} else {
				lastErr = err
			}
			if len(allBooks) >= desiredResults || (!hasFilters && !options.AutoPaging) {
				break
			}
		}

		if len(allBooks) > 0 {
			sorted := sortBooks(allBooks, options.SortBy, options.SortASC)
			if len(sorted) > desiredResults {
				sorted = sorted[:desiredResults]
			}
			if options.Print {
				for _, book := range sorted {
					_ = printDetails(book)
				}
			}
			if options.WorkingMirror != nil {
				*options.WorkingMirror = mirror
			}
			return sorted, nil
		}

		if atLeastOneSuccess {
			// Mirror responded successfully, but had 0 books matching criteria.
			// Do not fall through to probe other mirrors.
			if options.WorkingMirror != nil {
				*options.WorkingMirror = mirror
			}
			return nil, errors.New("no books found matching criteria")
		}
	}

	if lastErr != nil {
		return nil, lastErr
	}
	return nil, errors.New("no books found matching criteria")
}

// GetDetails retrieves more details about a specific piece of media
// based off of its unique hash/id.
func GetDetails(options *GetDetailsOptions) ([]*Book, error) {
	var books []*Book
	var lastErr error

	for _, hash := range options.Hashes {
		hash = strings.TrimSpace(hash)
		if hash == "" {
			continue
		}

		book, err := getBookDetailsForHash(hash, options.SearchMirror)
		if err != nil {
			lastErr = err
			continue
		}

		// Apply filters
		if options.RequireAuthor && (book.Author == "" || book.Author == "N/A") {
			continue
		}
		if len(options.Extension) > 0 {
			validExtension := false
			for _, ext := range options.Extension {
				ext = strings.TrimSpace(ext)
				if ext == "" || strings.EqualFold(ext, book.Extension) {
					validExtension = true
					break
				}
			}
			if !validExtension {
				continue
			}
		}
		if options.Year != 0 {
			y, err := strconv.Atoi(book.Year)
			if err != nil || options.Year != y {
				continue
			}
		}
		if options.SortBy == "year" {
			if book.Year == "" || book.Year == "0" {
				continue
			}
		}
		if options.Publisher != "" {
			if !strings.Contains(strings.ToLower(book.Publisher), strings.ToLower(options.Publisher)) {
				continue
			}
		}
		if options.Language != "" {
			if !strings.EqualFold(book.Language, options.Language) {
				continue
			}
		}
		if options.Print {
			if err := printDetails(book); err != nil {
				return nil, err
			}
		}

		books = append(books, book)
	}

	if len(books) == 0 && len(options.Hashes) > 0 {
		if lastErr != nil {
			return nil, lastErr
		}
		return nil, errors.New("no books found matching criteria")
	}

	return books, nil
}

func getBookDetailsForHash(hash string, mirror url.URL) (*Book, error) {
	candidates := []url.URL{mirror}
	for _, m := range SearchMirrors {
		if m.Host != mirror.Host {
			candidates = append(candidates, m)
		}
	}

	var lastErr error
	for _, cand := range candidates {
		// 1. Try classic json.php?ids=<hash>&fields=...
		jsonURL := cand
		jsonURL.Path = "/json.php"
		q := url.Values{}
		q.Set("ids", hash)
		q.Set("fields", JSONQuery)
		jsonURL.RawQuery = q.Encode()

		b, err := GetBody(jsonURL.String())
		if err == nil {
			if book, parseErr := parseResponse(b); parseErr == nil && book.ID != "" {
				book.PageURL = fmt.Sprintf("https://%s/ads.php?md5=%s", cand.Host, strings.ToLower(hash))
				return book, nil
			}
		}

		// 2. Try new engine json.php?object=f&md5=<hash> + ads.php
		book, errNew := getNewEngineDetails(cand, hash)
		if errNew == nil && (book.Title != "" || book.Filesize != "") {
			return book, nil
		}
		if err != nil {
			lastErr = err
		} else if errNew != nil {
			lastErr = errNew
		}
	}

	if lastErr != nil {
		return nil, lastErr
	}
	return nil, fmt.Errorf("unable to retrieve details for hash %s", hash)
}

func getNewEngineDetails(mirror url.URL, hash string) (*Book, error) {
	cleanHash := strings.ToLower(strings.TrimSpace(hash))
	book := &Book{
		Md5:     cleanHash,
		PageURL: fmt.Sprintf("https://%s/ads.php?md5=%s", mirror.Host, cleanHash),
	}

	// Fetch json.php?object=f&md5=<cleanHash>
	jsonURL := fmt.Sprintf("https://%s/json.php?object=f&md5=%s", mirror.Host, cleanHash)
	bJSON, errJSON := GetBody(jsonURL)
	if errJSON == nil {
		var rawMap map[string]map[string]interface{}
		if err := json.Unmarshal(bJSON, &rawMap); err == nil {
			for fileID, item := range rawMap {
				getString := func(key string) string {
					if val, ok := item[key]; ok && val != nil {
						return fmt.Sprint(val)
					}
					return ""
				}
				book.ID = getString("libgen_id")
				if book.ID == "" {
					book.ID = fileID
				}
				book.Filesize = getString("filesize")
				book.Extension = getString("extension")
				book.Pages = getString("pages")
				if book.Title == "" {
					book.Title = getString("locator")
				}
				break
			}
		}
	}

	// Fetch ads.php?md5=<cleanHash> with Referer: https://<mirror.Host>/index.php
	adsURL := book.PageURL
	referer := fmt.Sprintf("https://%s/index.php", mirror.Host)
	bAds, errAds := GetBodyWithReferer(adsURL, referer)
	if errAds == nil {
		adsHTML := string(bAds)
		extractBib := func(field string) string {
			re := regexp.MustCompile(`(?i)` + field + `\s*=\s*\{([^}]+)\}`)
			if m := re.FindStringSubmatch(adsHTML); len(m) > 1 {
				return stripHTML(m[1])
			}
			return ""
		}
		if t := extractBib("title"); t != "" {
			book.Title = t
		}
		if a := extractBib("author"); a != "" {
			book.Author = a
		} else {
			reAuthor := regexp.MustCompile(`(?i)Author\(s\):\s*([^<]+)`)
			if m := reAuthor.FindStringSubmatch(adsHTML); len(m) > 1 {
				book.Author = stripHTML(m[1])
			}
		}
		if p := extractBib("publisher"); p != "" {
			book.Publisher = p
		}
		if y := extractBib("year"); y != "" {
			book.Year = y
		}
		reCover := regexp.MustCompile(`(?i)<img[^>]+src=['"](/covers/[^'"]+)['"]`)
		if m := reCover.FindStringSubmatch(adsHTML); len(m) > 1 {
			book.CoverURL = fmt.Sprintf("https://%s%s", mirror.Host, m[1])
		}
	}

	if book.Title == "" && book.Filesize == "" {
		return nil, errors.New("failed to retrieve book metadata from new engine")
	}

	return book, nil
}

func parseNewEngineTable(response []byte, mirrorHost string, maxResults int) []*Book {
	s := string(response)
	tableStart := strings.Index(s, "id=\"tablelibgen\"")
	if tableStart == -1 {
		tableStart = strings.Index(s, "table-striped")
	}
	if tableStart == -1 {
		return nil
	}
	tTagStart := strings.LastIndex(s[:tableStart], "<table")
	if tTagStart == -1 {
		tTagStart = tableStart
	}
	tTagEnd := strings.Index(s[tTagStart:], "</table>")
	if tTagEnd == -1 {
		return nil
	}
	tableHTML := s[tTagStart : tTagStart+tTagEnd+8]

	reTR := regexp.MustCompile(`(?s)<tr[^>]*>(.*?)</tr>`)
	reTD := regexp.MustCompile(`(?s)<td[^>]*>(.*?)</td>`)
	reMD5 := regexp.MustCompile(`(?i)(?:md5=|\/ads\.php\?md5=|\/book\/)([a-f0-9]{32})`)
	reLibgenID := regexp.MustCompile(`(?i)\bl\s*(\d+)`)
	reAltID := regexp.MustCompile(`(?i)id:\s*(\d+)`)
	reFileID := regexp.MustCompile(`(?i)id=(\d+)`)

	trMatches := reTR.FindAllStringSubmatch(tableHTML, -1)
	var books []*Book

	for _, tr := range trMatches {
		if len(tr) < 2 {
			continue
		}
		rowContent := tr[1]
		if !strings.Contains(rowContent, "<td") {
			continue
		}
		tdMatches := reTD.FindAllStringSubmatch(rowContent, -1)
		if len(tdMatches) < 8 {
			continue
		}

		md5Match := reMD5.FindStringSubmatch(rowContent)
		if len(md5Match) < 2 {
			continue
		}
		md5 := strings.ToLower(md5Match[1])

		col0 := tdMatches[0][1]
		var title string
		if idx := strings.Index(col0, "edition.php"); idx != -1 {
			closeB := strings.Index(col0[idx:], ">")
			if closeB != -1 {
				start := idx + closeB + 1
				endA := strings.Index(col0[start:], "</a>")
				if endA != -1 {
					title = stripHTML(col0[start : start+endA])
				}
			}
		}
		if title == "" {
			reA := regexp.MustCompile(`(?s)<a[^>]*>(.*?)</a>`)
			if m := reA.FindStringSubmatch(col0); len(m) > 1 {
				title = stripHTML(m[1])
			} else {
				title = stripHTML(col0)
			}
		}

		var bookID string
		if m := reLibgenID.FindStringSubmatch(col0); len(m) > 1 {
			bookID = m[1]
		} else if m := reAltID.FindStringSubmatch(col0); len(m) > 1 {
			bookID = m[1]
		} else if len(tdMatches) > 6 {
			if m := reFileID.FindStringSubmatch(tdMatches[6][1]); len(m) > 1 {
				bookID = m[1]
			}
		}

		author := stripHTML(tdMatches[1][1])
		publisher := stripHTML(tdMatches[2][1])
		year := stripHTML(tdMatches[3][1])
		language := stripHTML(tdMatches[4][1])
		pages := stripHTML(tdMatches[5][1])
		rawFilesize := stripHTML(tdMatches[6][1])
		extension := strings.ToLower(stripHTML(tdMatches[7][1]))

		filesize := parseFilesizeBytes(rawFilesize)

		book := &Book{
			ID:        bookID,
			Title:     title,
			Author:    author,
			Publisher: publisher,
			Year:      year,
			Language:  language,
			Pages:     pages,
			Filesize:  filesize,
			Extension: extension,
			Md5:       md5,
			PageURL:   fmt.Sprintf("https://%s/ads.php?md5=%s", mirrorHost, md5),
		}

		books = append(books, book)
		if maxResults > 0 && len(books) >= maxResults*2 {
			break
		}
	}

	return books
}

func filterBooks(books []*Book, options *SearchOptions) []*Book {
	var filtered []*Book
	for _, book := range books {
		if options.RequireAuthor && (book.Author == "" || book.Author == "N/A") {
			continue
		}
		if len(options.Extension) > 0 {
			validExtension := false
			bookExt := strings.ToLower(strings.TrimSpace(strings.TrimPrefix(book.Extension, ".")))
			for _, ext := range options.Extension {
				ext = strings.ToLower(strings.TrimSpace(strings.TrimPrefix(ext, ".")))
				if ext == "" || ext == bookExt {
					validExtension = true
					break
				}
			}
			if !validExtension {
				continue
			}
		}
		if options.Year != 0 {
			y, err := strconv.Atoi(book.Year)
			if err != nil || options.Year != y {
				continue
			}
		}
		if options.SortBy == "year" {
			if book.Year == "" || book.Year == "0" {
				continue
			}
		}
		if options.Publisher != "" {
			if !strings.Contains(strings.ToLower(book.Publisher), strings.ToLower(options.Publisher)) {
				continue
			}
		}
		if options.Language != "" {
			if !strings.EqualFold(book.Language, options.Language) {
				continue
			}
		}
		filtered = append(filtered, book)
	}
	return filtered
}

func sortBooks(books []*Book, sortBy string, sortASC bool) []*Book {
	if sortBy == "" {
		return books
	}
	sorted := make([]*Book, len(books))
	copy(sorted, books)

	sort.SliceStable(sorted, func(i, j int) bool {
		var less bool
		switch sortBy {
		case "id":
			idI, _ := strconv.Atoi(sorted[i].ID)
			idJ, _ := strconv.Atoi(sorted[j].ID)
			less = idI < idJ
		case "title":
			less = strings.ToLower(sorted[i].Title) < strings.ToLower(sorted[j].Title)
		case "author":
			less = strings.ToLower(sorted[i].Author) < strings.ToLower(sorted[j].Author)
		case "pub":
			less = strings.ToLower(sorted[i].Publisher) < strings.ToLower(sorted[j].Publisher)
		case "ext":
			less = strings.ToLower(sorted[i].Extension) < strings.ToLower(sorted[j].Extension)
		case "year":
			yI, _ := strconv.Atoi(sorted[i].Year)
			yJ, _ := strconv.Atoi(sorted[j].Year)
			less = yI < yJ
		case "size":
			sI, _ := strconv.Atoi(sorted[i].Filesize)
			sJ, _ := strconv.Atoi(sorted[j].Filesize)
			less = sI < sJ
		case "lang":
			less = strings.ToLower(sorted[i].Language) < strings.ToLower(sorted[j].Language)
		default:
			return false
		}
		if !sortASC {
			return !less
		}
		return less
	})
	return sorted
}

var reHTMLTag = regexp.MustCompile(`(?s)<[^>]*>`)

func stripHTML(s string) string {
	s = reHTMLTag.ReplaceAllString(s, "")
	s = html.UnescapeString(s)
	return strings.TrimSpace(s)
}

func parseFilesizeBytes(raw string) string {
	raw = strings.TrimSpace(raw)
	noSpace := strings.ReplaceAll(raw, " ", "")
	b, err := humanize.ParseBytes(noSpace)
	if err == nil && b > 0 {
		return fmt.Sprintf("%d", b)
	}
	if _, err := strconv.Atoi(raw); err == nil {
		return raw
	}
	return raw
}

// DefaultTransport allows injecting custom transports (e.g. for testing or custom proxies).
var DefaultTransport http.RoundTripper

func getHTTPClient(timeout time.Duration) *http.Client {
	c := &http.Client{
		Timeout: timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 10 {
				return errors.New("stopped after 10 redirects")
			}
			req.Header.Set("User-Agent", UserAgent)
			if len(via) > 0 {
				req.Header.Set("Referer", via[len(via)-1].URL.String())
			}
			return nil
		},
	}
	if DefaultTransport != nil {
		c.Transport = DefaultTransport
	} else {
		c.Transport = &http.Transport{
			Proxy:             http.ProxyFromEnvironment,
			TLSClientConfig:   &tls.Config{InsecureSkipVerify: true},
			TLSNextProto:      make(map[string]func(authority string, c *tls.Conn) http.RoundTripper),
			ForceAttemptHTTP2: false,
		}
	}
	return c
}

// CheckMirror returns the HTTP status code of the DownloadURL provided.
func CheckMirror(url url.URL) int {
	return CheckMirrorWithTimeout(url, 5*time.Second)
}

// CheckMirrorWithTimeout checks mirror reachability within the specified timeout.
func CheckMirrorWithTimeout(url url.URL, timeout time.Duration) int {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return CheckMirrorWithContext(ctx, url)
}

// CheckMirrorWithContext checks mirror reachability within the specified context.
func CheckMirrorWithContext(ctx context.Context, url url.URL) int {
	client := getHTTPClient(0)

	req, err := http.NewRequestWithContext(ctx, "GET", url.String(), nil)
	if err != nil {
		return http.StatusBadGateway
	}
	req.Header.Set("User-Agent", UserAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")

	r, err := client.Do(req)
	if err != nil {
		return http.StatusBadGateway
	}
	defer r.Body.Close()

	if r.StatusCode >= 200 && r.StatusCode < 400 {
		return http.StatusOK
	}
	return r.StatusCode
}

// ParseMirrorURL normalizes and validates a mirror URL provided by a flag or env var.
// It auto-prefixes https:// if missing, ensures host is present, and sets defaultPath if empty.
func ParseMirrorURL(raw string, defaultPath string) (url.URL, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return url.URL{}, errors.New("empty mirror URL")
	}
	if !strings.HasPrefix(raw, "http://") && !strings.HasPrefix(raw, "https://") {
		raw = "https://" + raw
	}
	parsed, err := url.Parse(raw)
	if err != nil {
		return url.URL{}, err
	}
	if parsed.Host == "" {
		return url.URL{}, fmt.Errorf("invalid mirror host in %q", raw)
	}
	if parsed.Scheme == "" {
		parsed.Scheme = "https"
	}
	if parsed.Path == "" || parsed.Path == "/" {
		parsed.Path = defaultPath
	}
	return *parsed, nil
}

// GetWorkingMirror selects a working mirror from the []url.URL provided.
// It probes candidate mirrors concurrently with a fast timeout and falls back to the first mirror.
func GetWorkingMirror(urls []url.URL) url.URL {
	if len(urls) == 0 {
		return url.URL{Scheme: "https", Host: "libgen.li", Path: "index.php"}
	}

	// Check environment variable override
	if envMirror := os.Getenv("LIBGEN_MIRROR"); envMirror != "" {
		if parsed, err := ParseMirrorURL(envMirror, "index.php"); err == nil {
			return parsed
		}
	}
	if envMirror := os.Getenv("LIBGEN_SEARCH_MIRROR"); envMirror != "" {
		if parsed, err := ParseMirrorURL(envMirror, "index.php"); err == nil {
			return parsed
		}
	}

	// Shuffle candidate mirrors
	shuffled := make([]url.URL, len(urls))
	copy(shuffled, urls)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	r.Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})

	resultChan := make(chan url.URL, len(shuffled))
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var activeProbes int32 = int32(len(shuffled))
	for _, m := range shuffled {
		go func(target url.URL) {
			if CheckMirrorWithContext(ctx, target) == http.StatusOK {
				select {
				case resultChan <- target:
					cancel() // early abort remaining in-flight probes
				default:
				}
			}
			if atomic.AddInt32(&activeProbes, -1) == 0 {
				cancel()
			}
		}(m)
	}

	select {
	case working := <-resultChan:
		return working
	case <-ctx.Done():
		select {
		case working := <-resultChan:
			return working
		default:
			return urls[0]
		}
	}
}

// ParseDbdumps takes in a HTTP response and scans it for
// any string that matches a filepath and returns all unique results.
func ParseDbdumps(response []byte) []string {
	re := regexp.MustCompile(dbdumpReg)
	matches := re.FindAllStringSubmatch(string(response), -1)
	var dbdumps []string
	seen := make(map[string]bool)

	for _, m := range matches {
		if len(m) > 2 {
			filename := m[2]
			if !seen[filename] {
				seen[filename] = true
				dbdumps = append(dbdumps, filename)
			}
		}
	}

	return dbdumps
}

// GetBody performs an HTTP GET request with standard browser headers and timeout.
func GetBody(baseURL string) ([]byte, error) {
	return GetBodyWithRefererContext(context.Background(), baseURL, "")
}

// GetBodyWithReferer performs an HTTP GET request with standard browser headers, referer, and timeout.
func GetBodyWithReferer(baseURL string, referer string) ([]byte, error) {
	return GetBodyWithRefererContext(context.Background(), baseURL, referer)
}

// GetBodyWithRefererContext performs an HTTP GET request with standard browser headers, referer, and context.
func GetBodyWithRefererContext(ctx context.Context, baseURL string, referer string) ([]byte, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	client := getHTTPClient(HTTPClientTimeout)

	req, err := http.NewRequestWithContext(ctx, "GET", baseURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", UserAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	if referer != "" {
		req.Header.Set("Referer", referer)
	}

	r, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	if r.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unable to reach mirror %v: %v", baseURL, r.StatusCode)
	}

	b, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func getBody(baseURL string) ([]byte, error) {
	return GetBody(baseURL)
}

// parseHashes takes in a HTTP response and scans it for
// an MD5 hash and then returns the found hashes in uppercase.
func parseHashes(response []byte, results int) []string {
	var hashes []string
	seen := make(map[string]bool)

	// Pattern 1: Support both single and double quotes, relative and absolute book/index.php links
	reHref := regexp.MustCompile(`(?i)(?:href=['"][^'"]*?(?:book/index\.php\?md5=|/main/|item/detail/id/|\?md5=)|md5=)([A-Za-z0-9]{32})`)
	matches := reHref.FindAllSubmatch(response, -1)
	for _, m := range matches {
		if len(m) > 1 {
			hash := strings.ToUpper(string(m[1]))
			if len(hash) == 32 && !seen[hash] {
				seen[hash] = true
				hashes = append(hashes, hash)
				if len(hashes) >= results {
					return hashes
				}
			}
		}
	}

	// Pattern 2: Legacy fallback matching original SearchHref with single quotes
	reLegacy := regexp.MustCompile(SearchHref)
	matchesLegacy := reLegacy.FindAllString(string(response), -1)
	for _, m := range matchesLegacy {
		re := regexp.MustCompile(SearchMD5)
		hash := strings.ToUpper(re.FindString(m))
		if len(hash) == 32 && !seen[hash] {
			seen[hash] = true
			hashes = append(hashes, hash)
			if len(hashes) >= results {
				return hashes
			}
		}
	}

	// Pattern 3: Any standalone 32-hex character string in href
	reAny := regexp.MustCompile(`(?i)href=['"][^'"]*?([A-Za-z0-9]{32})[^'"]*?['"]`)
	matchesAny := reAny.FindAllSubmatch(response, -1)
	for _, m := range matchesAny {
		if len(m) > 1 {
			hash := strings.ToUpper(string(m[1]))
			if len(hash) == 32 && !seen[hash] {
				seen[hash] = true
				hashes = append(hashes, hash)
				if len(hashes) >= results {
					return hashes
				}
			}
		}
	}

	return hashes
}

// parseResponse takes in a slice of bytes and formats it
// returns a Book object from the slice of bytes.
func parseResponse(response []byte) (*Book, error) {
	// Check if the response contains an error object: {"error": "..."}
	var errCheck map[string]interface{}
	if err := json.Unmarshal(response, &errCheck); err == nil {
		if errMsg, ok := errCheck["error"]; ok {
			return nil, fmt.Errorf("mirror returned error: %v", errMsg)
		}
	}

	var book Book
	var rawResp []map[string]interface{}

	if err := json.Unmarshal(response, &rawResp); err != nil {
		// Try unmarshaling into map of objects: {"93485370": {"md5": ...}}
		var mapResp map[string]map[string]interface{}
		if errMap := json.Unmarshal(response, &mapResp); errMap == nil && len(mapResp) > 0 {
			for k, v := range mapResp {
				if v != nil {
					v["id"] = k
					rawResp = append(rawResp, v)
					break
				}
			}
		} else {
			// Try unmarshaling into []map[string]string as fallback
			var stringResp []map[string]string
			if errStr := json.Unmarshal(response, &stringResp); errStr == nil && len(stringResp) > 0 {
				rawResp = make([]map[string]interface{}, len(stringResp))
				for i, m := range stringResp {
					rawResp[i] = make(map[string]interface{})
					for k, v := range m {
						rawResp[i][k] = v
					}
				}
			} else {
				return nil, err
			}
		}
	}

	if len(rawResp) == 0 {
		return nil, errors.New("empty response or unexpected JSON")
	}

	item := rawResp[0]
	getString := func(key string) string {
		if val, ok := item[key]; ok && val != nil {
			return fmt.Sprint(val)
		}
		return ""
	}

	book.ID = getString("id")
	book.Title = getString("title")
	book.Author = getString("author")
	book.Filesize = getString("filesize")
	book.Extension = getString("extension")
	book.Md5 = strings.ToLower(getString("md5"))
	book.Year = getString("year")
	book.Language = getString("language")
	book.Pages = getString("pages")
	book.Publisher = getString("publisher")
	book.Edition = getString("edition")
	book.CoverURL = getString("coverurl")

	return &book, nil
}

func printDetails(book *Book) error {
	var fsize string
	size, err := strconv.Atoi(book.Filesize)
	if err != nil {
		fsize = "N/A"
	} else {
		fsize = humanize.Bytes(uint64(size))
	}

	// Print separation lines
	fmt.Println(strings.Repeat("-", 80))

	// Print ID + Title
	fTitle := fmt.Sprintf("%5s %s", color.New(color.FgHiBlue).Sprint(book.ID), book.Title)
	fTitle = formatTitle(fTitle, TitleMaxLength)
	if runtime.GOOS == "windows" {
		_, err = fmt.Fprintf(color.Output, "%s\n    ++ ", fTitle)
		if err != nil {
			return err
		}
	} else {
		fmt.Printf("%s\n    ++ ", fTitle)
	}

	// Slice author name if it exceeds AuthorMaxLength
	var formatAuthor string
	if len(book.Author) > AuthorMaxLength {
		formatAuthor = book.Author[:AuthorMaxLength]
	} else if book.Author == "" {
		formatAuthor = "N/A"
	} else {
		formatAuthor = book.Author
	}

	err = prettify("author", formatAuthor, color.FgYellow, "-25")
	if err != nil {
		return err
	}
	err = prettify("year", book.Year, color.FgCyan, "4")
	if err != nil {
		return err
	}
	err = prettify("size", fsize, color.FgGreen, "6")
	if err != nil {
		return err
	}
	err = prettify("type", book.Extension, color.FgRed, "4")
	if err != nil {
		return err
	}
	fmt.Println()

	return nil
}

// formatTitle shortens the title of a Book down to
// the maximum allowed by TitleMaxLength.
func formatTitle(title string, maximumLength int) string {
	var fTitle []string
	var counter int

	if len(title) <= maximumLength {
		return title
	}

	title = strings.TrimSpace(title)
	for _, t := range strings.Split(title, " ") {
		counter += len(t)

		if counter > maximumLength {
			counter = 0
			t = t + "...\n"
		}
		fTitle = append(fTitle, t)
	}

	return strings.Join(fTitle, " ")
}

// prettify is a helper function that adds color and
// formats text returned to the user.
func prettify(key string, value string, col color.Attribute, align string) error {
	c := color.New(col).SprintFunc()
	a := fmt.Sprintf("%%%ss ", align)
	s := fmt.Sprintf("@%s "+a, c(key), value)
	if runtime.GOOS == "windows" {
		_, err := fmt.Fprintf(color.Output, a, s)
		if err != nil {
			return err
		}
	} else {
		fmt.Printf(a, s)
	}
	return nil
}

// RemoveQuotes is a helper function that removes the quotes from
// dbdumps page results.
func RemoveQuotes(s string) string {
	if s == "" {
		return ""
	}
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		return s[1 : len(s)-1]
	}
	return s
}

func setSortASC(q url.Values, sortASC bool) {
	if sortASC {
		q.Set("sortmode", "ASC")
	} else {
		q.Set("sortmode", "DESC")
	}
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

