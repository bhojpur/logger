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

// Usage:
//
// import "github.com/bhojpur/logger/pkg/engine"
//
//	log := NewLogger(10000)
//	log.SetLogger("console", "")
//
//	> the first params stand for how many channel
//
// Use it like this:
//
//	log.Trace("trace")
//	log.Info("info")
//	log.Warn("warning")
//	log.Debug("debug")
//	log.Critical("critical")

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
)

// RFC5424 log message levels.
const (
	LevelEmergency = iota
	LevelAlert
	LevelCritical
	LevelError
	LevelWarning
	LevelNotice
	LevelInformational
	LevelDebug
)

// levelLogLogger is defined to implement log.Logger
// the real log level will be LevelEmergency
const levelLoggerImpl = -1

// Name for adapter with Bhojpur official support
const (
	AdapterConsole    = "console"
	AdapterFile       = "file"
	AdapterMultiFile  = "multifile"
	AdapterMail       = "smtp"
	AdapterConn       = "conn"
	AdapterEs         = "es"
	AdapterRamChandra = "ramchandra"
	AdapterSlack      = "slack"
	AdapterAliLS      = "alils"
)

// Legacy log level constants to ensure backwards compatibility.
const (
	LevelInfo  = LevelInformational
	LevelTrace = LevelDebug
	LevelWarn  = LevelWarning
)

type newLoggerFunc func() Logger

// Logger defines the behavior of a log provider.
type Logger interface {
	Init(config string) error
	WriteMsg(lm *LogMsg) error
	Destroy()
	Flush()
	SetFormatter(f LogFormatter)
}

var adapters = make(map[string]newLoggerFunc)
var levelPrefix = [LevelDebug + 1]string{"[M]", "[A]", "[C]", "[E]", "[W]", "[N]", "[I]", "[D]"}

// Register makes a log provide available by the provided name.
// If Register is called twice with the same name or if driver is nil,
// it panics.
func Register(name string, log newLoggerFunc) {
	if log == nil {
		panic("logs: Register provide is nil")
	}
	if _, dup := adapters[name]; dup {
		panic("logs: Register called twice for provider " + name)
	}
	adapters[name] = log
}

// BhojpurLogger is default logger in Bhojpur.NET Platform application.
// Can contain several providers and log message into all providers.
type BhojpurLogger struct {
	lock                sync.Mutex
	level               int
	init                bool
	enableFuncCallDepth bool
	loggerFuncCallDepth int
	enableFullFilePath  bool
	asynchronous        bool
	prefix              string
	msgChanLen          int64
	msgChan             chan *LogMsg
	signalChan          chan string
	wg                  sync.WaitGroup
	outputs             []*nameLogger
	globalFormatter     string
}

const defaultAsyncMsgLen = 1e3

type nameLogger struct {
	Logger
	name string
}

var logMsgPool *sync.Pool

// NewLogger returns a new BhojpurLogger.
// channelLen: the number of messages in chan(used where asynchronous is true).
// if the buffering chan is full, logger adapters write to file or other way.
func NewLogger(channelLens ...int64) *BhojpurLogger {
	bl := new(BhojpurLogger)
	bl.level = LevelDebug
	bl.loggerFuncCallDepth = 3
	bl.msgChanLen = append(channelLens, 0)[0]
	if bl.msgChanLen <= 0 {
		bl.msgChanLen = defaultAsyncMsgLen
	}
	bl.signalChan = make(chan string, 1)
	bl.setLogger(AdapterConsole)
	return bl
}

// Async sets the log to asynchronous and start the goroutine
func (bl *BhojpurLogger) Async(msgLen ...int64) *BhojpurLogger {
	bl.lock.Lock()
	defer bl.lock.Unlock()
	if bl.asynchronous {
		return bl
	}
	bl.asynchronous = true
	if len(msgLen) > 0 && msgLen[0] > 0 {
		bl.msgChanLen = msgLen[0]
	}
	bl.msgChan = make(chan *LogMsg, bl.msgChanLen)
	logMsgPool = &sync.Pool{
		New: func() interface{} {
			return &LogMsg{}
		},
	}
	bl.wg.Add(1)
	go bl.startLogger()
	return bl
}

// SetLogger provides a given logger adapter into BhojpurLogger with config string.
// config must in in JSON format like {"interval":360}}
func (bl *BhojpurLogger) setLogger(adapterName string, configs ...string) error {
	config := append(configs, "{}")[0]
	for _, l := range bl.outputs {
		if l.name == adapterName {
			return fmt.Errorf("logs: duplicate adaptername %q (you have set this logger before)", adapterName)
		}
	}

	logAdapter, ok := adapters[adapterName]
	if !ok {
		return fmt.Errorf("logs: unknown adaptername %q (forgotten Register?)", adapterName)
	}

	lg := logAdapter()

	// Global formatter overrides the default set formatter
	if len(bl.globalFormatter) > 0 {
		fmtr, ok := GetFormatter(bl.globalFormatter)
		if !ok {
			return errors.New(fmt.Sprintf("the formatter with name: %s not found", bl.globalFormatter))
		}
		lg.SetFormatter(fmtr)
	}

	err := lg.Init(config)

	if err != nil {
		fmt.Fprintln(os.Stderr, "logs.BhojpurLogger.SetLogger: "+err.Error())
		return err
	}
	bl.outputs = append(bl.outputs, &nameLogger{name: adapterName, Logger: lg})
	return nil
}

// SetLogger provides a given logger adapter into BhojpurLogger with config string.
// config must in in JSON format like {"interval":360}}
func (bl *BhojpurLogger) SetLogger(adapterName string, configs ...string) error {
	bl.lock.Lock()
	defer bl.lock.Unlock()
	if !bl.init {
		bl.outputs = []*nameLogger{}
		bl.init = true
	}
	return bl.setLogger(adapterName, configs...)
}

// DelLogger removes a logger adapter in BhojpurLogger.
func (bl *BhojpurLogger) DelLogger(adapterName string) error {
	bl.lock.Lock()
	defer bl.lock.Unlock()
	outputs := []*nameLogger{}
	for _, lg := range bl.outputs {
		if lg.name == adapterName {
			lg.Destroy()
		} else {
			outputs = append(outputs, lg)
		}
	}
	if len(outputs) == len(bl.outputs) {
		return fmt.Errorf("logs: unknown adaptername %q (forgotten Register?)", adapterName)
	}
	bl.outputs = outputs
	return nil
}

func (bl *BhojpurLogger) writeToLoggers(lm *LogMsg) {
	for _, l := range bl.outputs {
		err := l.WriteMsg(lm)
		if err != nil {
			fmt.Fprintf(os.Stderr, "unable to WriteMsg to adapter:%v,error:%v\n", l.name, err)
		}
	}
}

func (bl *BhojpurLogger) Write(p []byte) (n int, err error) {
	if len(p) == 0 {
		return 0, nil
	}
	// writeMsg will always add a '\n' character
	if p[len(p)-1] == '\n' {
		p = p[0 : len(p)-1]
	}
	lm := &LogMsg{
		Msg:   string(p),
		Level: levelLoggerImpl,
	}

	// set levelLoggerImpl to ensure all log message will be write out
	err = bl.writeMsg(lm)
	if err == nil {
		return len(p), err
	}
	return 0, err
}

func (bl *BhojpurLogger) writeMsg(lm *LogMsg) error {
	if !bl.init {
		bl.lock.Lock()
		bl.setLogger(AdapterConsole)
		bl.lock.Unlock()
	}

	var (
		file string
		line int
		ok   bool
	)

	_, file, line, ok = runtime.Caller(bl.loggerFuncCallDepth)
	if !ok {
		file = "???"
		line = 0
	}
	lm.FilePath = file
	lm.LineNumber = line

	lm.enableFullFilePath = bl.enableFullFilePath
	lm.enableFuncCallDepth = bl.enableFuncCallDepth

	// set level info in front of filename info
	if lm.Level == levelLoggerImpl {
		// set to emergency to ensure all log will be print out correctly
		lm.Level = LevelEmergency
	}

	if bl.asynchronous {
		logM := logMsgPool.Get().(*LogMsg)
		logM.Level = lm.Level
		logM.Msg = lm.Msg
		logM.When = lm.When
		logM.Args = lm.Args
		logM.FilePath = lm.FilePath
		logM.LineNumber = lm.LineNumber
		logM.Prefix = lm.Prefix
		if bl.outputs != nil {
			bl.msgChan <- lm
		} else {
			logMsgPool.Put(lm)
		}
	} else {
		bl.writeToLoggers(lm)
	}
	return nil
}

// SetLevel sets log message level.
// If message level (such as LevelDebug) is higher than logger level (such as LevelWarning),
// log providers will not be sent the message.
func (bl *BhojpurLogger) SetLevel(l int) {
	bl.level = l
}

// GetLevel Get Current log message level.
func (bl *BhojpurLogger) GetLevel() int {
	return bl.level
}

// SetLogFuncCallDepth set log funcCallDepth
func (bl *BhojpurLogger) SetLogFuncCallDepth(d int) {
	bl.loggerFuncCallDepth = d
}

// GetLogFuncCallDepth return log funcCallDepth for wrapper
func (bl *BhojpurLogger) GetLogFuncCallDepth() int {
	return bl.loggerFuncCallDepth
}

// EnableFuncCallDepth enable log funcCallDepth
func (bl *BhojpurLogger) EnableFuncCallDepth(b bool) {
	bl.enableFuncCallDepth = b
}

// set prefix
func (bl *BhojpurLogger) SetPrefix(s string) {
	bl.prefix = s
}

// start logger chan reading.
// when chan is not empty, write logs.
func (bl *BhojpurLogger) startLogger() {
	gameOver := false
	for {
		select {
		case bm := <-bl.msgChan:
			bl.writeToLoggers(bm)
			logMsgPool.Put(bm)
		case sg := <-bl.signalChan:
			// Now should only send "flush" or "close" to bl.signalChan
			bl.flush()
			if sg == "close" {
				for _, l := range bl.outputs {
					l.Destroy()
				}
				bl.outputs = nil
				gameOver = true
			}
			bl.wg.Done()
		}
		if gameOver {
			break
		}
	}
}

func (bl *BhojpurLogger) setGlobalFormatter(fmtter string) error {
	bl.globalFormatter = fmtter
	return nil
}

// SetGlobalFormatter sets the global formatter for all log adapters
// don't forget to register the formatter by invoking RegisterFormatter
func SetGlobalFormatter(fmtter string) error {
	return bhojpurLogger.setGlobalFormatter(fmtter)
}

// Emergency Log EMERGENCY level message.
func (bl *BhojpurLogger) Emergency(format string, v ...interface{}) {
	if LevelEmergency > bl.level {
		return
	}

	lm := &LogMsg{
		Level: LevelEmergency,
		Msg:   format,
		When:  time.Now(),
	}
	if len(v) > 0 {
		lm.Msg = fmt.Sprintf(lm.Msg, v...)
	}

	bl.writeMsg(lm)
}

// Alert Log ALERT level message.
func (bl *BhojpurLogger) Alert(format string, v ...interface{}) {
	if LevelAlert > bl.level {
		return
	}

	lm := &LogMsg{
		Level: LevelAlert,
		Msg:   format,
		When:  time.Now(),
		Args:  v,
	}
	bl.writeMsg(lm)
}

// Critical Log CRITICAL level message.
func (bl *BhojpurLogger) Critical(format string, v ...interface{}) {
	if LevelCritical > bl.level {
		return
	}
	lm := &LogMsg{
		Level: LevelCritical,
		Msg:   format,
		When:  time.Now(),
		Args:  v,
	}

	bl.writeMsg(lm)
}

// Error Log ERROR level message.
func (bl *BhojpurLogger) Error(format string, v ...interface{}) {
	if LevelError > bl.level {
		return
	}
	lm := &LogMsg{
		Level: LevelError,
		Msg:   format,
		When:  time.Now(),
		Args:  v,
	}

	bl.writeMsg(lm)
}

// Warning Log WARNING level message.
func (bl *BhojpurLogger) Warning(format string, v ...interface{}) {
	if LevelWarn > bl.level {
		return
	}
	lm := &LogMsg{
		Level: LevelWarn,
		Msg:   format,
		When:  time.Now(),
		Args:  v,
	}

	bl.writeMsg(lm)
}

// Notice Log NOTICE level message.
func (bl *BhojpurLogger) Notice(format string, v ...interface{}) {
	if LevelNotice > bl.level {
		return
	}
	lm := &LogMsg{
		Level: LevelNotice,
		Msg:   format,
		When:  time.Now(),
		Args:  v,
	}

	bl.writeMsg(lm)
}

// Informational Log INFORMATIONAL level message.
func (bl *BhojpurLogger) Informational(format string, v ...interface{}) {
	if LevelInfo > bl.level {
		return
	}
	lm := &LogMsg{
		Level: LevelInfo,
		Msg:   format,
		When:  time.Now(),
		Args:  v,
	}

	bl.writeMsg(lm)
}

// Debug Log DEBUG level message.
func (bl *BhojpurLogger) Debug(format string, v ...interface{}) {
	if LevelDebug > bl.level {
		return
	}
	lm := &LogMsg{
		Level: LevelDebug,
		Msg:   format,
		When:  time.Now(),
		Args:  v,
	}

	bl.writeMsg(lm)
}

// Warn Log WARN level message.
// compatibility alias for Warning()
func (bl *BhojpurLogger) Warn(format string, v ...interface{}) {
	if LevelWarn > bl.level {
		return
	}
	lm := &LogMsg{
		Level: LevelWarn,
		Msg:   format,
		When:  time.Now(),
		Args:  v,
	}

	bl.writeMsg(lm)
}

// Info Log INFO level message.
// compatibility alias for Informational()
func (bl *BhojpurLogger) Info(format string, v ...interface{}) {
	if LevelInfo > bl.level {
		return
	}
	lm := &LogMsg{
		Level: LevelInfo,
		Msg:   format,
		When:  time.Now(),
		Args:  v,
	}

	bl.writeMsg(lm)
}

// Trace Log TRACE level message.
// compatibility alias for Debug()
func (bl *BhojpurLogger) Trace(format string, v ...interface{}) {
	if LevelDebug > bl.level {
		return
	}
	lm := &LogMsg{
		Level: LevelDebug,
		Msg:   format,
		When:  time.Now(),
		Args:  v,
	}

	bl.writeMsg(lm)
}

// Flush flush all chan data.
func (bl *BhojpurLogger) Flush() {
	if bl.asynchronous {
		bl.signalChan <- "flush"
		bl.wg.Wait()
		bl.wg.Add(1)
		return
	}
	bl.flush()
}

// Close close logger, flush all chan data and destroy all adapters in BhojpurLogger.
func (bl *BhojpurLogger) Close() {
	if bl.asynchronous {
		bl.signalChan <- "close"
		bl.wg.Wait()
		close(bl.msgChan)
	} else {
		bl.flush()
		for _, l := range bl.outputs {
			l.Destroy()
		}
		bl.outputs = nil
	}
	close(bl.signalChan)
}

// Reset close all outputs, and set bl.outputs to nil
func (bl *BhojpurLogger) Reset() {
	bl.Flush()
	for _, l := range bl.outputs {
		l.Destroy()
	}
	bl.outputs = nil
}

func (bl *BhojpurLogger) flush() {
	if bl.asynchronous {
		for {
			if len(bl.msgChan) > 0 {
				bm := <-bl.msgChan
				bl.writeToLoggers(bm)
				logMsgPool.Put(bm)
				continue
			}
			break
		}
	}
	for _, l := range bl.outputs {
		l.Flush()
	}
}

// bhojpurLogger references the used application logger.
var bhojpurLogger = NewLogger()

// GetBhojpurLogger returns the default BhojpurLogger
func GetBhojpurLogger() *BhojpurLogger {
	return bhojpurLogger
}

var bhojpurLoggerMap = struct {
	sync.RWMutex
	logs map[string]*log.Logger
}{
	logs: map[string]*log.Logger{},
}

// GetLogger returns the default BhojpurLogger
func GetLogger(prefixes ...string) *log.Logger {
	prefix := append(prefixes, "")[0]
	if prefix != "" {
		prefix = fmt.Sprintf(`[%s] `, strings.ToUpper(prefix))
	}
	bhojpurLoggerMap.RLock()
	l, ok := bhojpurLoggerMap.logs[prefix]
	if ok {
		bhojpurLoggerMap.RUnlock()
		return l
	}
	bhojpurLoggerMap.RUnlock()
	bhojpurLoggerMap.Lock()
	defer bhojpurLoggerMap.Unlock()
	l, ok = bhojpurLoggerMap.logs[prefix]
	if !ok {
		l = log.New(bhojpurLogger, prefix, 0)
		bhojpurLoggerMap.logs[prefix] = l
	}
	return l
}

// EnableFullFilePath enables full file path logging. Disabled by default
// e.g "/home/Documents/GitHub/bhojpur/mainapp/" instead of "mainapp"
func EnableFullFilePath(b bool) {
	bhojpurLogger.enableFullFilePath = b
}

// Reset will remove all the adapter
func Reset() {
	bhojpurLogger.Reset()
}

// Async set the bhojpurlogger with Async mode and hold msglen messages
func Async(msgLen ...int64) *BhojpurLogger {
	return bhojpurLogger.Async(msgLen...)
}

// SetLevel sets the global log level used by the simple logger.
func SetLevel(l int) {
	bhojpurLogger.SetLevel(l)
}

// SetPrefix sets the prefix
func SetPrefix(s string) {
	bhojpurLogger.SetPrefix(s)
}

// EnableFuncCallDepth enable log funcCallDepth
func EnableFuncCallDepth(b bool) {
	bhojpurLogger.enableFuncCallDepth = b
}

// SetLogFuncCall set the CallDepth, default is 4
func SetLogFuncCall(b bool) {
	bhojpurLogger.EnableFuncCallDepth(b)
	bhojpurLogger.SetLogFuncCallDepth(3)
}

// SetLogFuncCallDepth set log funcCallDepth
func SetLogFuncCallDepth(d int) {
	bhojpurLogger.loggerFuncCallDepth = d
}

// SetLogger sets a new logger.
func SetLogger(adapter string, config ...string) error {
	return bhojpurLogger.SetLogger(adapter, config...)
}

// Emergency logs a message at emergency level.
func Emergency(f interface{}, v ...interface{}) {
	bhojpurLogger.Emergency(formatLog(f, v...))
}

// Alert logs a message at alert level.
func Alert(f interface{}, v ...interface{}) {
	bhojpurLogger.Alert(formatLog(f, v...))
}

// Critical logs a message at critical level.
func Critical(f interface{}, v ...interface{}) {
	bhojpurLogger.Critical(formatLog(f, v...))
}

// Error logs a message at error level.
func Error(f interface{}, v ...interface{}) {
	bhojpurLogger.Error(formatLog(f, v...))
}

// Warning logs a message at warning level.
func Warning(f interface{}, v ...interface{}) {
	bhojpurLogger.Warn(formatLog(f, v...))
}

// Warn compatibility alias for Warning()
func Warn(f interface{}, v ...interface{}) {
	bhojpurLogger.Warn(formatLog(f, v...))
}

// Notice logs a message at notice level.
func Notice(f interface{}, v ...interface{}) {
	bhojpurLogger.Notice(formatLog(f, v...))
}

// Informational logs a message at info level.
func Informational(f interface{}, v ...interface{}) {
	bhojpurLogger.Info(formatLog(f, v...))
}

// Info compatibility alias for Warning()
func Info(f interface{}, v ...interface{}) {
	bhojpurLogger.Info(formatLog(f, v...))
}

// Debug logs a message at debug level.
func Debug(f interface{}, v ...interface{}) {
	bhojpurLogger.Debug(formatLog(f, v...))
}

// Trace logs a message at trace level.
// compatibility alias for Warning()
func Trace(f interface{}, v ...interface{}) {
	bhojpurLogger.Trace(formatLog(f, v...))
}

func formatLog(f interface{}, v ...interface{}) string {
	var msg string
	switch f.(type) {
	case string:
		msg = f.(string)
		if len(v) == 0 {
			return msg
		}
		if !strings.Contains(msg, "%") {
			// do not contain format char
			msg += strings.Repeat(" %v", len(v))
		}
	default:
		msg = fmt.Sprint(f)
		if len(v) == 0 {
			return msg
		}
		msg += strings.Repeat(" %v", len(v))
	}
	return fmt.Sprintf(msg, v...)
}