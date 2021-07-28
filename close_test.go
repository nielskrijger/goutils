package goutils_test

import (
	"errors"
	"testing"

	utils "github.com/nielskrijger/goutils"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

var errTest = errors.New("test error")

type closerMock struct {
	err error
}

func (c *closerMock) Close() error {
	return c.err
}

func TestClose_LogError(t *testing.T) {
	testLogger := &utils.TestLogger{}

	utils.Close(zerolog.New(testLogger), &closerMock{err: errTest})
	assert.Equal(t, "error while closing: test error", testLogger.LastLine()["message"])
	assert.Equal(t, "test error", testLogger.LastLine()["error"])
}

func TestClose_NoLog(t *testing.T) {
	testLogger := &utils.TestLogger{}

	utils.Close(zerolog.New(testLogger), &closerMock{err: errTest})
	assert.Equal(t, "error while closing: test error", testLogger.LastLine()["message"])
	assert.Equal(t, "test error", testLogger.LastLine()["error"])
}
