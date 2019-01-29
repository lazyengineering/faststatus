// Copyright 2017 Jesse Allen. All rights reserved
// Released under the MIT license found in the LICENSE file.

package rest

import (
	"fmt"
	"net/http"

	"github.com/pkg/errors"
)

type restError struct {
	err  error
	code int
}

func (e restError) Error() string {
	return fmt.Sprintf("%03d %+v", e.code, e.err)
}

func (e restError) Code() int {
	return e.code
}

func errorCode(e error) int {
	type codeError interface {
		Code() int
	}
	if e, ok := errors.Cause(e).(codeError); ok {
		return e.Code()
	}
	return http.StatusInternalServerError
}
