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
	"io"
	"net"

	"github.com/pkg/errors"
)

// connWriter implements LoggerInterface.
// Writes messages in keep-live tcp connection.
type connWriter struct {
	lg             *logWriter
	innerWriter    io.WriteCloser
	formatter      LogFormatter
	Formatter      string `json:"formatter"`
	ReconnectOnMsg bool   `json:"reconnectOnMsg"`
	Reconnect      bool   `json:"reconnect"`
	Net            string `json:"net"`
	Addr           string `json:"addr"`
	Level          int    `json:"level"`
}

// NewConn creates new ConnWrite returning as LoggerInterface.
func NewConn() Logger {
	conn := new(connWriter)
	conn.Level = LevelTrace
	conn.formatter = conn
	return conn
}

func (c *connWriter) Format(lm *LogMsg) string {
	return lm.OldStyleFormat()
}

// Init initializes a connection writer with json config.
// json config only needs they "level" key
func (c *connWriter) Init(config string) error {
	res := json.Unmarshal([]byte(config), c)
	if res == nil && len(c.Formatter) > 0 {
		fmtr, ok := GetFormatter(c.Formatter)
		if !ok {
			return errors.New(fmt.Sprintf("the formatter with name: %s not found", c.Formatter))
		}
		c.formatter = fmtr
	}
	return res
}

func (c *connWriter) SetFormatter(f LogFormatter) {
	c.formatter = f
}

// WriteMsg writes message in connection.
// If connection is down, try to re-connect.
func (c *connWriter) WriteMsg(lm *LogMsg) error {
	if lm.Level > c.Level {
		return nil
	}
	if c.needToConnectOnMsg() {
		err := c.connect()
		if err != nil {
			return err
		}
	}

	if c.ReconnectOnMsg {
		defer c.innerWriter.Close()
	}

	msg := c.formatter.Format(lm)

	_, err := c.lg.writeln(msg)
	if err != nil {
		return err
	}
	return nil
}

// Flush implementing method. empty.
func (c *connWriter) Flush() {

}

// Destroy destroy connection writer and close tcp listener.
func (c *connWriter) Destroy() {
	if c.innerWriter != nil {
		c.innerWriter.Close()
	}
}

func (c *connWriter) connect() error {
	if c.innerWriter != nil {
		c.innerWriter.Close()
		c.innerWriter = nil
	}

	conn, err := net.Dial(c.Net, c.Addr)
	if err != nil {
		return err
	}

	if tcpConn, ok := conn.(*net.TCPConn); ok {
		tcpConn.SetKeepAlive(true)
	}

	c.innerWriter = conn
	c.lg = newLogWriter(conn)
	return nil
}

func (c *connWriter) needToConnectOnMsg() bool {
	if c.Reconnect {
		return true
	}

	if c.innerWriter == nil {
		return true
	}

	return c.ReconnectOnMsg
}

func init() {
	Register(AdapterConn, NewConn)
}
