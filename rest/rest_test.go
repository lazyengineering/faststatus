// Copyright 2017 Jesse Allen. All rights reserved
// Released under the MIT license found in the LICENSE file.

package rest_test

import (
	"bytes"
	"math/rand"
	"mime"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"sync"
	"testing"
	"testing/quick"

	"github.com/lazyengineering/faststatus"
	"github.com/lazyengineering/faststatus/rest"
)

func TestHandlerOnlyValidPaths(t *testing.T) {
	var s, _ = rest.New()
	invalidPathIsNotFound := func(method, path string) bool {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(method, path, nil)
		s.ServeHTTP(w, r)

		if w.Code != http.StatusNotFound {
			t.Logf("%s %q got %03d, expected 404", method, path, w.Code)
			return false
		}
		return true
	}
	err := quick.Check(invalidPathIsNotFound, &quick.Config{
		Values: func(args []reflect.Value, gen *rand.Rand) {
			args[0] = reflect.ValueOf(possibleMethods[gen.Intn(len(possibleMethods))])
			args[1] = reflect.ValueOf(genInvalidPath(gen.Intn(100)+1, gen))
		},
	})
	if err != nil {
		t.Fatalf("unexpected response to bad client request: %+v", err)
	}
}

func TestHandlerOnlyValidPathsAndMethods(t *testing.T) {
	var s, _ = rest.New()
	isNotAllowed := func(path string) func(string) bool {
		return func(method string) bool {
			w := httptest.NewRecorder()
			r := httptest.NewRequest(method, path, nil)
			s.ServeHTTP(w, r)

			validMethods := validMethodsByPath[path]
			var isValid bool
			for _, validMethod := range validMethods {
				if validMethod == method {
					isValid = true
					break
				}
			}
			if !isValid && w.Code != http.StatusMethodNotAllowed {
				t.Logf("%s %q got %03d, expected 405", method, path, w.Code)
				return false
			}
			if isValid && w.Code == http.StatusMethodNotAllowed {
				t.Logf("%s %q got %03d, expected not 405", method, path, w.Code)
				return false
			}
			return true
		}
	}
	genMethod := func(args []reflect.Value, gen *rand.Rand) {
		args[0] = reflect.ValueOf(possibleMethods[gen.Intn(len(possibleMethods))])
	}
	for path := range validMethodsByPath {
		if err := quick.Check(isNotAllowed(path), &quick.Config{Values: genMethod}); err != nil {
			t.Fatalf("unexpected response to bad client request: %+v", err)
		}
	}
}

var possibleMethods = []string{
	http.MethodGet,
	http.MethodHead,
	http.MethodPut,
	http.MethodPost,
	http.MethodPatch,
	http.MethodDelete,
}
var validMethodsByPath = map[string][]string{
	"/new": []string{http.MethodGet, http.MethodHead},
}

func TestHandlerGetNew(t *testing.T) {
	const defaultContentType = "text/plain"
	var s, _ = rest.New()
	var mu sync.Mutex
	var usedIDs = make(map[faststatus.ID]struct{})
	getsNewResource := func() bool {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/new", nil)
		s.ServeHTTP(w, r)

		if w.Code != http.StatusOK {
			t.Logf(
				"returned Status Code %03d, expected %03d",
				w.Code,
				http.StatusOK,
			)
			return false
		}

		gotType, _, err := mime.ParseMediaType(w.HeaderMap.Get("Content-Type"))
		if err != nil {
			t.Logf("error parsing content type: %+v", err)
			return false
		}
		if gotType != defaultContentType {
			t.Logf("Content-Type %q, expected %q", gotType, defaultContentType)
			return false
		}

		var (
			gotResource  faststatus.Resource
			zeroResource faststatus.Resource
		)
		err = (&gotResource).UnmarshalText(w.Body.Bytes())
		if err != nil {
			t.Logf("error unmarshaling Resource from text body: %+v", err)
			return false
		}

		if bytes.Equal(gotResource.ID[:], zeroResource.ID[:]) {
			t.Logf("ID should be non-zero")
			return false
		}

		if _, exists := usedIDs[gotResource.ID]; exists {
			t.Logf("ID should be unique")
			return false
		}
		mu.Lock()
		usedIDs[gotResource.ID] = struct{}{}
		mu.Unlock()

		gotResource.ID = zeroResource.ID
		if !gotResource.Equal(zeroResource) {
			t.Logf("with ID set to zero-value got %+v, expected %+v", gotResource, zeroResource)
			return false
		}

		return true
	}
	if err := quick.Check(getsNewResource, nil); err != nil {
		t.Fatalf("GET /new does not get a new Resource: %+v", err)
	}
}

var possibleMethods = []string{
	http.MethodGet,
	http.MethodHead,
	http.MethodPut,
	http.MethodPost,
	http.MethodPatch,
	http.MethodDelete,
}

func validMethodsByPath(path string) ([]string, bool) {
	if path == "/new" {
		return []string{http.MethodGet, http.MethodHead}, true
	}
	return nil, false
}

func genInvalidPath(maxLen int, r *rand.Rand) string {
	for {
		path := genPath(maxLen, r)
		_, ok := validMethodsByPath(path)
		if !ok {
			return path
		}
	}
}

func genPath(maxLen int, r *rand.Rand) string {
	scratch := make([]byte, 1000)
	r.Read(scratch)
	var path = "/" + strings.Map(func(char rune) rune {
		switch {
		case char >= '0' && char <= '9',
			char >= 'A' && char <= 'Z',
			char >= 'a' && char <= 'z',
			char == '!',
			char == '$',
			char >= 0x27 && char <= '/':
			return char
		default:
			return -1
		}
	}, string(scratch))
	return path[:maxLen%len(path)]
}
