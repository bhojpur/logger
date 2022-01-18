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
	"path"
	"strconv"
)

var formatterMap = make(map[string]LogFormatter, 4)

type LogFormatter interface {
	Format(lm *LogMsg) string
}

// PatternLogFormatter provides a quick format method
// for example:
// tes := &PatternLogFormatter{Pattern: "%F:%n|%w %t>> %m", WhenFormat: "2006-01-02"}
// RegisterFormatter("tes", tes)
// SetGlobalFormatter("tes")
type PatternLogFormatter struct {
	Pattern    string
	WhenFormat string
}

func (p *PatternLogFormatter) getWhenFormatter() string {
	s := p.WhenFormat
	if s == "" {
		s = "2006/01/02 15:04:05.123" // default style
	}
	return s
}

func (p *PatternLogFormatter) Format(lm *LogMsg) string {
	return p.ToString(lm)
}

// RegisterFormatter register an formatter. Usually you should use this to extend your custom formatter
// for example:
// RegisterFormatter("my-fmt", &MyFormatter{})
// logs.SetFormatter(Console, `{"formatter": "my-fmt"}`)
func RegisterFormatter(name string, fmtr LogFormatter) {
	formatterMap[name] = fmtr
}

func GetFormatter(name string) (LogFormatter, bool) {
	res, ok := formatterMap[name]
	return res, ok
}

// 'w' when, 'm' msg,'f' filename，'F' full path，'n' line number
// 'l' level number, 't' prefix of level type, 'T' full name of level type
func (p *PatternLogFormatter) ToString(lm *LogMsg) string {
	s := []rune(p.Pattern)
	m := map[rune]string{
		'w': lm.When.Format(p.getWhenFormatter()),
		'm': lm.Msg,
		'n': strconv.Itoa(lm.LineNumber),
		'l': strconv.Itoa(lm.Level),
		't': levelPrefix[lm.Level-1],
		'T': levelNames[lm.Level-1],
		'F': lm.FilePath,
	}
	_, m['f'] = path.Split(lm.FilePath)
	res := ""
	for i := 0; i < len(s)-1; i++ {
		if s[i] == '%' {
			if k, ok := m[s[i+1]]; ok {
				res += k
				i++
				continue
			}
		}
		res += string(s[i])
	}
	return res
}
