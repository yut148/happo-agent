package util

import (
	"errors"
	"fmt"
	"os"
	"sync"
)

// LogWriter is custom log writer
type LogWriter struct {
	fp *os.File
	sync.Mutex
	filename string
}

// NewLogWriter returns new LogWriter pointer
func NewLogWriter(filename string) *LogWriter {
	w := &LogWriter{filename: filename}
	err := w.ReOpen()
	if err != nil {
		fmt.Println(err)
	}

	return w
}

// ReOpen do reopen logs for rotation
func (w *LogWriter) ReOpen() error {
	w.Lock()
	defer w.Unlock()

	w.fp.Close()
	fp, err := os.Open(w.filename)
	if err != nil {
		fp, err = os.Create(w.filename)
	}
	w.fp = fp
	return err
}

// Close close LogWriter
func (w *LogWriter) Close() error {
	return w.fp.Close()
}

// File returns fp
func (w *LogWriter) File() *os.File {
	return w.fp
}

func (w *LogWriter) Write(output []byte) (int, error) {
	w.Lock()
	defer w.Unlock()

	if w.fp == nil {
		return 0, errors.New("fp is nil")
	}
	return w.fp.Write(output)
}
