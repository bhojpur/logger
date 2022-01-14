package logutil

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
	"fmt"
	"os"
	"path"
	"path/filepath"
	"testing"
	"time"
)

func TestParsing(t *testing.T) {

	path := []string{
		"/tmp/something.foo/zkocc.goedel.szopa.log.INFO.20130806-151006.10530",
		"/tmp/something.foo/zkocc.goedel.szopa.test.log.ERROR.20130806-151006.10530"}

	for _, filepath := range path {
		ts, err := parseCreatedTimestamp(filepath)
		if err != nil {
			t.Fatalf("parse: %v", err)
		}

		if want := time.Date(2013, 8, 6, 15, 10, 06, 0, time.Now().Location()); ts != want {
			t.Errorf("timestamp: want %v, got %v", want, ts)
		}
	}
}

func TestPurgeByCtime(t *testing.T) {
	logDir := path.Join(os.TempDir(), fmt.Sprintf("%v-%v", os.Args[0], os.Getpid()))
	if err := os.MkdirAll(logDir, 0777); err != nil {
		t.Fatalf("os.MkdirAll: %v", err)
	}
	defer os.RemoveAll(logDir)

	now := time.Date(2013, 8, 6, 15, 10, 06, 0, time.Now().Location())
	files := []string{
		"zkocc.goedel.szopa.log.INFO.20130806-121006.10530",
		"zkocc.goedel.szopa.log.INFO.20130806-131006.10530",
		"zkocc.goedel.szopa.log.INFO.20130806-141006.10530",
		"zkocc.goedel.szopa.log.INFO.20130806-151006.10530",
	}

	for _, file := range files {
		if _, err := os.Create(path.Join(logDir, file)); err != nil {
			t.Fatalf("os.Create: %v", err)
		}
	}
	if err := os.Symlink(files[1], path.Join(logDir, "zkocc.INFO")); err != nil {
		t.Fatalf("os.Symlink: %v", err)
	}

	purgeLogsOnce(now, logDir, "zkocc", 30*time.Minute, 0)

	left, err := filepath.Glob(path.Join(logDir, "zkocc.*"))
	if err != nil {
		t.Fatalf("filepath.Glob: %v", err)
	}

	if len(left) != 3 {
		// 131006 is current
		// 151006 is within 30 min
		// symlink remains
		// the rest should be removed.
		t.Errorf("wrong number of files remain: want %v, got %v", 3, len(left))
	}
}

func TestPurgeByMtime(t *testing.T) {
	logDir := path.Join(os.TempDir(), fmt.Sprintf("%v-%v", os.Args[0], os.Getpid()))
	if err := os.MkdirAll(logDir, 0777); err != nil {
		t.Fatalf("os.MkdirAll: %v", err)
	}
	defer os.RemoveAll(logDir)
	createFileWithMtime := func(filename, mtimeStr string) {
		var err error
		var mtime time.Time
		filepath := path.Join(logDir, filename)
		if mtime, err = time.Parse(time.RFC3339, mtimeStr); err != nil {
			t.Fatalf("time.Parse: %v", err)
		}
		if _, err = os.Create(filepath); err != nil {
			t.Fatalf("os.Create: %v", err)
		}
		if err = os.Chtimes(filepath, mtime, mtime); err != nil {
			t.Fatalf("os.Chtimes: %v", err)
		}
	}
	now := time.Date(2020, 1, 1, 12, 0, 0, 0, time.UTC)
	filenameMtimeMap := map[string]string{
		"adam.localhost.bhojpur.log.INFO.20200101-113000.00000": "2020-01-01T11:30:00.000Z",
		"adam.localhost.bhojpur.log.INFO.20200101-100000.00000": "2020-01-01T10:00:00.000Z",
		"adam.localhost.bhojpur.log.INFO.20200101-090000.00000": "2020-01-01T09:00:00.000Z",
		"adam.localhost.bhojpur.log.INFO.20200101-080000.00000": "2020-01-01T08:00:00.000Z",
	}
	for filename, mtimeStr := range filenameMtimeMap {
		createFileWithMtime(filename, mtimeStr)
	}

	// Create adam.INFO symlink to 100000. This is a contrived example in that
	// current log (100000) is not the latest log (113000). This will not happen
	// IRL but it helps us test edge cases of purging by mtime.
	if err := os.Symlink("adam.localhost.bhojpur.log.INFO.20200101-100000.00000", path.Join(logDir, "adam.INFO")); err != nil {
		t.Fatalf("os.Symlink: %v", err)
	}

	purgeLogsOnce(now, logDir, "adam", 0, 1*time.Hour)

	left, err := filepath.Glob(path.Join(logDir, "adam.*"))
	if err != nil {
		t.Fatalf("filepath.Glob: %v", err)
	}

	if len(left) != 3 {
		// 1. 113000 is within 1 hour
		// 2. 100000 is current (vtadam.INFO)
		// 3. vtadam.INFO symlink remains
		// rest are removed
		t.Errorf("wrong number of files remain: want %v, got %v", 3, len(left))
	}
}
