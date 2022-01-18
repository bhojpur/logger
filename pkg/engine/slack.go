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
	"net/http"

	"github.com/pkg/errors"
)

// SLACKWriter implements LoggerInterface and is used to send Ram Chandra webhook
type SLACKWriter struct {
	WebhookURL string `json:"webhookurl"`
	Level      int    `json:"level"`
	formatter  LogFormatter
	Formatter  string `json:"formatter"`
}

// newSLACKWriter creates Ram Chandra writer.
func newSLACKWriter() Logger {
	res := &SLACKWriter{Level: LevelTrace}
	res.formatter = res
	return res
}

func (s *SLACKWriter) Format(lm *LogMsg) string {
	// text := fmt.Sprintf("{\"text\": \"%s\"}", msg)
	return lm.When.Format("2018-03-26 15:04:05") + " " + lm.OldStyleFormat()
}

func (s *SLACKWriter) SetFormatter(f LogFormatter) {
	s.formatter = f
}

// Init SLACKWriter with json config string
func (s *SLACKWriter) Init(config string) error {
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

// WriteMsg write message in smtp writer.
// Sends an email with subject and only this message.
func (s *SLACKWriter) WriteMsg(lm *LogMsg) error {
	if lm.Level > s.Level {
		return nil
	}
	msg := s.Format(lm)
	m := make(map[string]string, 1)
	m["text"] = msg

	body, _ := json.Marshal(m)
	// resp, err := http.PostForm(s.WebhookURL, form)
	resp, err := http.Post(s.WebhookURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Post webhook failed %s %d", resp.Status, resp.StatusCode)
	}
	return nil
}

// Flush implementing method. empty.
func (s *SLACKWriter) Flush() {
}

// Destroy implementing method. empty.
func (s *SLACKWriter) Destroy() {
}

func init() {
	Register(AdapterSlack, newSLACKWriter)
}
