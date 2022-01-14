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
	"io"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"testing"
)

func TestConsoleLogger(t *testing.T) {
	testConsoleLogger(t, false, "TestConsoleLogger")
}

func TestTeeConsoleLogger(t *testing.T) {
	testConsoleLogger(t, true, "TestTeeConsoleLogger")
}

func testConsoleLogger(t *testing.T, tee bool, entrypoint string) {
	if os.Getenv("TEST_CONSOLE_LOGGER") == "1" {
		// Generate output in subprocess.
		var logger Logger
		if tee {
			logger = NewTeeLogger(NewConsoleLogger(), NewMemoryLogger())
		} else {
			logger = NewConsoleLogger()
		}
		// Add 'tee' to the output to make sure we've
		// called the right method in the subprocess.
		logger.Infof("info %v %v", 1, tee)
		logger.Warningf("warning %v %v", 2, tee)
		logger.Errorf("error %v %v", 3, tee)
		return
	}

	// Run subprocess and collect console output.
	cmd := exec.Command(os.Args[0], "-test.run=^"+entrypoint+"$", "-logtostderr")
	cmd.Env = append(os.Environ(), "TEST_CONSOLE_LOGGER=1")
	stderr, err := cmd.StderrPipe()
	if err != nil {
		t.Fatalf("cmd.StderrPipe() error: %v", err)
	}
	if err := cmd.Start(); err != nil {
		t.Fatalf("cmd.Start() error: %v", err)
	}
	out, err := io.ReadAll(stderr)
	if err != nil {
		t.Fatalf("io.ReadAll(sterr) error: %v", err)
	}
	if err := cmd.Wait(); err != nil {
		t.Fatalf("cmd.Wait() error: %v", err)
	}

	// Check output. Filter out entries that are not from console_logger_test.go
	lines := strings.Split(string(out), "\n")
	gotlines := []string{}
	for _, line := range lines {
		if strings.Contains(line, "console_logger_test.go") {
			gotlines = append(gotlines, line)
		}
	}
	wantlines := []string{
		fmt.Sprintf("^I.*info 1 %v$", tee),
		fmt.Sprintf("^W.*warning 2 %v$", tee),
		fmt.Sprintf("^E.*error 3 %v$", tee),
	}
	for i, want := range wantlines {
		got := gotlines[i]
		match, err := regexp.MatchString(want, got)
		if err != nil {
			t.Errorf("regexp.MatchString error: %v", err)
		}
		if !match {
			t.Errorf("got %q, want %q", got, want)
		}
	}
}
