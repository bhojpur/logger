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
	"net"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// ConnTCPListener takes a TCP listener and accepts n TCP connections
// Returns connections using connChan
func connTCPListener(t *testing.T, n int, ln net.Listener, connChan chan<- net.Conn) {

	// Listen and accept n incoming connections
	for i := 0; i < n; i++ {
		conn, err := ln.Accept()
		if err != nil {
			t.Log("Error accepting connection: ", err.Error())
			os.Exit(1)
		}

		// Send accepted connection to channel
		connChan <- conn
	}
	ln.Close()
	close(connChan)
}

func TestConn(t *testing.T) {
	log := NewLogger(1000)
	log.SetLogger("conn", `{"net":"tcp","addr":":7020"}`)
	log.Informational("informational")
}

// need to rewrite this test, it's not stable
func TestReconnect(t *testing.T) {
	// Setup connection listener
	newConns := make(chan net.Conn)
	connNum := 2
	ln, err := net.Listen("tcp", ":6002")
	if err != nil {
		t.Log("Error listening:", err.Error())
		os.Exit(1)
	}
	go connTCPListener(t, connNum, ln, newConns)

	// Setup logger
	log := NewLogger(1000)
	log.SetPrefix("test")
	log.SetLogger(AdapterConn, `{"net":"tcp","reconnect":true,"level":6,"addr":":6002"}`)
	log.Informational("informational 1")

	// Refuse first connection
	first := <-newConns
	first.Close()

	// Send another log after conn closed
	log.Informational("informational 2")

	// Check if there was a second connection attempt
	select {
	case second := <-newConns:
		second.Close()
	default:
		t.Error("Did not reconnect")
	}
}

func TestConnWriter_Format(t *testing.T) {
	lg := &LogMsg{
		Level:      LevelDebug,
		Msg:        "Hello, world",
		When:       time.Date(2020, 9, 19, 20, 12, 37, 9, time.UTC),
		FilePath:   "/user/home/main.go",
		LineNumber: 13,
		Prefix:     "Cus",
	}
	cw := NewConn().(*connWriter)
	res := cw.Format(lg)
	assert.Equal(t, "[D] Cus Hello, world", res)
}
