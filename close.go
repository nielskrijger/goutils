package goutils

import (
	"io"

	"github.com/rs/zerolog"
)

// Close is a simple utility used to log error messages riased during a deferred closing function.
// For example:
//
//    defer utils.Close(myLog, f)
//
func Close(log zerolog.Logger, c io.Closer) {
	if err := c.Close(); err != nil {
		log.Error().Err(err).Msgf("error while closing: %s", err)
	}
}
