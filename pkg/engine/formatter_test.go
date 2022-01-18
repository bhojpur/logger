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
	"encoding/json"
	"errors"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type CustomFormatter struct{}

func (c *CustomFormatter) Format(lm *LogMsg) string {
	return "hello, msg: " + lm.Msg
}

type TestLogger struct {
	Formatter string `json:"formatter"`
	Expected  string
	formatter LogFormatter
}

func (t *TestLogger) Init(config string) error {
	er := json.Unmarshal([]byte(config), t)
	t.formatter, _ = GetFormatter(t.Formatter)
	return er
}

func (t *TestLogger) WriteMsg(lm *LogMsg) error {
	msg := t.formatter.Format(lm)
	if msg != t.Expected {
		return errors.New("not equal")
	}
	return nil
}

func (t *TestLogger) Destroy() {
	panic("implement me")
}

func (t *TestLogger) Flush() {
	panic("implement me")
}

func (t *TestLogger) SetFormatter(f LogFormatter) {
	panic("implement me")
}

func TestCustomFormatter(t *testing.T) {
	RegisterFormatter("custom", &CustomFormatter{})
	tl := &TestLogger{
		Expected: "hello, msg: world",
	}
	assert.Nil(t, tl.Init(`{"formatter": "custom"}`))
	assert.Nil(t, tl.WriteMsg(&LogMsg{
		Msg: "world",
	}))
}

func TestPatternLogFormatter(t *testing.T) {
	tes := &PatternLogFormatter{
		Pattern:    "%F:%n|%w%t>> %m",
		WhenFormat: "2006-01-02",
	}
	when := time.Now()
	lm := &LogMsg{
		Msg:        "message",
		FilePath:   "/User/go/bhojpur/main.go",
		Level:      LevelWarn,
		LineNumber: 10,
		When:       when,
	}
	got := tes.ToString(lm)
	want := lm.FilePath + ":" + strconv.Itoa(lm.LineNumber) + "|" +
		when.Format(tes.WhenFormat) + levelPrefix[lm.Level-1] + ">> " + lm.Msg
	if got != want {
		t.Errorf("want %s, got %s", want, got)
	}
}
