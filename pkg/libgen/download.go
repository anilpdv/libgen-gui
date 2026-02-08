// Copyright © 2019 Antoine Chiny <antoine.chiny@inria.fr>
// Copyright © 2019 Ryan Ciehanski <ryan@ciehanski.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package libgen

import (
	"errors"
	"fmt"
	"io"
	"net/http"

	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/cheggaaa/pb/v3"
)

// DownloadBook grabs the download URL for the book requested and initiates
// the download process with a progress bar displayed to the user's CLI.
func DownloadBook(book *Book, outputPath string) error {
	var filesize int64
	filename := getBookFilename(book)

	req, err := http.NewRequest("GET", book.DownloadURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", UserAgent)
	req.Header.Add("Accept-Encoding", "*")
	if book.PageURL != "" {
		req.Header.Set("Referer", book.PageURL)
	}

	client := getHTTPClient(time.Minute * 10)
	r, err := client.Do(req)
	if err != nil {
		return err
	}
	defer r.Body.Close()

	if r.StatusCode == http.StatusOK {
		filesize = r.ContentLength
		bar := pb.Full.Start64(filesize)

		out, err := makeFile(outputPath, filename)
		if err != nil {
			return err
		}
		defer out.Close()

		_, err = io.Copy(out, bar.NewProxyReader(r.Body))
		if err != nil {
			return err
		}

		bar.Finish()
	} else {
		return fmt.Errorf("unable to reach mirror %v: HTTP %v", req.Host, r.StatusCode)
	}

	return nil
}

// GetDownloadURL picks download mirrors and resolves the direct download link.
func GetDownloadURL(book *Book, useIpfs bool) error {
	// Check environment variable override
	if envMirror := os.Getenv("LIBGEN_DOWNLOAD_MIRROR"); envMirror != "" {
		if parsed, err := ParseMirrorURL(envMirror, "ads.php"); err == nil {
			if err := resolveMirrorDownloadURL(parsed, book, useIpfs); err == nil && book.DownloadURL != "" {
				return nil
			}
		}
	}

	// 1. If book has a PageURL from search, try that host first!
	if book.PageURL != "" {
		if u, err := url.Parse(book.PageURL); err == nil && u.Host != "" {
			pageMirror := url.URL{Scheme: u.Scheme, Host: u.Host, Path: "ads.php"}
			if err := resolveMirrorDownloadURL(pageMirror, book, useIpfs); err == nil && book.DownloadURL != "" {
				return nil
			}
		}
	}

	// 2. Try verified working mirror
	working := GetWorkingMirror(DownloadMirrors)
	if working.Host != "" {
		if err := resolveMirrorDownloadURL(working, book, useIpfs); err == nil && book.DownloadURL != "" {
			return nil
		}
	}

	// 3. Try remaining download mirrors
	for _, mirror := range DownloadMirrors {
		if mirror.Host != working.Host {
			if err := resolveMirrorDownloadURL(mirror, book, useIpfs); err == nil && book.DownloadURL != "" {
				return nil
			}
		}
	}

	return fmt.Errorf("unable to retrieve download link for desired resource")
}


// ResolveMirrorDownloadURL resolves the direct download link using a specific mirror.
func ResolveMirrorDownloadURL(mirror url.URL, book *Book, useIpfs bool) error {
	return resolveMirrorDownloadURL(mirror, book, useIpfs)
}

func resolveMirrorDownloadURL(mirror url.URL, book *Book, useIpfs bool) error {
	switch {
	case strings.Contains(mirror.Hostname(), "library.lol"):
		return getLibraryLolURL(book, useIpfs)
	case strings.Contains(mirror.Hostname(), "libgen.pm"):
		return getLibgenPMURL(book)
	default:
		// Modern mirror resolution (libgen.li, libgen.vg, libgen.bz, libgen.is, etc.)
		return getGenericMirrorURL(mirror, book, useIpfs)
	}
}

func getGenericMirrorURL(mirror url.URL, book *Book, useIpfs bool) error {
	queryURL := fmt.Sprintf("%s://%s/ads.php?md5=%s", mirror.Scheme, mirror.Host, strings.ToLower(book.Md5))
	book.PageURL = queryURL

	referer := fmt.Sprintf("%s://%s/index.php", mirror.Scheme, mirror.Host)
	b, err := GetBodyWithReferer(queryURL, referer)
	if err != nil {
		return err
	}

	if useIpfs {
		// Try file.php for direct IPFS links
		fileURL := fmt.Sprintf("%s://%s/file.php?md5=%s", mirror.Scheme, mirror.Host, strings.ToLower(book.Md5))
		if bFile, errFile := GetBodyWithReferer(fileURL, referer); errFile == nil {
			reIPFS := regexp.MustCompile(libraryLolIPFSCFReg)
			if m := reIPFS.Find(bFile); m != nil {
				book.DownloadURL = string(m)
				return nil
			}
			reCID := regexp.MustCompile(ipfsCIDReg)
			if m := reCID.FindSubmatch(bFile); len(m) > 1 {
				book.DownloadURL = IPFSGateways[0] + string(m[1])
				return nil
			}
		}

		// Look for IPFS link or CID in ads.php
		reIPFS := regexp.MustCompile(libraryLolIPFSCFReg)
		if m := reIPFS.Find(b); m != nil {
			book.DownloadURL = string(m)
			return nil
		}
		reCID := regexp.MustCompile(ipfsCIDReg)
		if m := reCID.FindSubmatch(b); len(m) > 1 {
			book.DownloadURL = IPFSGateways[0] + string(m[1])
			return nil
		}
		return errors.New("no IPFS link found on mirror")
	}

	// 1. Look for get.php?md5=...&key=...
	reGet := regexp.MustCompile(libgenGetReg)
	if m := reGet.Find(b); m != nil {
		path := string(m)
		if !strings.HasPrefix(path, "/") {
			path = "/" + path
		}
		book.DownloadURL = fmt.Sprintf("%s://%s%s", mirror.Scheme, mirror.Host, path)
		return nil
	}

	// 2. Look for GET / DOWNLOAD button href: <a ... href="..." ...>GET</a>
	reHref := regexp.MustCompile(`(?i)<a\s+[^>]*href=['"]([^'"]+)['"][^>]*>[\s\S]*?(?:GET|DOWNLOAD)[\s\S]*?</a>`)
	if m := reHref.FindSubmatch(b); len(m) > 1 {
		link := string(m[1])
		if strings.HasPrefix(link, "http://") || strings.HasPrefix(link, "https://") {
			book.DownloadURL = link
			return nil
		}
		if !strings.HasPrefix(link, "/") {
			link = "/" + link
		}
		book.DownloadURL = fmt.Sprintf("%s://%s%s", mirror.Scheme, mirror.Host, link)
		return nil
	}

	// 3. Look for download.library.lol URL
	reLibLol := regexp.MustCompile(libraryLolReg)
	if m := reLibLol.Find(b); m != nil {
		book.DownloadURL = string(m)
		return nil
	}

	return errors.New("no valid download URL found on mirror")
}

// DownloadDbdump downloads the selected database dump from
// Library Genesis.
func DownloadDbdump(filename string, outputPath string) error {
	mirror := GetWorkingMirror(DbdumpsMirrors)
	client := getHTTPClient(HTTPClientTimeout * 4)

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/%s", mirror.String(), filename), nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", UserAgent)

	r, err := client.Do(req)
	if err != nil {
		return err
	}
	defer r.Body.Close()

	if r.StatusCode == http.StatusOK {
		filesize := r.ContentLength
		bar := pb.Full.Start64(filesize)

		out, err := makeFile(outputPath, filename)
		if err != nil {
			return err
		}
		defer out.Close()

		_, err = io.Copy(out, bar.NewProxyReader(r.Body))
		if err != nil {
			return err
		}

		bar.Finish()
	} else {
		return fmt.Errorf("unable to reach mirror: HTTP %v", r.StatusCode)
	}

	return nil
}

func getLibraryLolURL(book *Book, useIpfs bool) error {
	queryURL := fmt.Sprintf("https://library.lol/main/%s", strings.ToLower(book.Md5))
	book.PageURL = queryURL

	b, err := getBody(queryURL)
	if err != nil {
		return err
	}

	downloadURL := []byte{}
	if useIpfs {
		// Attempt to find IPFS download URL via gateway.ipfs.io
		downloadURL = findMatch(libraryLolIPFSReg, b)
		if downloadURL == nil {
			// Fallback to cloudflare-ipfs.com
			downloadURL = findMatch(libraryLolIPFSCFReg, b)
			if downloadURL == nil {
				return errors.New("no valid LibraryLol IPFS download URL found")
			}
		}
	} else {
		downloadURL = findMatch(libraryLolReg, b)
		if downloadURL == nil {
			// Also check for standard href on GET button
			reHref := regexp.MustCompile(`(?i)<a\s+[^>]*href=['"]([^'"]+)['"][^>]*>GET</a>`)
			if m := reHref.FindSubmatch(b); len(m) > 1 {
				downloadURL = m[1]
			} else {
				return errors.New("no valid LibraryLol download URL found")
			}
		}
	}

	book.DownloadURL = string(downloadURL)
	return nil
}

func getLibgenPMURL(book *Book) error {
	queryURL := fmt.Sprintf("https://libgen.li/ads.php?md5=%s", strings.ToLower(book.Md5))
	book.PageURL = queryURL

	b, err := getBody(queryURL)
	if err != nil {
		return err
	}

	downloadURL := findMatch(libgenPMReg, b)
	if downloadURL == nil {
		downloadURL = findMatch(libgenGetReg, b)
		if downloadURL == nil {
			return errors.New("no valid Libgen download URL found")
		}
	}
	path := string(downloadURL)
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	book.DownloadURL = fmt.Sprintf("https://libgen.li%s", path)

	return nil
}

func makeFile(outputPath, filename string) (*os.File, error) {
	var out *os.File
	var mkErr error

	// Handle long titles
	if len(filename) >= 256 {
		filename = filename[:255]
	}

	// if output path was not provided
	if outputPath == "" {
		wd, err := os.Getwd()
		if err != nil {
			return nil, err
		}
		targetDir := filepath.Join(wd, "libgen")
		if stat, err := os.Stat(targetDir); err != nil || !stat.IsDir() {
			if err := os.MkdirAll(targetDir, 0755); err != nil {
				return nil, err
			}
		}
		out, mkErr = os.Create(filepath.Join(targetDir, filename))
		if mkErr != nil {
			return nil, mkErr
		}
	} else {
		// If output path was provided
		if stat, err := os.Stat(outputPath); err == nil && stat.IsDir() {
			out, err = os.Create(filepath.Join(outputPath, filename))
			if err != nil {
				return nil, err
			}
		} else {
			if err := os.MkdirAll(outputPath, 0755); err != nil {
				return nil, errors.New("invalid output path")
			}
			out, err = os.Create(filepath.Join(outputPath, filename))
			if err != nil {
				return nil, err
			}
		}
	}

	return out, nil
}

// findMatch is a helper function that searches an []byte
// for a specified regex and returns the matches.
func findMatch(reg string, response []byte) []byte {
	re := regexp.MustCompile(reg)
	match := re.FindString(string(response))

	if match != "" {
		return []byte(match)
	}

	return nil
}

func getBookFilename(book *Book) string {
	var tmp []string
	cleanTitle := sanitizeFilename(book.Title)
	tmp = append(tmp, cleanTitle)
	if book.Author != "" {
		cleanAuthor := sanitizeFilename(book.Author)
		tmp = append(tmp, fmt.Sprintf(" by %s", cleanAuthor))
	}
	ext := book.Extension
	if ext == "" {
		ext = "pdf"
	}
	tmp = append(tmp, fmt.Sprintf(".%s", ext))
	return strings.Join(tmp, "")
}

func sanitizeFilename(s string) string {
	s = strings.ReplaceAll(s, "\x00", "")
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", " ")
	s = strings.ReplaceAll(s, "\t", " ")
	s = strings.ReplaceAll(s, "/", "_")
	s = strings.ReplaceAll(s, "\\", "_")
	s = strings.ReplaceAll(s, ":", " -")
	s = strings.ReplaceAll(s, "*", "_")
	s = strings.ReplaceAll(s, "?", "_")
	s = strings.ReplaceAll(s, "\"", "'")
	s = strings.ReplaceAll(s, "<", "_")
	s = strings.ReplaceAll(s, ">", "_")
	s = strings.ReplaceAll(s, "|", "_")
	for strings.Contains(s, "  ") {
		s = strings.ReplaceAll(s, "  ", " ")
	}
	return strings.TrimSpace(s)
}
