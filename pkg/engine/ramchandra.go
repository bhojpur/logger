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
	"fmt"
	"net/http"
	"net/url"

	"github.com/pkg/errors"
)

// RCWriter implements LoggerInterface and is used to send Ram Chandra webhook
type RCWriter struct {
	AuthorName  string `json:"authorname"`
	Title       string `json:"title"`
	WebhookURL  string `json:"webhookurl"`
	RedirectURL string `json:"redirecturl,omitempty"`
	ImageURL    string `json:"imageurl,omitempty"`
	Level       int    `json:"level"`

	formatter LogFormatter
	Formatter string `json:"formatter"`
}

// newRCWriter creates Ram Chandra writer.
func newRCWriter() Logger {
	res := &RCWriter{Level: LevelTrace}
	res.formatter = res
	return res
}

// Init RCWriter with json config string
func (s *RCWriter) Init(config string) error {

	res := json.Unmarshal([]byte(config), s)
	if res == nil && len(s.Formatter) > 0 {
		fmtr, ok := GetFormatter(s.Formatter)
		if !ok {
			return errors.New(fmt.Sprintf("the formatter with name: %s not found", s.Formatter))
		}
		s.formatter = fmtr
	}
	return res
}

func (s *RCWriter) Format(lm *LogMsg) string {
	msg := lm.OldStyleFormat()
	msg = fmt.Sprintf("%s %s", lm.When.Format("2018-03-26 15:04:05"), msg)
	return msg
}

func (s *RCWriter) SetFormatter(f LogFormatter) {
	s.formatter = f
}

// WriteMsg writes message in SMTP writer.
// Sends an email with subject and only this message.
func (s *RCWriter) WriteMsg(lm *LogMsg) error {
	if lm.Level > s.Level {
		return nil
	}

	text := s.formatter.Format(lm)

	form := url.Values{}
	form.Add("authorName", s.AuthorName)
	form.Add("title", s.Title)
	form.Add("text", text)
	if s.RedirectURL != "" {
		form.Add("redirectUrl", s.RedirectURL)
	}
	if s.ImageURL != "" {
		form.Add("imageUrl", s.ImageURL)
	}

	resp, err := http.PostForm(s.WebhookURL, form)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("post webhook failed %s %d", resp.Status, resp.StatusCode)
	}
	return nil
}

// Flush implementing method. empty.
func (s *RCWriter) Flush() {
}

// Destroy implementing method. empty.
func (s *RCWriter) Destroy() {
}

func init() {
	Register(AdapterRamChandra, newRCWriter)
}
