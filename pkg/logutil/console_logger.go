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
	"fmt"

	"github.com/bhojpur/events/pkg/log"
)

// ConsoleLogger is a Logger that uses glog directly to log, at the right level.
//
// Note that methods on ConsoleLogger must use pointer receivers,
// because otherwise an auto-generated conversion method will be inserted in the
// call stack when ConsoleLogger is used via TeeLogger, making the log depth
// incorrect.
type ConsoleLogger struct{}

// NewConsoleLogger returns a simple ConsoleLogger.
func NewConsoleLogger() *ConsoleLogger {
	return &ConsoleLogger{}
}

// Infof is part of the Logger interface
func (cl *ConsoleLogger) Infof(format string, v ...interface{}) {
	cl.InfoDepth(1, fmt.Sprintf(format, v...))
}

// Warningf is part of the Logger interface
func (cl *ConsoleLogger) Warningf(format string, v ...interface{}) {
	cl.WarningDepth(1, fmt.Sprintf(format, v...))
}

// Errorf is part of the Logger interface
func (cl *ConsoleLogger) Errorf(format string, v ...interface{}) {
	cl.ErrorDepth(1, fmt.Sprintf(format, v...))
}

// Errorf2 is part of the Logger interface
func (cl *ConsoleLogger) Errorf2(err error, format string, v ...interface{}) {
	cl.ErrorDepth(1, fmt.Sprintf(format+": %+v", append(v, err)))
}

// Error is part of the Logger interface
func (cl *ConsoleLogger) Error(err error) {
	cl.ErrorDepth(1, fmt.Sprintf("%+v", err))
}

// Printf is part of the Logger interface
func (cl *ConsoleLogger) Printf(format string, v ...interface{}) {
	fmt.Printf(format, v...)
}

// InfoDepth is part of the Logger interface.
func (cl *ConsoleLogger) InfoDepth(depth int, s string) {
	log.InfoDepth(1+depth, s)
}

// WarningDepth is part of the Logger interface.
func (cl *ConsoleLogger) WarningDepth(depth int, s string) {
	log.WarningDepth(1+depth, s)
}

// ErrorDepth is part of the Logger interface.
func (cl *ConsoleLogger) ErrorDepth(depth int, s string) {
	log.ErrorDepth(1+depth, s)
}
