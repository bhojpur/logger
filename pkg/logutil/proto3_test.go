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
	"math"
	"testing"
	"time"

	"google.golang.org/protobuf/proto"

	loggerpb "github.com/bhojpur/logger/pkg/api/v1"
)

const (
	// Seconds field of the earliest valid Timestamp.
	// This is time.Date(1, 1, 1, 0, 0, 0, 0, time.UTC).Unix().
	minValidSeconds = -62135596800
	// Seconds field just after the latest valid Timestamp.
	// This is time.Date(10000, 1, 1, 0, 0, 0, 0, time.UTC).Unix().
	maxValidSeconds = 253402300800
)

func utcDate(year, month, day int) time.Time {
	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
}

var tests = []struct {
	pt *loggerpb.Time
	t  time.Time
}{
	// The timestamp representing the Unix epoch date.
	{pt: &loggerpb.Time{Seconds: 0, Nanoseconds: 0},
		t: utcDate(1970, 1, 1)},

	// The smallest representable timestamp with non-negative nanos.
	{pt: &loggerpb.Time{Seconds: math.MinInt64, Nanoseconds: 0},
		t: time.Unix(math.MinInt64, 0).UTC()},

	// The earliest valid timestamp.
	{pt: &loggerpb.Time{Seconds: minValidSeconds, Nanoseconds: 0},
		t: utcDate(1, 1, 1)},

	// The largest representable timestamp with nanos in range.
	{pt: &loggerpb.Time{Seconds: math.MaxInt64, Nanoseconds: 1e9 - 1},
		t: time.Unix(math.MaxInt64, 1e9-1).UTC()},

	// The largest valid timestamp.
	{pt: &loggerpb.Time{Seconds: maxValidSeconds - 1, Nanoseconds: 1e9 - 1},
		t: time.Date(9999, 12, 31, 23, 59, 59, 1e9-1, time.UTC)},

	// The smallest invalid timestamp that is larger than the valid range.
	{pt: &loggerpb.Time{Seconds: maxValidSeconds, Nanoseconds: 0},
		t: time.Unix(maxValidSeconds, 0).UTC()},

	// A date before the epoch.
	{pt: &loggerpb.Time{Seconds: -281836800, Nanoseconds: 0},
		t: utcDate(1961, 1, 26)},

	// A date after the epoch.
	{pt: &loggerpb.Time{Seconds: 1296000000, Nanoseconds: 0},
		t: utcDate(2011, 1, 26)},

	// A date after the epoch, in the middle of the day.
	{pt: &loggerpb.Time{Seconds: 1296012345, Nanoseconds: 940483},
		t: time.Date(2011, 1, 26, 3, 25, 45, 940483, time.UTC)},
}

func TestProtoToTime(t *testing.T) {
	for i, s := range tests {
		got := ProtoToTime(s.pt)
		if got != s.t {
			t.Errorf("ProtoToTime[%v](%v) = %v, want %v", i, s.pt, got, s.t)
		}
	}
}

func TestTimeToProto(t *testing.T) {
	for i, s := range tests {
		got := TimeToProto(s.t)
		if !proto.Equal(got, s.pt) {
			t.Errorf("TimeToProto[%v](%v) = %v, want %v", i, s.t, got, s.pt)
		}
	}
}
