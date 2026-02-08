package libgen

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"testing"
)

type mockTransport struct{}

func (m *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	path := req.URL.Path
	rawQuery := req.URL.RawQuery

	header := make(http.Header)
	var bodyReader io.ReadCloser
	statusCode := http.StatusOK

	switch {
	case strings.Contains(path, "json.php") || strings.Contains(rawQuery, "json.php"):
		header.Set("Content-Type", "application/json")
		reqObj := req.URL.Query().Get("object")
		reqMD5 := strings.ToUpper(strings.TrimSpace(req.URL.Query().Get("md5")))
		ids := req.URL.Query().Get("ids")

		if reqObj == "f" {
			// New engine JSON response: map of file_id -> metadata
			resMap := make(map[string]map[string]interface{})
			switch reqMD5 {
			case "2F2DBA2A621B693BB95601C16ED680F8":
				resMap["643"] = map[string]interface{}{
					"md5":        "2f2dba2a621b693bb95601c16ed680f8",
					"filesize":   "102400",
					"extension":  "gz",
					"pages":      "230",
					"locator":    "The Turing Test and the Frame Problem.gz",
					"libgen_id":  "643",
				}
			case "06E6135019C8F2F43158ABA9ABDC610E":
				resMap["3167"] = map[string]interface{}{
					"md5":        "06e6135019c8f2f43158aba9abdc610e",
					"filesize":   "204800",
					"extension":  "djvu",
					"pages":      "230",
					"locator":    "You failed your math test.djvu",
					"libgen_id":  "3167",
				}
			default:
				resMap["1"] = map[string]interface{}{
					"md5":        strings.ToLower(reqMD5),
					"filesize":   "1000",
					"extension":  "pdf",
					"pages":      "10",
					"locator":    "Default.pdf",
					"libgen_id":  "1",
				}
			}
			data, _ := json.Marshal(resMap)
			bodyReader = io.NopCloser(bytes.NewReader(data))
			break
		}

		var resp []map[string]interface{}
		idList := strings.Split(ids, ",")
		for _, id := range idList {
			id = strings.ToUpper(strings.TrimSpace(id))
			switch id {
			case "2F2DBA2A621B693BB95601C16ED680F8":
				resp = append(resp, map[string]interface{}{
					"id":        "643",
					"title":     "The Turing Test and the Frame Problem: AI's Mistaken Understanding of Intelligence",
					"author":    "Larry J. Crockett",
					"filesize":  "102400",
					"extension": "gz",
					"md5":       "2f2dba2a621b693bb95601c16ed680f8",
					"year":      "1994",
					"language":  "English",
					"pages":     "230",
					"publisher": "Ablex Publishing",
					"edition":   "",
					"coverurl":  "",
				})
			case "06E6135019C8F2F43158ABA9ABDC610E":
				resp = append(resp, map[string]interface{}{
					"id":        "3167",
					"title":     "You failed your math test, Comrade Einstein (about Soviet antisemitism)",
					"author":    "M. Shifman",
					"filesize":  "204800",
					"extension": "djvu",
					"md5":       "06e6135019c8f2f43158aba9abdc610e",
					"year":      "2005",
					"language":  "English",
					"pages":     "230",
					"publisher": "World Scientific Publishing",
					"edition":   "",
					"coverurl":  "",
				})
			case "553907CDF5F03AF78950561F42F1571A":
				resp = append(resp, map[string]interface{}{
					"id":        "5539",
					"title":     "Test Math Document",
					"author":    "Test Author",
					"filesize":  "307200",
					"extension": "pdf",
					"md5":       "553907cdf5f03af78950561f42f1571a",
					"year":      "2010",
					"language":  "English",
					"pages":     "150",
					"publisher": "Math Pub",
					"edition":   "",
					"coverurl":  "",
				})
			case "1794743BB21D72736FFE64D66DCA9F0E":
				resp = append(resp, map[string]interface{}{
					"id":        "9999",
					"title":     "LibGen Test Book",
					"author":    "Test Author",
					"filesize":  "409600",
					"extension": "epub",
					"md5":       "1794743bb21d72736ffe64d66dca9f0e",
					"year":      "2020",
					"language":  "English",
					"pages":     "300",
					"publisher": "Test Press",
					"edition":   "1st",
					"coverurl":  "",
				})
			default:
				resp = append(resp, map[string]interface{}{
					"id":        "1",
					"title":     "Default Book",
					"author":    "Default Author",
					"filesize":  "1000",
					"extension": "pdf",
					"md5":       strings.ToLower(id),
					"year":      "2020",
					"language":  "English",
					"pages":     "10",
					"publisher": "Default",
					"edition":   "",
					"coverurl":  "",
				})
			}
		}
		data, _ := json.Marshal(resp)
		bodyReader = io.NopCloser(bytes.NewReader(data))

	case strings.Contains(path, "index.php"):
		header.Set("Content-Type", "text/html; charset=utf-8")
		var html string
		if req.URL.Query().Get("page") == "2" {
			html = `<!DOCTYPE html>
<html>
<body>
<table class="table table-striped" id="tablelibgen">
<thead><tr><th>Title</th><th>Author</th><th>Publisher</th><th>Year</th><th>Lang</th><th>Pages</th><th>Size</th><th>Ext</th><th>Mirrors</th></tr></thead>
<tbody>
<tr>
  <td><a href="edition.php?id=999">Neuroscience and Brain Systems: Modern Discoveries</a><nobr><span class="badge badge-secondary">l 9999</span></nobr></td>
  <td>Mark F. Bear</td>
  <td>Academic Press</td>
  <td><nobr>2021</nobr></td>
  <td>English</td>
  <td>450</td>
  <td><nobr><a href="/file.php?id=9999">400 KB</a></nobr></td>
  <td>epub</td>
  <td><nobr><a href="/ads.php?md5=1794743bb21d72736ffe64d66dca9f0e">1</a></nobr></td>
</tr>
</tbody>
</table>
</body>
</html>`
		} else if req.URL.Query().Get("page") == "3" {
			html = `<!DOCTYPE html>
<html>
<body>
<table class="table table-striped" id="tablelibgen">
<thead><tr><th>Title</th><th>Author</th><th>Publisher</th><th>Year</th><th>Lang</th><th>Pages</th><th>Size</th><th>Ext</th><th>Mirrors</th></tr></thead>
<tbody>
</tbody>
</table>
</body>
</html>`
		} else {
			html = `<!DOCTYPE html>
<html>
<body>
<table class="table table-striped" id="tablelibgen">
<thead><tr><th>Title</th><th>Author</th><th>Publisher</th><th>Year</th><th>Lang</th><th>Pages</th><th>Size</th><th>Ext</th><th>Mirrors</th></tr></thead>
<tbody>
<tr>
  <td><a href="edition.php?id=123">The Turing Test and the Frame Problem: AI's Mistaken Understanding of Intelligence</a><nobr><span class="badge badge-secondary">l 643</span></nobr></td>
  <td>Larry J. Crockett</td>
  <td>Ablex Publishing</td>
  <td><nobr>1994</nobr></td>
  <td>English</td>
  <td>230</td>
  <td><nobr><a href="/file.php?id=643">100 KB</a></nobr></td>
  <td>gz</td>
  <td><nobr><a href="/ads.php?md5=2F2DBA2A621B693BB95601C16ED680F8">1</a></nobr></td>
</tr>
<tr>
  <td><a href="edition.php?id=124">You failed your math test, Comrade Einstein (about Soviet antisemitism)</a><nobr><span class="badge badge-secondary">l 3167</span></nobr></td>
  <td>M. Shifman</td>
  <td>World Scientific Publishing</td>
  <td><nobr>2005</nobr></td>
  <td>English</td>
  <td>230</td>
  <td><nobr><a href="/file.php?id=3167">200 KB</a></nobr></td>
  <td>djvu</td>
  <td><nobr><a href="/ads.php?md5=06E6135019C8F2F43158ABA9ABDC610E">1</a></nobr></td>
</tr>
<tr>
  <td><a href="edition.php?id=125">Test Math Document</a><nobr><span class="badge badge-secondary">l 5539</span></nobr></td>
  <td>Test Author</td>
  <td>Math Pub</td>
  <td><nobr>2010</nobr></td>
  <td>English</td>
  <td>150</td>
  <td><nobr><a href="/file.php?id=5539">300 KB</a></nobr></td>
  <td>pdf</td>
  <td><nobr><a href="/ads.php?md5=553907CDF5F03AF78950561F42F1571A">1</a></nobr></td>
</tr>
</tbody>
</table>
</body>
</html>`
		}
		bodyReader = io.NopCloser(strings.NewReader(html))

	case strings.Contains(path, "search.php"):
		header.Set("Content-Type", "text/html; charset=utf-8")
		html := `<!DOCTYPE html>
<html>
<body>
<table>
<tr>
  <td><a href="book/index.php?md5=2F2DBA2A621B693BB95601C16ED680F8">The Turing Test and the Frame Problem</a></td>
</tr>
</table>
</body>
</html>`
		bodyReader = io.NopCloser(strings.NewReader(html))

	case strings.Contains(path, "ads.php") || strings.Contains(path, "main"):
		header.Set("Content-Type", "text/html; charset=utf-8")
		html := `<!DOCTYPE html>
<html>
<body>
@book{book,
  title = {The Turing Test and the Frame Problem: AI's Mistaken Understanding of Intelligence},
  author = {Larry J. Crockett},
  publisher = {Ablex Publishing},
  year = {1994},
}
Author(s): Larry J. Crockett<br>
<a href="get.php?md5=1794743bb21d72736ffe64d66dca9f0e&key=1234567812345678">GET</a>
<a href="https://libgen.li/get.php?md5=1794743bb21d72736ffe64d66dca9f0e&key=1234567812345678">GET</a>
<a href="https://download.library.lol/main/1440000/1794743bb21d72736ffe64d66dca9f0e/book.epub">GET</a>
<a href="https://gateway.ipfs.io/ipfs/bafykbzacectwnzckgcrnozlrkx7j5fbdwlf6qo7whmf2sksafwfwvunazyl4e?filename=book.epub">IPFS Gateway 1</a>
</body>
</html>`
		bodyReader = io.NopCloser(strings.NewReader(html))

	case strings.Contains(path, "file.php"):
		header.Set("Content-Type", "text/html; charset=utf-8")
		html := `<!DOCTYPE html>
<html>
<body>
<a href="https://cloudflare-ipfs.com/ipfs/bafykbzacectwnzckgcrnozlrkx7j5fbdwlf6qo7whmf2sksafwfwvunazyl4e?filename=book.epub">IPFS cloudflare</a>
<a href="https://gateway.ipfs.io/ipfs/bafykbzacectwnzckgcrnozlrkx7j5fbdwlf6qo7whmf2sksafwfwvunazyl4e?filename=book.epub">IPFS.io</a>
<a href="/ads.php?md5=1794743bb21d72736ffe64d66dca9f0e">Libgen</a>
</body>
</html>`
		bodyReader = io.NopCloser(strings.NewReader(html))

	case strings.Contains(path, "dbdumps") || strings.Contains(path, "dirlist"):
		header.Set("Content-Type", "text/html; charset=utf-8")
		html := `<!DOCTYPE html>
<html>
<body>
<a href="libgen_compact_2023-10-31.rar">"libgen_compact_2023-10-31.rar"</a>
<a href="libgen_fiction_2023-10-31.sql.gz">"libgen_fiction_2023-10-31.sql.gz"</a>
</body>
</html>`
		bodyReader = io.NopCloser(strings.NewReader(html))

	case strings.Contains(path, "get.php") || strings.Contains(path, "download"):
		header.Set("Content-Type", "application/octet-stream")
		header.Set("Content-Length", "28")
		bodyReader = io.NopCloser(strings.NewReader("Mock book binary content OK\n"))

	default:
		bodyReader = io.NopCloser(strings.NewReader("OK"))
	}

	res := &http.Response{
		StatusCode:    statusCode,
		Header:        header,
		Body:          bodyReader,
		ContentLength: -1,
		Request:       req,
	}
	return res, nil
}

func TestMain(m *testing.M) {
	// Unless LIBGEN_TEST_LIVE=1 is set, use the in-memory mock transport for deterministic tests.
	if os.Getenv("LIBGEN_TEST_LIVE") != "1" {
		DefaultTransport = &mockTransport{}
		http.DefaultTransport = &mockTransport{}
		mockURL, _ := url.Parse("https://libgen.li/index.php")
		SearchMirrors = []url.URL{*mockURL}
		DownloadMirrors = []url.URL{
			{Scheme: "https", Host: "libgen.li", Path: "ads.php"},
			{Scheme: "https", Host: "library.lol", Path: "main/"},
		}
		DbdumpsMirrors = []url.URL{
			{Scheme: "https", Host: "libgen.rs", Path: "dbdumps"},
		}
	}

	exitCode := m.Run()
	_ = os.RemoveAll("libgen")
	os.Exit(exitCode)
}

func TestParseMirrorURL(t *testing.T) {
	tests := []struct {
		input       string
		defaultPath string
		wantHost    string
		wantPath    string
		wantScheme  string
		wantErr     bool
	}{
		{"libgen.li", "index.php", "libgen.li", "index.php", "https", false},
		{"https://libgen.is", "index.php", "libgen.is", "index.php", "https", false},
		{"http://custom.mirror/search.php", "index.php", "custom.mirror", "/search.php", "http", false},
		{"   libgen.rs/   ", "ads.php", "libgen.rs", "ads.php", "https", false},
		{"", "index.php", "", "", "", true},
	}

	for _, tt := range tests {
		u, err := ParseMirrorURL(tt.input, tt.defaultPath)
		if tt.wantErr {
			if err == nil {
				t.Errorf("ParseMirrorURL(%q) expected error, got nil", tt.input)
			}
			continue
		}
		if err != nil {
			t.Errorf("ParseMirrorURL(%q) unexpected error: %v", tt.input, err)
			continue
		}
		if u.Host != tt.wantHost || u.Path != tt.wantPath || u.Scheme != tt.wantScheme {
			t.Errorf("ParseMirrorURL(%q) = (%s, %s, %s), want (%s, %s, %s)",
				tt.input, u.Scheme, u.Host, u.Path, tt.wantScheme, tt.wantHost, tt.wantPath)
		}
	}
}

func TestSanitizeFilename(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{"Book: Subtitle / Part 1 *", "Book - Subtitle _ Part 1 _"},
		{"Hello\nWorld\r\tTitle", "Hello World Title"},
		{"Evil\x00Name? <|> \"quoted\"", "EvilName_ ___ 'quoted'"},
		{"   Spaced    Out   ", "Spaced Out"},
	}
	for _, c := range cases {
		got := sanitizeFilename(c.input)
		if got != c.want {
			t.Errorf("sanitizeFilename(%q) = %q, want %q", c.input, got, c.want)
		}
	}
}

func TestExtractCID(t *testing.T) {
	cases := []struct {
		urlStr string
		want   string
	}{
		{"https://gateway.ipfs.io/ipfs/bafykbzacectwnzckgcrnozlrkx7j5fbdwlf6qo7whmf2sksafwfwvunazyl4e?filename=b.epub", "bafykbzacectwnzckgcrnozlrkx7j5fbdwlf6qo7whmf2sksafwfwvunazyl4e"},
		{"https://ipfs.io/ipfs/QmXoypizjW3WknFiJnKLwHCnL72vedxjQkDDP1mXWo6uco", "QmXoypizjW3WknFiJnKLwHCnL72vedxjQkDDP1mXWo6uco"},
		{"bafykbzacectwnzckgcrnozlrkx7j5fbdwlf6qo7whmf2sksafwfwvunazyl4e", "bafykbzacectwnzckgcrnozlrkx7j5fbdwlf6qo7whmf2sksafwfwvunazyl4e"},
		{"not-a-cid", ""},
	}
	for _, c := range cases {
		got := extractCID(c.urlStr)
		if got != c.want {
			t.Errorf("extractCID(%q) = %q, want %q", c.urlStr, got, c.want)
		}
	}
}

func TestValidMD5(t *testing.T) {
	re := regexp.MustCompile(ValidMD5)
	if !re.MatchString("2F2DBA2A621B693BB95601C16ED680F8") {
		t.Errorf("expected valid uppercase MD5 to match")
	}
	if !re.MatchString("2f2dba2a621b693bb95601c16ed680f8") {
		t.Errorf("expected valid lowercase MD5 to match")
	}
	if re.MatchString("not-an-md5") {
		t.Errorf("invalid string matched ValidMD5")
	}
	if re.MatchString("2F2DBA2A621B693BB95601C16ED680F8_extra") {
		t.Errorf("unanchored string matched ValidMD5")
	}
	if re.MatchString("2F2DBA2A621B693BB95601C16ED680GZ") {
		t.Errorf("non-hex characters matched ValidMD5")
	}
}

func TestGetGenericMirrorURL(t *testing.T) {
	book := &Book{Md5: "1794743bb21d72736ffe64d66dca9f0e"}
	mirror := url.URL{Scheme: "https", Host: "libgen.li", Path: "ads.php"}
	err := getGenericMirrorURL(mirror, book, false)
	if err != nil {
		t.Fatalf("getGenericMirrorURL error: %v", err)
	}
	if !strings.Contains(book.DownloadURL, "get.php?md5=1794743bb21d72736ffe64d66dca9f0e") {
		t.Errorf("unexpected DownloadURL: %s", book.DownloadURL)
	}
}

func TestDbdumpsParsing(t *testing.T) {
	mirror := url.URL{Scheme: "https", Host: "libgen.rs", Path: "dbdumps"}
	b, err := GetBody(mirror.String())
	if err != nil {
		t.Fatalf("GetBody error: %v", err)
	}
	dumps := ParseDbdumps(b)
	if len(dumps) != 2 {
		t.Fatalf("expected 2 dumps, got %d: %v", len(dumps), dumps)
	}
	if dumps[0] != "libgen_compact_2023-10-31.rar" {
		t.Errorf("unexpected dump: %s", dumps[0])
	}
}

func TestGetDetails_Resilient(t *testing.T) {
	// One valid hash and one hash that fails parsing or returns error
	books, err := GetDetails(&GetDetailsOptions{
		Hashes: []string{
			"2F2DBA2A621B693BB95601C16ED680F8", // valid
		},
		SearchMirror: SearchMirrors[0],
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(books) != 1 {
		t.Fatalf("expected 1 book, got %d", len(books))
	}
	if books[0].ID != "643" {
		t.Errorf("expected ID 643, got %s", books[0].ID)
	}
}

func TestDownloadBookIPFS_Resilient(t *testing.T) {
	book := &Book{
		Title:       "Test IPFS Book",
		Author:      "Author",
		Extension:   "epub",
		DownloadURL: "https://gateway.ipfs.io/ipfs/bafykbzacectwnzckgcrnozlrkx7j5fbdwlf6qo7whmf2sksafwfwvunazyl4e?filename=book.epub",
	}
	err := DownloadBookIPFS(book, "")
	if err != nil {
		t.Fatalf("DownloadBookIPFS error: %v", err)
	}
}

func TestSearchPagination(t *testing.T) {
	mirror := url.URL{Scheme: "https", Host: "libgen.li", Path: "index.php"}

	// Page 1 should return gz, djvu, pdf
	booksP1, err := Search(&SearchOptions{
		Query:        "test",
		SearchMirror: mirror,
		Page:         1,
		Results:      5,
		MaxPages:     1,
	})
	if err != nil {
		t.Fatalf("Page 1 error: %v", err)
	}
	if len(booksP1) != 3 {
		t.Fatalf("expected 3 books on page 1, got %d", len(booksP1))
	}
	if booksP1[0].Extension != "gz" {
		t.Errorf("expected gz on page 1, got %s", booksP1[0].Extension)
	}

	// Page 2 should return epub
	booksP2, err := Search(&SearchOptions{
		Query:        "test",
		SearchMirror: mirror,
		Page:         2,
		Results:      5,
		MaxPages:     1,
	})
	if err != nil {
		t.Fatalf("Page 2 error: %v", err)
	}
	if len(booksP2) != 1 {
		t.Fatalf("expected 1 book on page 2, got %d", len(booksP2))
	}
	if booksP2[0].Extension != "epub" {
		t.Errorf("expected epub on page 2, got %s", booksP2[0].Extension)
	}
	if booksP2[0].ID != "9999" {
		t.Errorf("expected ID 9999 on page 2, got %s", booksP2[0].ID)
	}
}

func TestIsNetworkOrDNSError(t *testing.T) {
	cases := []struct {
		err  error
		want bool
	}{
		{nil, false},
		{errors.New("dial tcp: lookup libgen.gs: no such host"), true},
		{errors.New("Get \"https://libgen.li\": i/o timeout"), true},
		{errors.New("context deadline exceeded"), true},
		{errors.New("connection refused"), true},
		{errors.New("network is unreachable"), true},
		{errors.New("read: connection reset by peer"), true},
		{errors.New("unexpected EOF"), false},
		{errors.New("no books found matching criteria"), false},
	}
	for _, c := range cases {
		got := isNetworkOrDNSError(c.err)
		if got != c.want {
			t.Errorf("isNetworkOrDNSError(%v) = %v, want %v", c.err, got, c.want)
		}
	}
}

func TestGetWorkingMirror_Concurrent(t *testing.T) {
	urls := []url.URL{
		{Scheme: "https", Host: "libgen.li", Path: "index.php"},
		{Scheme: "https", Host: "libgen.vg", Path: "index.php"},
	}
	working := GetWorkingMirror(urls)
	if working.Host == "" {
		t.Errorf("expected non-empty working mirror host")
	}
}

func TestSearchAutoPaging(t *testing.T) {
	mirror := url.URL{Scheme: "https", Host: "libgen.li", Path: "index.php"}

	// When filtering for epub, page 1 has 0 epubs, so auto-paging should fetch page 2 and find the epub!
	books, err := Search(&SearchOptions{
		Query:        "test",
		SearchMirror: mirror,
		Extension:    []string{"epub"},
		Results:      1,
		MaxPages:     3,
	})
	if err != nil {
		t.Fatalf("Auto-paging error: %v", err)
	}
	if len(books) != 1 {
		t.Fatalf("expected 1 epub via auto-paging, got %d", len(books))
	}
	if books[0].Extension != "epub" {
		t.Errorf("expected epub, got %s", books[0].Extension)
	}
	if books[0].ID != "9999" {
		t.Errorf("expected ID 9999, got %s", books[0].ID)
	}
}

func TestSearchContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	mirror := url.URL{Scheme: "https", Host: "libgen.li", Path: "index.php"}
	_, err := Search(&SearchOptions{
		Query:        "test",
		SearchMirror: mirror,
		Context:      ctx,
	})
	if err == nil {
		t.Fatalf("expected context cancellation error, got nil")
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}

func TestSearchWorkingMirrorReporting(t *testing.T) {
	var working url.URL
	books, err := Search(&SearchOptions{
		Query:         "test",
		SearchMirror:  url.URL{Scheme: "https", Host: "libgen.li", Path: "index.php"},
		Results:       1,
		WorkingMirror: &working,
	})
	if err != nil {
		t.Fatalf("Search error: %v", err)
	}
	if len(books) == 0 {
		t.Fatalf("expected books, got 0")
	}
	if working.Host != "libgen.li" {
		t.Errorf("expected WorkingMirror host libgen.li, got %s", working.Host)
	}
}

