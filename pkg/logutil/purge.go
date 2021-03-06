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
	"flag"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

var (
	keepLogsByCtime   = flag.Duration("keep_logs", 0, "keep logs for this long (using ctime) (zero to keep forever)")
	keepLogsByMtime   = flag.Duration("keep_logs_by_mtime", 0, "keep logs for this long (using mtime) (zero to keep forever)")
	purgeLogsInterval = flag.Duration("purge_logs_interval", 1*time.Hour, "how often try to remove old logs")
)

// parse parses a file name (as used by glog) and returns its process
// name and timestamp.
func parseCreatedTimestamp(filename string) (timestamp time.Time, err error) {
	parts := strings.Split(filepath.Base(filename), ".")
	if len(parts) < 6 {
		return time.Time{}, fmt.Errorf("malformed logfile name: %v", filename)
	}
	return time.ParseInLocation("20060102-150405", parts[len(parts)-2], time.Now().Location())

}

func getModifiedTimestamp(filename string) (timestamp time.Time, err error) {
	fileInfo, err := os.Stat(filename)
	if err != nil {
		return time.Time{}, err
	}
	return fileInfo.ModTime(), nil
}

var levels = []string{"INFO", "ERROR", "WARNING", "FATAL"}

// purgeLogsOnce removes logfiles for program for dir, if their age
// relative to now is greater than [cm]timeDelta
func purgeLogsOnce(now time.Time, dir, program string, ctimeDelta time.Duration, mtimeDelta time.Duration) {
	current := make(map[string]bool)
	for _, level := range levels {
		c, err := filepath.EvalSymlinks(path.Join(dir, fmt.Sprintf("%s.%s", program, level)))
		if err != nil {
			continue
		}
		current[c] = true
	}

	files, err := filepath.Glob(path.Join(dir, fmt.Sprintf("%s.*", program)))
	if err != nil {
		return
	}
	for _, file := range files {
		statInfo, err := os.Lstat(file)
		if err != nil {
			// Failed to stat file
			continue
		}
		if current[file] || !statInfo.Mode().IsRegular() {
			// Do not purge current file or any non-regular files (symlinks etc)
			continue
		}
		purgeFile := false
		if ctimeDelta != 0 {
			createdTs, err := parseCreatedTimestamp(file)
			if err != nil {
				continue
			}
			purgeFile = purgeFile || now.Sub(createdTs) > ctimeDelta
		}
		if mtimeDelta != 0 {
			modifiedTs, err := getModifiedTimestamp(file)
			if err != nil {
				continue
			}
			purgeFile = purgeFile || now.Sub(modifiedTs) > mtimeDelta
		}
		if purgeFile {
			os.Remove(file)
		}
	}
}

// PurgeLogs removes any log files that were started more than
// keepLogs ago and that aren't the current log.
func PurgeLogs() {
	f := flag.Lookup("log_dir")
	if f == nil {
		panic("the logging module doesn't specify a log_dir flag")
	}
	if *keepLogsByCtime == 0 && *keepLogsByMtime == 0 {
		return
	}
	logDir := f.Value.String()
	program := filepath.Base(os.Args[0])
	ticker := time.NewTicker(*purgeLogsInterval)
	for range ticker.C {
		purgeLogsOnce(time.Now(), logDir, program, *keepLogsByCtime, *keepLogsByMtime)
	}
}
