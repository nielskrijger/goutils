package goutils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
)

type AssertJSON struct {
	T    *testing.T
	Body []byte
}

func NewAssertJSON(t *testing.T, body []byte) *AssertJSON {
	t.Helper()

	return &AssertJSON{T: t, Body: body}
}

func (a *AssertJSON) Regexp(path string, rx interface{}, msgAndArgs ...interface{}) {
	assert.Regexp(a.T, rx, gjson.GetBytes(a.Body, path).Value(), msgAndArgs...)
}

func (a *AssertJSON) Equal(path string, expected interface{}, msgAndArgs ...interface{}) {
	assert.Equal(a.T, expected, gjson.GetBytes(a.Body, path).Value(), msgAndArgs...)
}

func (a *AssertJSON) Raw(path string, expected interface{}, msgAndArgs ...interface{}) {
	assert.Equal(a.T, expected, gjson.GetBytes(a.Body, path).Raw, msgAndArgs...)
}

func (a *AssertJSON) Len(path string, length int, msgAndArgs ...interface{}) {
	assert.Len(a.T, gjson.GetBytes(a.Body, path).Array(), length, msgAndArgs...)
}

func (a *AssertJSON) Nil(path string, msgAndArgs ...interface{}) {
	assert.Nil(a.T, gjson.GetBytes(a.Body, path).Value(), msgAndArgs...)
}

func (a *AssertJSON) TimeBetween(path string, minDur time.Duration, maxDur time.Duration, msgAndArgs ...interface{}) {
	timeUntil := time.Until(gjson.GetBytes(a.Body, path).Time())
	assert.GreaterOrEqual(a.T, timeUntil, minDur, msgAndArgs...)
	assert.LessOrEqual(a.T, timeUntil, maxDur, msgAndArgs...)
}

func (a *AssertJSON) True(path string, msgAndArgs ...interface{}) {
	assert.True(a.T, gjson.GetBytes(a.Body, path).Bool(), msgAndArgs...)
}

func (a *AssertJSON) False(path string, msgAndArgs ...interface{}) {
	assert.False(a.T, gjson.GetBytes(a.Body, path).Bool(), msgAndArgs...)
}
