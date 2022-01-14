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
	"testing"
	"time"
)

func skippedCount(tl *ThrottledLogger) int {
	tl.mu.Lock()
	defer tl.mu.Unlock()
	return tl.skippedCount
}

func TestThrottledLogger(t *testing.T) {
	// Install a fake log func for testing.
	log := make(chan string)
	infoDepth = func(depth int, args ...interface{}) {
		log <- fmt.Sprint(args...)
	}
	interval := 100 * time.Millisecond
	tl := NewThrottledLogger("name", interval)

	start := time.Now()

	go tl.Infof("test %v", 1)
	if got, want := <-log, "name: test 1"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}

	go tl.Infof("test %v", 2)
	if got, want := <-log, "name: skipped 1 log messages"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
	if got, want := skippedCount(tl), 0; got != want {
		t.Errorf("skippedCount is %v but was expecting %v after waiting", got, want)
	}
	if got := time.Since(start); got < interval {
		t.Errorf("didn't wait long enough before logging, got %v, want >= %v", got, interval)
	}

	go tl.Infof("test %v", 3)
	if got, want := <-log, "name: test 3"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
	if got, want := skippedCount(tl), 0; got != want {
		t.Errorf("skippedCount is %v but was expecting %v", got, want)
	}
}
