package es

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
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/elastic/go-elasticsearch/v6"
	"github.com/elastic/go-elasticsearch/v6/esapi"

	logs "github.com/bhojpur/logger/pkg/engine"
)

// NewES returns a LoggerInterface
func NewES() logs.Logger {
	cw := &esLogger{
		Level:       logs.LevelDebug,
		indexNaming: indexNaming,
	}
	return cw
}

// esLogger will log msg into ES
// before you using this implementation,
// please import this package
// usually means that you can import this package in your main package
// for example, anonymous:
// import _ "github.com/bhojpur/logger/pkg/engine/es"
type esLogger struct {
	*elasticsearch.Client
	DSN       string `json:"dsn"`
	Level     int    `json:"level"`
	formatter logs.LogFormatter
	Formatter string `json:"formatter"`

	indexNaming IndexNaming
}

func (el *esLogger) Format(lm *logs.LogMsg) string {

	msg := lm.OldStyleFormat()
	idx := LogDocument{
		Timestamp: lm.When.Format(time.RFC3339),
		Msg:       msg,
	}
	body, err := json.Marshal(idx)
	if err != nil {
		return msg
	}
	return string(body)
}

func (el *esLogger) SetFormatter(f logs.LogFormatter) {
	el.formatter = f
}

// {"dsn":"http://localhost:9200/","level":1}
func (el *esLogger) Init(config string) error {

	err := json.Unmarshal([]byte(config), el)
	if err != nil {
		return err
	}
	if el.DSN == "" {
		return errors.New("empty dsn")
	} else if u, err := url.Parse(el.DSN); err != nil {
		return err
	} else if u.Path == "" {
		return errors.New("missing prefix")
	} else {
		conn, err := elasticsearch.NewClient(elasticsearch.Config{
			Addresses: []string{el.DSN},
		})
		if err != nil {
			return err
		}
		el.Client = conn
	}
	if len(el.Formatter) > 0 {
		fmtr, ok := logs.GetFormatter(el.Formatter)
		if !ok {
			return errors.New(fmt.Sprintf("the formatter with name: %s not found", el.Formatter))
		}
		el.formatter = fmtr
	}
	return nil
}

// WriteMsg writes the msg and level into es
func (el *esLogger) WriteMsg(lm *logs.LogMsg) error {
	if lm.Level > el.Level {
		return nil
	}

	msg := el.formatter.Format(lm)

	req := esapi.IndexRequest{
		Index:        indexNaming.IndexName(lm),
		DocumentType: "logs",
		Body:         strings.NewReader(msg),
	}
	_, err := req.Do(context.Background(), el.Client)
	return err
}

// Destroy is a empty method
func (el *esLogger) Destroy() {
}

// Flush is a empty method
func (el *esLogger) Flush() {

}

type LogDocument struct {
	Timestamp string `json:"timestamp"`
	Msg       string `json:"msg"`
}

func init() {
	logs.Register(logs.AdapterEs, NewES)
}
