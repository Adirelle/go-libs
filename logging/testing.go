package logging

import (
	"errors"
	"io"
	"io/ioutil"
	"log"
	"testing"

	"go.uber.org/zap/zapcore"
)

//===========================================================================
// testingLogger
//===========================================================================

type testingLogger struct {
	t *testing.T
}

// NewTesting creates a logger that forwards everything to the testing log.
func NewTesting(t *testing.T) Logger {
	return &testingLogger{t}
}

func (l *testingLogger) DPanic(a ...interface{})            { l.t.Log(a...) }
func (l *testingLogger) DPanicf(s string, a ...interface{}) { l.t.Logf(s, a...) }
func (l *testingLogger) DPanicw(s string, a ...interface{}) { l.t.Log(append([]interface{}{s}, a...)) }
func (l *testingLogger) Debug(a ...interface{})             { l.t.Log(a...) }
func (l *testingLogger) Debugf(s string, a ...interface{})  { l.t.Logf(s, a...) }
func (l *testingLogger) Debugw(s string, a ...interface{})  { l.t.Log(append([]interface{}{s}, a...)) }
func (l *testingLogger) Error(a ...interface{})             { l.t.Log(a...) }
func (l *testingLogger) Errorf(s string, a ...interface{})  { l.t.Logf(s, a...) }
func (l *testingLogger) Errorw(s string, a ...interface{})  { l.t.Log(append([]interface{}{s}, a...)) }
func (l *testingLogger) Fatal(a ...interface{})             { l.t.Log(a...) }
func (l *testingLogger) Fatalf(s string, a ...interface{})  { l.t.Logf(s, a...) }
func (l *testingLogger) Fatalw(s string, a ...interface{})  { l.t.Log(append([]interface{}{s}, a...)) }
func (l *testingLogger) Info(a ...interface{})              { l.t.Log(a...) }
func (l *testingLogger) Infof(s string, a ...interface{})   { l.t.Logf(s, a...) }
func (l *testingLogger) Infow(s string, a ...interface{})   { l.t.Log(append([]interface{}{s}, a...)) }
func (l *testingLogger) Panic(a ...interface{})             { l.t.Log(a...) }
func (l *testingLogger) Panicf(s string, a ...interface{})  { l.t.Logf(s, a...) }
func (l *testingLogger) Panicw(s string, a ...interface{})  { l.t.Log(append([]interface{}{s}, a...)) }
func (l *testingLogger) Warn(a ...interface{})              { l.t.Log(a...) }
func (l *testingLogger) Warnf(s string, a ...interface{})   { l.t.Logf(s, a...) }
func (l *testingLogger) Warnw(s string, a ...interface{})   { l.t.Log(append([]interface{}{s}, a...)) }
func (l *testingLogger) Named(string) Logger                { return l }
func (l *testingLogger) With(...interface{}) Logger         { return l }
func (l *testingLogger) Sync() error                        { return nil }
func (l *testingLogger) Writer() io.WriteCloser             { return nopWriter{ioutil.Discard} }

func (l *testingLogger) StdLoggerAt(_ zapcore.Level) (*log.Logger, error) {
	return nil, errors.New("Not implemented")
}

//===========================================================================
// nopWriter
//===========================================================================

type nopWriter struct{ io.Writer }

func (nopWriter) Close() error {
	return nil
}
