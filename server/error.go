// Copyright 2016 Jesse Allen. All rights reserved
// Released under the MIT license found in the LICENSE file.

package server

import (
	"fmt"
	"net/http"
)

type srvError struct {
	cause   error
	code    int
	message string
}

func errf(cause error, code int, format string, a ...interface{}) error {
	return srvError{
		cause:   cause,
		code:    code,
		message: fmt.Sprintf(format, a...),
	}
}

func (e srvError) Error() string {
	if e.cause != nil {
		return fmt.Sprintf("%03d: %s: %+v", e.code, e.message, e.cause)
	}
	return fmt.Sprintf("%03d: %s", e.code, e.message)
}

func (e srvError) ErrorCode() int {
	return e.code
}

type errorCoder interface {
	ErrorCode() int
}

func ErrorCode(e error) int {
	ev, ok := e.(errorCoder)
	if !ok {
		return http.StatusInternalServerError
	}
	return ev.ErrorCode()
}
