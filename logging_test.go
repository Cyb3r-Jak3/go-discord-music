package main

import (
	"testing"

	"github.com/sirupsen/logrus"
)

type testHook struct {
	entries []logrus.Entry
}

func (h *testHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (h *testHook) Fire(e *logrus.Entry) error {
	h.entries = append(h.entries, *e)
	return nil
}

func TestSetLogLevel(t *testing.T) {
	logger := logrus.New()
	hook := &testHook{}
	logger.AddHook(hook)

	err := SetLogLevel(logger, "debug")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if logger.Level != logrus.DebugLevel {
		t.Errorf("expected log level debug, got %v", logger.Level)
	}

	err = SetLogLevel(logger, "invalid")
	if err == nil {
		t.Errorf("expected error for invalid log level, got nil")
	}
}
