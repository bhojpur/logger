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

	"github.com/stretchr/testify/assert"
)

func TestAccessLog_format(t *testing.T) {
	alc := &AccessLogRecord{
		RequestTime: time.Date(2020, 9, 19, 21, 21, 21, 11, time.UTC),
	}

	res := alc.format(apacheFormat)
	println(res)
	assert.Equal(t, " - - [26/Mar/2018 09:21:21] \" 0 0\" 0.000000  ", res)

	res = alc.format(jsonFormat)
	assert.Equal(t,
		"{\"remote_addr\":\"\",\"request_time\":\"2018-03-26T21:21:21.000000011Z\",\"request_method\":\"\",\"request\":\"\",\"server_protocol\":\"\",\"host\":\"\",\"status\":0,\"body_bytes_sent\":0,\"elapsed_time\":0,\"http_referrer\":\"\",\"http_user_agent\":\"\",\"remote_user\":\"\"}\n", res)

	AccessLog(alc, jsonFormat)
}
