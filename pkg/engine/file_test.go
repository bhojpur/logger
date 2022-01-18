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
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFilePerm(t *testing.T) {
	log := NewLogger(10000)
	// use 0666 as test perm cause the default umask is 022
	log.SetLogger("file", `{"filename":"test.log", "perm": "0666"}`)
	log.Debug("debug")
	log.Informational("info")
	log.Notice("notice")
	log.Warning("warning")
	log.Error("error")
	log.Alert("alert")
	log.Critical("critical")
	log.Emergency("emergency")
	file, err := os.Stat("test.log")
	if err != nil {
		t.Fatal(err)
	}
	if file.Mode() != 0666 {
		t.Fatal("unexpected log file permission")
	}
	os.Remove("test.log")
}

func TestFile1(t *testing.T) {
	log := NewLogger(10000)
	log.SetLogger("file", `{"filename":"test.log"}`)
	log.Debug("debug")
	log.Informational("info")
	log.Notice("notice")
	log.Warning("warning")
	log.Error("error")
	log.Alert("alert")
	log.Critical("critical")
	log.Emergency("emergency")
	f, err := os.Open("test.log")
	if err != nil {
		t.Fatal(err)
	}
	b := bufio.NewReader(f)
	lineNum := 0
	for {
		line, _, err := b.ReadLine()
		if err != nil {
			break
		}
		if len(line) > 0 {
			lineNum++
		}
	}
	var expected = LevelDebug + 1
	if lineNum != expected {
		t.Fatal(lineNum, "not "+strconv.Itoa(expected)+" lines")
	}
	os.Remove("test.log")
}

func TestFile2(t *testing.T) {
	log := NewLogger(10000)
	log.SetLogger("file", fmt.Sprintf(`{"filename":"test2.log","level":%d}`, LevelError))
	log.Debug("debug")
	log.Info("info")
	log.Notice("notice")
	log.Warning("warning")
	log.Error("error")
	log.Alert("alert")
	log.Critical("critical")
	log.Emergency("emergency")
	f, err := os.Open("test2.log")
	if err != nil {
		t.Fatal(err)
	}
	b := bufio.NewReader(f)
	lineNum := 0
	for {
		line, _, err := b.ReadLine()
		if err != nil {
			break
		}
		if len(line) > 0 {
			lineNum++
		}
	}
	var expected = LevelError + 1
	if lineNum != expected {
		t.Fatal(lineNum, "not "+strconv.Itoa(expected)+" lines")
	}
	os.Remove("test2.log")
}

func TestFileDailyRotate_01(t *testing.T) {
	log := NewLogger(10000)
	log.SetLogger("file", `{"filename":"test3.log","maxlines":4}`)
	log.Debug("debug")
	log.Info("info")
	log.Notice("notice")
	log.Warning("warning")
	log.Error("error")
	log.Alert("alert")
	log.Critical("critical")
	log.Emergency("emergency")
	rotateName := "test3" + fmt.Sprintf(".%s.%03d", time.Now().Format("2006-01-02"), 1) + ".log"
	b, err := exists(rotateName)
	if !b || err != nil {
		os.Remove("test3.log")
		t.Fatal("rotate not generated")
	}
	os.Remove(rotateName)
	os.Remove("test3.log")
}

func TestFileDailyRotate_02(t *testing.T) {
	fn1 := "rotate_day.log"
	fn2 := "rotate_day." + time.Now().Add(-24*time.Hour).Format("2006-01-02") + ".001.log"
	testFileRotate(t, fn1, fn2, true, false)
}

func TestFileDailyRotate_03(t *testing.T) {
	fn1 := "rotate_day.log"
	fn := "rotate_day." + time.Now().Add(-24*time.Hour).Format("2006-01-02") + ".log"
	os.Create(fn)
	fn2 := "rotate_day." + time.Now().Add(-24*time.Hour).Format("2006-01-02") + ".001.log"
	testFileRotate(t, fn1, fn2, true, false)
	os.Remove(fn)
}

func TestFileDailyRotate_04(t *testing.T) {
	fn1 := "rotate_day.log"
	fn2 := "rotate_day." + time.Now().Add(-24*time.Hour).Format("2006-01-02") + ".001.log"
	testFileDailyRotate(t, fn1, fn2)
}

func TestFileDailyRotate_05(t *testing.T) {
	fn1 := "rotate_day.log"
	fn := "rotate_day." + time.Now().Add(-24*time.Hour).Format("2006-01-02") + ".log"
	os.Create(fn)
	fn2 := "rotate_day." + time.Now().Add(-24*time.Hour).Format("2006-01-02") + ".001.log"
	testFileDailyRotate(t, fn1, fn2)
	os.Remove(fn)
}
func TestFileDailyRotate_06(t *testing.T) { //test file mode
	log := NewLogger(10000)
	log.SetLogger("file", `{"filename":"test3.log","maxlines":4}`)
	log.Debug("debug")
	log.Info("info")
	log.Notice("notice")
	log.Warning("warning")
	log.Error("error")
	log.Alert("alert")
	log.Critical("critical")
	log.Emergency("emergency")
	rotateName := "test3" + fmt.Sprintf(".%s.%03d", time.Now().Format("2006-01-02"), 1) + ".log"
	s, _ := os.Lstat(rotateName)
	if s.Mode() != 0440 {
		os.Remove(rotateName)
		os.Remove("test3.log")
		t.Fatal("rotate file mode error")
	}
	os.Remove(rotateName)
	os.Remove("test3.log")
}

func TestFileHourlyRotate_01(t *testing.T) {
	log := NewLogger(10000)
	log.SetLogger("file", `{"filename":"test3.log","hourly":true,"maxlines":4}`)
	log.Debug("debug")
	log.Info("info")
	log.Notice("notice")
	log.Warning("warning")
	log.Error("error")
	log.Alert("alert")
	log.Critical("critical")
	log.Emergency("emergency")
	rotateName := "test3" + fmt.Sprintf(".%s.%03d", time.Now().Format("2006010215"), 1) + ".log"
	b, err := exists(rotateName)
	if !b || err != nil {
		os.Remove("test3.log")
		t.Fatal("rotate not generated")
	}
	os.Remove(rotateName)
	os.Remove("test3.log")
}

func TestFileHourlyRotate_02(t *testing.T) {
	fn1 := "rotate_hour.log"
	fn2 := "rotate_hour." + time.Now().Add(-1*time.Hour).Format("2006010215") + ".001.log"
	testFileRotate(t, fn1, fn2, false, true)
}

func TestFileHourlyRotate_03(t *testing.T) {
	fn1 := "rotate_hour.log"
	fn := "rotate_hour." + time.Now().Add(-1*time.Hour).Format("2006010215") + ".log"
	os.Create(fn)
	fn2 := "rotate_hour." + time.Now().Add(-1*time.Hour).Format("2006010215") + ".001.log"
	testFileRotate(t, fn1, fn2, false, true)
	os.Remove(fn)
}

func TestFileHourlyRotate_04(t *testing.T) {
	fn1 := "rotate_hour.log"
	fn2 := "rotate_hour." + time.Now().Add(-1*time.Hour).Format("2006010215") + ".001.log"
	testFileHourlyRotate(t, fn1, fn2)
}

func TestFileHourlyRotate_05(t *testing.T) {
	fn1 := "rotate_hour.log"
	fn := "rotate_hour." + time.Now().Add(-1*time.Hour).Format("2006010215") + ".log"
	os.Create(fn)
	fn2 := "rotate_hour." + time.Now().Add(-1*time.Hour).Format("2006010215") + ".001.log"
	testFileHourlyRotate(t, fn1, fn2)
	os.Remove(fn)
}

func TestFileHourlyRotate_06(t *testing.T) { //test file mode
	log := NewLogger(10000)
	log.SetLogger("file", `{"filename":"test3.log", "hourly":true, "maxlines":4}`)
	log.Debug("debug")
	log.Info("info")
	log.Notice("notice")
	log.Warning("warning")
	log.Error("error")
	log.Alert("alert")
	log.Critical("critical")
	log.Emergency("emergency")
	rotateName := "test3" + fmt.Sprintf(".%s.%03d", time.Now().Format("2006010215"), 1) + ".log"
	s, _ := os.Lstat(rotateName)
	if s.Mode() != 0440 {
		os.Remove(rotateName)
		os.Remove("test3.log")
		t.Fatal("rotate file mode error")
	}
	os.Remove(rotateName)
	os.Remove("test3.log")
}

func testFileRotate(t *testing.T, fn1, fn2 string, daily, hourly bool) {
	fw := &fileLogWriter{
		Daily:      daily,
		MaxDays:    7,
		Hourly:     hourly,
		MaxHours:   168,
		Rotate:     true,
		Level:      LevelTrace,
		Perm:       "0660",
		RotatePerm: "0440",
	}
	fw.formatter = fw

	if daily {
		fw.Init(fmt.Sprintf(`{"filename":"%v","maxdays":1}`, fn1))
		fw.dailyOpenTime = time.Now().Add(-24 * time.Hour)
		fw.dailyOpenDate = fw.dailyOpenTime.Day()
	}

	if hourly {
		fw.Init(fmt.Sprintf(`{"filename":"%v","maxhours":1}`, fn1))
		fw.hourlyOpenTime = time.Now().Add(-1 * time.Hour)
		fw.hourlyOpenDate = fw.hourlyOpenTime.Day()
	}
	lm := &LogMsg{
		Msg:   "Test message",
		Level: LevelDebug,
		When:  time.Now(),
	}

	fw.WriteMsg(lm)

	for _, file := range []string{fn1, fn2} {
		_, err := os.Stat(file)
		if err != nil {
			t.Log(err)
			t.FailNow()
		}
		os.Remove(file)
	}
	fw.Destroy()
}

func testFileDailyRotate(t *testing.T, fn1, fn2 string) {
	fw := &fileLogWriter{
		Daily:      true,
		MaxDays:    7,
		Rotate:     true,
		Level:      LevelTrace,
		Perm:       "0660",
		RotatePerm: "0440",
	}
	fw.formatter = fw

	fw.Init(fmt.Sprintf(`{"filename":"%v","maxdays":1}`, fn1))
	fw.dailyOpenTime = time.Now().Add(-24 * time.Hour)
	fw.dailyOpenDate = fw.dailyOpenTime.Day()
	today, _ := time.ParseInLocation("2006-01-02", time.Now().Format("2006-01-02"), fw.dailyOpenTime.Location())
	today = today.Add(-1 * time.Second)
	fw.dailyRotate(today)
	for _, file := range []string{fn1, fn2} {
		_, err := os.Stat(file)
		if err != nil {
			t.FailNow()
		}
		content, err := ioutil.ReadFile(file)
		if err != nil {
			t.FailNow()
		}
		if len(content) > 0 {
			t.FailNow()
		}
		os.Remove(file)
	}
	fw.Destroy()
}

func testFileHourlyRotate(t *testing.T, fn1, fn2 string) {
	fw := &fileLogWriter{
		Hourly:     true,
		MaxHours:   168,
		Rotate:     true,
		Level:      LevelTrace,
		Perm:       "0660",
		RotatePerm: "0440",
	}

	fw.formatter = fw
	fw.Init(fmt.Sprintf(`{"filename":"%v","maxhours":1}`, fn1))
	fw.hourlyOpenTime = time.Now().Add(-1 * time.Hour)
	fw.hourlyOpenDate = fw.hourlyOpenTime.Hour()
	hour, _ := time.ParseInLocation("2006010215", time.Now().Format("2006010215"), fw.hourlyOpenTime.Location())
	hour = hour.Add(-1 * time.Second)
	fw.hourlyRotate(hour)
	for _, file := range []string{fn1, fn2} {
		_, err := os.Stat(file)
		if err != nil {
			t.FailNow()
		}
		content, err := ioutil.ReadFile(file)
		if err != nil {
			t.FailNow()
		}
		if len(content) > 0 {
			t.FailNow()
		}
		os.Remove(file)
	}
	fw.Destroy()
}
func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func BenchmarkFile(b *testing.B) {
	log := NewLogger(100000)
	log.SetLogger("file", `{"filename":"test4.log"}`)
	for i := 0; i < b.N; i++ {
		log.Debug("debug")
	}
	os.Remove("test4.log")
}

func BenchmarkFileAsynchronous(b *testing.B) {
	log := NewLogger(100000)
	log.SetLogger("file", `{"filename":"test4.log"}`)
	log.Async()
	for i := 0; i < b.N; i++ {
		log.Debug("debug")
	}
	os.Remove("test4.log")
}

func BenchmarkFileCallDepth(b *testing.B) {
	log := NewLogger(100000)
	log.SetLogger("file", `{"filename":"test4.log"}`)
	log.EnableFuncCallDepth(true)
	log.SetLogFuncCallDepth(2)
	for i := 0; i < b.N; i++ {
		log.Debug("debug")
	}
	os.Remove("test4.log")
}

func BenchmarkFileAsynchronousCallDepth(b *testing.B) {
	log := NewLogger(100000)
	log.SetLogger("file", `{"filename":"test4.log"}`)
	log.EnableFuncCallDepth(true)
	log.SetLogFuncCallDepth(2)
	log.Async()
	for i := 0; i < b.N; i++ {
		log.Debug("debug")
	}
	os.Remove("test4.log")
}

func BenchmarkFileOnGoroutine(b *testing.B) {
	log := NewLogger(100000)
	log.SetLogger("file", `{"filename":"test4.log"}`)
	for i := 0; i < b.N; i++ {
		go log.Debug("debug")
	}
	os.Remove("test4.log")
}

func TestFileLogWriter_Format(t *testing.T) {
	lg := &LogMsg{
		Level:      LevelDebug,
		Msg:        "Hello, world",
		When:       time.Date(2020, 9, 19, 20, 12, 37, 9, time.UTC),
		FilePath:   "/user/home/main.go",
		LineNumber: 13,
		Prefix:     "Cus",
	}

	fw := newFileWriter().(*fileLogWriter)
	res := fw.Format(lg)
	assert.Equal(t, "2020/09/19 20:12:37.000  [D] Cus Hello, world\n", res)
}
