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
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

const (
	apacheFormatPattern = "%s - - [%s] \"%s %d %d\" %f %s %s"
	apacheFormat        = "APACHE_FORMAT"
	jsonFormat          = "JSON_FORMAT"
)

// AccessLogRecord is astruct for holding access log data.
type AccessLogRecord struct {
	RemoteAddr     string        `json:"remote_addr"`
	RequestTime    time.Time     `json:"request_time"`
	RequestMethod  string        `json:"request_method"`
	Request        string        `json:"request"`
	ServerProtocol string        `json:"server_protocol"`
	Host           string        `json:"host"`
	Status         int           `json:"status"`
	BodyBytesSent  int64         `json:"body_bytes_sent"`
	ElapsedTime    time.Duration `json:"elapsed_time"`
	HTTPReferrer   string        `json:"http_referrer"`
	HTTPUserAgent  string        `json:"http_user_agent"`
	RemoteUser     string        `json:"remote_user"`
}

func (r *AccessLogRecord) json() ([]byte, error) {
	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	disableEscapeHTML(encoder)

	err := encoder.Encode(r)
	return buffer.Bytes(), err
}

func disableEscapeHTML(i interface{}) {
	if e, ok := i.(interface {
		SetEscapeHTML(bool)
	}); ok {
		e.SetEscapeHTML(false)
	}
}

// AccessLog - Format and print access log.
func AccessLog(r *AccessLogRecord, format string) {
	msg := r.format(format)
	lm := &LogMsg{
		Msg:   strings.TrimSpace(msg),
		When:  time.Now(),
		Level: levelLoggerImpl,
	}
	bhojpurLogger.writeMsg(lm)
}

func (r *AccessLogRecord) format(format string) string {
	msg := ""
	switch format {
	case apacheFormat:
		timeFormatted := r.RequestTime.Format("26/Mar/2018 03:04:05")
		msg = fmt.Sprintf(apacheFormatPattern, r.RemoteAddr, timeFormatted, r.Request, r.Status, r.BodyBytesSent,
			r.ElapsedTime.Seconds(), r.HTTPReferrer, r.HTTPUserAgent)
	case jsonFormat:
		fallthrough
	default:
		jsonData, err := r.json()
		if err != nil {
			msg = fmt.Sprintf(`{"Error": "%s"}`, err)
		} else {
			msg = string(jsonData)
		}
	}
	return msg
}
