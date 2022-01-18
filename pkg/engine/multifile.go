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
)

// A filesLogWriter manages several fileLogWriter
// filesLogWriter will write logs to the file in json configuration  and write the same level log to correspond file
// means if the file name in configuration is project.log filesLogWriter will create project.error.log/project.debug.log
// and write the error-level logs to project.error.log and write the debug-level logs to project.debug.log
// the rotate attribute also  acts like fileLogWriter
type multiFileLogWriter struct {
	writers       [LevelDebug + 1 + 1]*fileLogWriter // the last one for fullLogWriter
	fullLogWriter *fileLogWriter
	Separate      []string `json:"separate"`
}

var levelNames = [...]string{"emergency", "alert", "critical", "error", "warning", "notice", "info", "debug"}

// Init file logger with json config.
// jsonConfig like:
//	{
//	"filename":"logs/bhojpur.log",
//	"maxLines":0,
//	"maxsize":0,
//	"daily":true,
//	"maxDays":15,
//	"rotate":true,
//  	"perm":0600,
//	"separate":["emergency", "alert", "critical", "error", "warning", "notice", "info", "debug"],
//	}

func (f *multiFileLogWriter) Init(config string) error {

	writer := newFileWriter().(*fileLogWriter)
	err := writer.Init(config)
	if err != nil {
		return err
	}
	f.fullLogWriter = writer
	f.writers[LevelDebug+1] = writer

	// unmarshal "separate" field to f.Separate
	err = json.Unmarshal([]byte(config), f)
	if err != nil {
		return err
	}

	jsonMap := map[string]interface{}{}
	err = json.Unmarshal([]byte(config), &jsonMap)
	if err != nil {
		return err
	}

	for i := LevelEmergency; i < LevelDebug+1; i++ {
		for _, v := range f.Separate {
			if v == levelNames[i] {
				jsonMap["filename"] = f.fullLogWriter.fileNameOnly + "." + levelNames[i] + f.fullLogWriter.suffix
				jsonMap["level"] = i
				bs, _ := json.Marshal(jsonMap)
				writer = newFileWriter().(*fileLogWriter)
				err := writer.Init(string(bs))
				if err != nil {
					return err
				}
				f.writers[i] = writer
			}
		}
	}
	return nil
}

func (f *multiFileLogWriter) Format(lm *LogMsg) string {
	return lm.OldStyleFormat()
}

func (f *multiFileLogWriter) SetFormatter(fmt LogFormatter) {
	f.fullLogWriter.SetFormatter(f)
}

func (f *multiFileLogWriter) Destroy() {
	for i := 0; i < len(f.writers); i++ {
		if f.writers[i] != nil {
			f.writers[i].Destroy()
		}
	}
}

func (f *multiFileLogWriter) WriteMsg(lm *LogMsg) error {
	if f.fullLogWriter != nil {
		f.fullLogWriter.WriteMsg(lm)
	}
	for i := 0; i < len(f.writers)-1; i++ {
		if f.writers[i] != nil {
			if lm.Level == f.writers[i].Level {
				f.writers[i].WriteMsg(lm)
			}
		}
	}
	return nil
}

func (f *multiFileLogWriter) Flush() {
	for i := 0; i < len(f.writers); i++ {
		if f.writers[i] != nil {
			f.writers[i].Flush()
		}
	}
}

// newFilesWriter create a FileLogWriter returning as LoggerInterface.
func newFilesWriter() Logger {
	res := &multiFileLogWriter{}
	return res
}

func init() {
	Register(AdapterMultiFile, newFilesWriter)
}
