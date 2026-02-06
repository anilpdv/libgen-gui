// Copyright © 2023 Ryan Ciehanski <ryan@ciehanski.com>
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
	"regexp"
	"strings"

	"github.com/cheggaaa/pb/v3"
)

// IPFSGateways contains reliable public IPFS gateways used for downloading
// IPFS-hosted content.
var IPFSGateways = []string{
	"https://cloudflare-ipfs.com/ipfs/",
	"https://ipfs.io/ipfs/",
	"https://dweb.link/ipfs/",
	"https://gateway.pinata.cloud/ipfs/",
}

// DownloadBookIPFS downloads the requested book via HTTP through an IPFS gateway.
// This replaces the heavy in-process Kubo node with fast and lightweight gateway requests.
func DownloadBookIPFS(book *Book, outputPath string) error {
	filename := getBookFilename(book)

	targetURL := book.DownloadURL
	cid := extractCID(targetURL)
	if cid == "" && !strings.HasPrefix(targetURL, "http://") && !strings.HasPrefix(targetURL, "https://") {
		return errors.New("unable to determine IPFS CID from URL")
	}

	var candidateURLs []string
	if strings.HasPrefix(targetURL, "http://") || strings.HasPrefix(targetURL, "https://") {
		candidateURLs = append(candidateURLs, targetURL)
	}
	if cid != "" {
		for _, gw := range IPFSGateways {
			gwURL := gw + cid
			alreadyAdded := false
			for _, u := range candidateURLs {
				if u == gwURL {
					alreadyAdded = true
					break
				}
			}
			if !alreadyAdded {
				candidateURLs = append(candidateURLs, gwURL)
			}
		}
	}

	client := getHTTPClient(HTTPClientTimeout * 2)

	var r *http.Response
	var lastErr error

	for _, u := range candidateURLs {
		req, reqErr := http.NewRequest("GET", u, nil)
		if reqErr != nil {
			lastErr = reqErr
			continue
		}
		req.Header.Set("User-Agent", UserAgent)
		req.Header.Set("Accept-Encoding", "*")

		resp, doErr := client.Do(req)
		if doErr != nil {
			lastErr = doErr
			continue
		}
		if resp.StatusCode == http.StatusOK {
			r = resp
			break
		}
		resp.Body.Close()
		lastErr = fmt.Errorf("HTTP %v from %s", resp.StatusCode, u)
	}

	if r == nil {
		if lastErr != nil {
			return fmt.Errorf("error downloading from IPFS: %w", lastErr)
		}
		return errors.New("unable to download via IPFS: all gateways failed")
	}
	defer r.Body.Close()

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
	return nil
}

func extractCID(urlStr string) string {
	re := regexp.MustCompile(ipfsCIDReg)
	m := re.FindStringSubmatch(urlStr)
	if len(m) > 1 {
		return m[1]
	}
	// Fallback to ipfsReg
	reLegacy := regexp.MustCompile(ipfsReg)
	mLegacy := reLegacy.FindStringSubmatch(urlStr)
	if len(mLegacy) > 1 {
		return mLegacy[1]
	}
	return ""
}
