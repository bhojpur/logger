package alils

// Copyright (c) 2018 Bhojpur Consulting Private Limited, India. All rights reserved.

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"time"
)

// GMT location
var gmtLoc = time.FixedZone("GMT", 0)

// NowRFC1123 returns now time in RFC1123 format with GMT timezone,
// e.g. "Mon, 02 Jan 2006 15:04:05 GMT".
func nowRFC1123() string {
	return time.Now().In(gmtLoc).Format(time.RFC1123)
}

// signature calculates a request's signature digest.
func signature(project *LogProject, method, uri string,
	headers map[string]string) (digest string, err error) {
	var contentMD5, contentType, date, canoHeaders, canoResource string
	var slsHeaderKeys sort.StringSlice

	// SignString = VERB + "\n"
	//              + CONTENT-MD5 + "\n"
	//              + CONTENT-TYPE + "\n"
	//              + DATE + "\n"
	//              + CanonicalizedSLSHeaders + "\n"
	//              + CanonicalizedResource

	if val, ok := headers["Content-MD5"]; ok {
		contentMD5 = val
	}

	if val, ok := headers["Content-Type"]; ok {
		contentType = val
	}

	date, ok := headers["Date"]
	if !ok {
		err = fmt.Errorf("Can't find 'Date' header")
		return
	}

	// Calc CanonicalizedSLSHeaders
	slsHeaders := make(map[string]string, len(headers))
	for k, v := range headers {
		l := strings.TrimSpace(strings.ToLower(k))
		if strings.HasPrefix(l, "x-sls-") {
			slsHeaders[l] = strings.TrimSpace(v)
			slsHeaderKeys = append(slsHeaderKeys, l)
		}
	}

	sort.Sort(slsHeaderKeys)
	for i, k := range slsHeaderKeys {
		canoHeaders += k + ":" + slsHeaders[k]
		if i+1 < len(slsHeaderKeys) {
			canoHeaders += "\n"
		}
	}

	// Calc CanonicalizedResource
	u, err := url.Parse(uri)
	if err != nil {
		return
	}

	canoResource += url.QueryEscape(u.Path)
	if u.RawQuery != "" {
		var keys sort.StringSlice

		vals := u.Query()
		for k := range vals {
			keys = append(keys, k)
		}

		sort.Sort(keys)
		canoResource += "?"
		for i, k := range keys {
			if i > 0 {
				canoResource += "&"
			}

			for _, v := range vals[k] {
				canoResource += k + "=" + v
			}
		}
	}

	signStr := method + "\n" +
		contentMD5 + "\n" +
		contentType + "\n" +
		date + "\n" +
		canoHeaders + "\n" +
		canoResource

	// Signature = base64(hmac-sha1(UTF8-Encoding-Of(SignString)，AccessKeySecret))
	mac := hmac.New(sha1.New, []byte(project.AccessKeySecret))
	_, err = mac.Write([]byte(signStr))
	if err != nil {
		return
	}
	digest = base64.StdEncoding.EncodeToString(mac.Sum(nil))
	return
}
