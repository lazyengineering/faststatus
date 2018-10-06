// Copyright 2016 Jesse Allen. All rights reserved
// Released under the MIT license found in the LICENSE file.

package server

import (
	"errors"
	"net/http"
	"reflect"
	"testing"
)

func TestSrvErrorError(t *testing.T) {
	testCases := []struct {
		cause       error
		code        int
		message     string
		errorString string
	}{
		{
			errors.New("bang"),
			http.StatusBadRequest,
			"bad request",
			"400: bad request: bang",
		},
		{
			nil,
			http.StatusNotFound,
			"not found",
			"404: not found",
		},
		{
			errors.New("lots of stuff"),
			0,
			"some stuff",
			"000: some stuff: lots of stuff",
		},
		{
			srvError{
				cause:   errors.New("bang"),
				code:    500,
				message: "something bad",
			},
			http.StatusBadRequest,
			"bad request",
			"400: bad request: 500: something bad: bang",
		},
	}

	for _, tc := range testCases {
		e := srvError{
			cause:   tc.cause,
			code:    tc.code,
			message: tc.message,
		}
		if e.Error() != tc.errorString {
			t.Fatalf("e.Error() = %q, expected %q", e.Error(), tc.errorString)
		}
	}
}

func TestErrf(t *testing.T) {
	testCases := []struct {
		cause   error
		code    int
		format  string
		a       []interface{}
		message string
	}{
		{
			errors.New("bang"),
			http.StatusTeapot,
			"%d, %v - %s",
			[]interface{}{
				15,
				nil,
				"boom",
			},
			"15, <nil> - boom",
		},
		{
			nil,
			http.StatusOK,
			"meh",
			nil,
			"meh",
		},
	}

	for _, tc := range testCases {
		te := srvError{
			cause:   tc.cause,
			code:    tc.code,
			message: tc.message,
		}
		e := errf(tc.cause, tc.code, tc.format, tc.a...)
		if !reflect.DeepEqual(e, te) {
			t.Fatalf("errf(%+v, %d, %q, %+v) = %+v, expected %+v", tc.cause, tc.code, tc.format, tc.a, e, te)
		}
	}
}

func TestErrorCode(t *testing.T) {
	testCases := []struct {
		err  error
		code int
	}{
		{
			srvError{
				cause:   errors.New("bang"),
				code:    http.StatusNotFound,
				message: "not found",
			},
			http.StatusNotFound,
		},
		{
			nil,
			http.StatusInternalServerError,
		},
		{
			errors.New("lots of stuff"),
			http.StatusInternalServerError,
		},
	}

	for _, tc := range testCases {
		code := ErrorCode(tc.err)
		if code != tc.code {
			t.Fatalf("ErrorCode(%+v) = %d, expected %d", code, tc.code)
		}
	}
}
