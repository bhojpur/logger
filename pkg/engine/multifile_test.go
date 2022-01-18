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
	"bufio"
	"os"
	"strconv"
	"strings"
	"testing"
)

func TestFiles_1(t *testing.T) {
	log := NewLogger(10000)
	log.SetLogger("multifile", `{"filename":"test.log","separate":["emergency", "alert", "critical", "error", "warning", "notice", "info", "debug"]}`)
	log.Debug("debug")
	log.Informational("info")
	log.Notice("notice")
	log.Warning("warning")
	log.Error("error")
	log.Alert("alert")
	log.Critical("critical")
	log.Emergency("emergency")
	fns := []string{""}
	fns = append(fns, levelNames[0:]...)
	name := "test"
	suffix := ".log"
	for _, fn := range fns {

		file := name + suffix
		if fn != "" {
			file = name + "." + fn + suffix
		}
		f, err := os.Open(file)
		if err != nil {
			t.Fatal(err)
		}
		b := bufio.NewReader(f)
		lineNum := 0
		lastLine := ""
		for {
			line, _, err := b.ReadLine()
			if err != nil {
				break
			}
			if len(line) > 0 {
				lastLine = string(line)
				lineNum++
			}
		}
		var expected = 1
		if fn == "" {
			expected = LevelDebug + 1
		}
		if lineNum != expected {
			t.Fatal(file, "has", lineNum, "lines not "+strconv.Itoa(expected)+" lines")
		}
		if lineNum == 1 {
			if !strings.Contains(lastLine, fn) {
				t.Fatal(file + " " + lastLine + " not contains the log msg " + fn)
			}
		}
		os.Remove(file)
	}

}
