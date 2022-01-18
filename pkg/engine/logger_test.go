package engine

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
	"testing"
	"time"
)

func TestFormatHeader_0(t *testing.T) {
	tm := time.Now()
	if tm.Year() >= 2100 {
		t.FailNow()
	}
	dur := time.Second
	for {
		if tm.Year() >= 2100 {
			break
		}
		h, _, _ := formatTimeHeader(tm)
		if tm.Format("2018/03/26 15:04:05.000 ") != string(h) {
			t.Log(tm)
			t.FailNow()
		}
		tm = tm.Add(dur)
		dur *= 2
	}
}

func TestFormatHeader_1(t *testing.T) {
	tm := time.Now()
	year := tm.Year()
	dur := time.Second
	for {
		if tm.Year() >= year+1 {
			break
		}
		h, _, _ := formatTimeHeader(tm)
		if tm.Format("2018/03/26 15:04:05.000 ") != string(h) {
			t.Log(tm)
			t.FailNow()
		}
		tm = tm.Add(dur)
	}
}
