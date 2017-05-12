// Copyright 2017 Jesse Allen. All rights reserved
// Released under the MIT license found in the LICENSE file.

package rest_test

import (
	"bytes"
	"fmt"
	"math/rand"
	"mime"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"sync"
	"testing"
	"testing/quick"
	"time"
	"unicode/utf8"

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
	var s, _ = rest.New(rest.WithStore(&mockStore{
		saveFn: func(faststatus.Resource) error {
			return fmt.Errorf("an error")
		},
		getFn: func(faststatus.ID) (faststatus.Resource, error) {
			return faststatus.Resource{}, fmt.Errorf("an error")
		},
	}))
	isNotAllowed := func(method, path string) bool {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(method, path, nil)
		s.ServeHTTP(w, r)

		validMethods, _ := validMethodsByPath(path)
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
	gen := func(args []reflect.Value, r *rand.Rand) {
		args[0] = reflect.ValueOf(possibleMethods[r.Intn(len(possibleMethods))])
		args[1] = reflect.ValueOf(genValidPath(r))
	}
	if err := quick.Check(isNotAllowed, &quick.Config{Values: gen}); err != nil {
		t.Fatalf("unexpected response to bad client request: %+v", err)
	}
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

func TestHandlerPutToID(t *testing.T) {
	//TODO(jesse@jessecarl.com): Once the errors can be inspected to identify conflicts, add 409 status
	//TODO(jesse@jessecarl.com): Content negotiation. For now, everything is text/plain.
	t.Run("bad requests", func(t *testing.T) {
		var s, _ = rest.New()
		rejectsBadRequests := func(path string, body []byte) bool {
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodPut, path, bytes.NewReader(body))
			s.ServeHTTP(w, r)
			if w.Code != http.StatusBadRequest {
				t.Logf("returned Status Code %03d, expected %03d", w.Code, http.StatusBadRequest)
				return false
			}
			return true
		}
		testCases := []struct {
			name   string
			values func([]reflect.Value, *rand.Rand)
		}{
			{"failure to unmarshal",
				func(val []reflect.Value, r *rand.Rand) {
					id, _ := faststatus.NewID()
					b, _ := id.MarshalText()
					val[0] = reflect.ValueOf("/" + string(b))
					val[1] = reflect.ValueOf(genBadBody(r))
				},
			},
			{"zero-value Since",
				func(val []reflect.Value, r *rand.Rand) {
					resource := faststatus.NewResource()
					resource.Status = faststatus.Status(r.Intn(2))
					id, _ := resource.ID.MarshalText()
					txt, _ := resource.MarshalText()
					val[0] = reflect.ValueOf("/" + string(id))
					val[1] = reflect.ValueOf(txt)
				},
			},
			{"id does not match ID in Resource",
				func(val []reflect.Value, r *rand.Rand) {
					resource := faststatus.NewResource()
					resource.Status = faststatus.Status(r.Intn(2))
					resource.Since = time.Now() // now is random enough..
					var id faststatus.ID
					for {
						id, _ = faststatus.NewID()
						if id != resource.ID {
							break
						}
					}
					idTxt, _ := id.MarshalText()
					txt, _ := resource.MarshalText()
					val[0] = reflect.ValueOf("/" + string(idTxt))
					val[1] = reflect.ValueOf(txt)
				},
			},
		}
		for _, tc := range testCases {
			tc := tc
			t.Run(tc.name, func(t *testing.T) {
				err := quick.Check(rejectsBadRequests, &quick.Config{Values: tc.values})
				if err != nil {
					t.Fatalf("failed to reject bad values: %+v", err)
				}
			})
		}
	})

	t.Run("body Read error", func(t *testing.T) {
		id, _ := faststatus.NewID()
		b, _ := id.MarshalText()

		var s, _ = rest.New()
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPut, "/"+string(b), errorReader{})
		s.ServeHTTP(w, r)
		if w.Code != http.StatusInternalServerError {
			t.Fatalf("returned Status Code %03d, expected %03d", w.Code, http.StatusInternalServerError)
		}
	})

	t.Run("store Save error", func(t *testing.T) {
		store := &mockStore{saveFn: func(r faststatus.Resource) error {
			return fmt.Errorf("an error")
		}}
		var s, err = rest.New(rest.WithStore(store))
		if err != nil {
			t.Fatalf("unexpected error creating store: %+v", err)
		}

		resource := faststatus.NewResource()
		resource.Since = time.Date(2017, 3, 14, 15, 9, 26, 5359, time.UTC)
		body, _ := resource.MarshalText()
		id, _ := resource.ID.MarshalText()

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPut, "/"+string(id), bytes.NewReader(body))
		s.ServeHTTP(w, r)
		if w.Code != http.StatusInternalServerError {
			t.Fatalf("returned Status Code %03d, expected %03d", w.Code, http.StatusInternalServerError)
		}
		if store.saveCalled != 1 {
			t.Fatalf("Store Save called %d times, expected exactly once", store.saveCalled)
		}
	})

	t.Run("good requests", func(t *testing.T) {
		goodRequestsSave := func(r faststatus.Resource) bool {
			path, _ := r.ID.MarshalText()
			body, _ := r.MarshalText()

			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPut, "/"+string(path), bytes.NewReader(body))

			var callSaveWithResource bool
			store := &mockStore{
				saveFn: func(r2 faststatus.Resource) error {
					callSaveWithResource = r.Equal(r2)
					return nil
				},
			}

			s, _ := rest.New(rest.WithStore(store))
			s.ServeHTTP(w, req)
			if w.Code != http.StatusOK {
				t.Logf("returned Status Code %03d, expected %03d", w.Code, http.StatusOK)
				return false
			}

			if !callSaveWithResource {
				t.Logf("did not call save with expected resource")
				return false
			}

			if !bytes.Equal(w.Body.Bytes(), body) {
				t.Logf("responded with %s, expected %s", w.Body.Bytes(), body)
				return false
			}

			return true
		}
		err := quick.Check(goodRequestsSave, &quick.Config{
			Values: func(val []reflect.Value, r *rand.Rand) {
				val[0] = reflect.ValueOf(genResource(140, r))
			},
		})
		if err != nil {
			t.Fatalf("failed to save a good request: %+v", err)
		}
	})
}

func TestHandlerGetFromID(t *testing.T) {
	//TODO(jesse@jessecarl.com): Once the errors can be inspected to identify conflicts, add 409 status
	//TODO(jesse@jessecarl.com): Content negotiation. For now, everything is text/plain.
	t.Run("store get error", func(t *testing.T) {
		store := &mockStore{getFn: func(faststatus.ID) (faststatus.Resource, error) {
			return faststatus.Resource{}, fmt.Errorf("an error")
		}}
		var s, err = rest.New(rest.WithStore(store))
		if err != nil {
			t.Fatalf("unexpected error creating store: %+v", err)
		}

		id, _ := faststatus.NewID()
		idB, _ := id.MarshalText()

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/"+string(idB), nil)
		s.ServeHTTP(w, r)
		if w.Code != http.StatusInternalServerError {
			t.Fatalf("returned Status Code %03d, expected %03d", w.Code, http.StatusInternalServerError)
		}
		if store.getCalled != 1 {
			t.Fatalf("Store Get called %d times, expected exactly once", store.getCalled)
		}
	})

	t.Run("store get not found", func(t *testing.T) {
		store := &mockStore{getFn: func(faststatus.ID) (faststatus.Resource, error) {
			return faststatus.Resource{}, nil
		}}
		var s, err = rest.New(rest.WithStore(store))
		if err != nil {
			t.Fatalf("unexpected error creating store: %+v", err)
		}

		id, _ := faststatus.NewID()
		idB, _ := id.MarshalText()

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/"+string(idB), nil)
		s.ServeHTTP(w, r)
		if w.Code != http.StatusNotFound {
			t.Fatalf("returned Status Code %03d, expected %03d", w.Code, http.StatusNotFound)
		}
		if store.getCalled != 1 {
			t.Fatalf("Store Get called %d times, expected exactly once", store.getCalled)
		}
	})

	t.Run("store get resource", func(t *testing.T) {
		// given assorted Resources, expect that resource to be returned from a get request
		getsBackExpectedResource := func(resource faststatus.Resource) bool {
			var calledCorrectly bool
			store := &mockStore{getFn: func(id faststatus.ID) (faststatus.Resource, error) {
				calledCorrectly = id == resource.ID
				return resource, nil
			}}
			var s, err = rest.New(rest.WithStore(store))
			if err != nil {
				t.Fatalf("unexpected error creating store: %+v", err)
			}

			idB, _ := resource.ID.MarshalText()

			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, "/"+string(idB), nil)
			s.ServeHTTP(w, r)
			if w.Code != http.StatusOK {
				t.Logf("returned Status Code %03d, expected %03d", w.Code, http.StatusOK)
				return false
			}
			if store.getCalled != 1 {
				t.Logf("Store Get called %d times, expected exactly once", store.getCalled)
				return false
			}
			var got faststatus.Resource
			err = (&got).UnmarshalText(w.Body.Bytes())
			if err != nil {
				t.Logf("Response body failed to unmarshal to Resource: %+v", err)
				return false
			}
			if !got.Equal(resource) {
				t.Logf("Response body unmarshals to %+v, expected %+v", got, resource)
				return false
			}
			return true
		}
		err := quick.Check(getsBackExpectedResource, &quick.Config{Values: func(val []reflect.Value, r *rand.Rand) {
			val[0] = reflect.ValueOf(genResource(140, r))
		}})
		if err != nil {
			t.Fatalf("unexpected response to good get resource request: %+v", err)
		}
	})
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
	parts := strings.SplitN(path, "/", 3)
	if len(parts) < 2 {
		return nil, false
	}
	id := new(faststatus.ID)
	err := id.UnmarshalText([]byte(parts[1]))
	if err != nil {
		return nil, false
	}
	methods := []string{http.MethodGet, http.MethodHead}
	if len(parts) < 3 { // no trailing slash (for now)
		methods = append(methods, http.MethodPut)
	}
	return methods, true
}

func genValidPath(r *rand.Rand) string {
	pathFuncs := []func() string{
		func() string { return "/new" },
		func() string { // base ID
			id, _ := faststatus.NewID()
			b, _ := id.MarshalText()
			return "/" + string(b)
		},
	}
	return pathFuncs[r.Intn(len(pathFuncs))]()
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

func genBadBody(r *rand.Rand) []byte {
	scratch := make([]byte, r.Intn(1000))
	r.Read(scratch)
	return scratch
}

func genResource(size int, rgen *rand.Rand) faststatus.Resource {
	rr := faststatus.Resource{}
	rr.ID, _ = faststatus.NewID()
	rr.FriendlyName = func(size int, rgen *rand.Rand) string {
		txt := make([]byte, 0, size)
		for len(txt) < size {
			p := make([]byte, 1)
			n, err := rgen.Read(p)
			if err != nil {
				panic(err)
			}
			if n != 1 {
				continue
			}
			if utf8.Valid(p) {
				txt = append(txt, p...)
			}
		}
		return string(txt)
	}(rgen.Intn(100), rgen)
	rr.Status = faststatus.Status(rgen.Int() % int(faststatus.Occupied))
	rr.Since = time.Date(
		2016+rgen.Intn(10),
		time.Month(rgen.Intn(11)+1),
		rgen.Intn(27)+1,
		rgen.Intn(24),
		rgen.Intn(60),
		rgen.Intn(60),
		0,
		time.UTC,
	)
	return rr
}

type errorReader struct{}

func (r errorReader) Read([]byte) (int, error) {
	return 0, fmt.Errorf("an error")
}

type mockStore struct {
	saveCalled int
	saveFn     func(faststatus.Resource) error
	getCalled  int
	getFn      func(faststatus.ID) (faststatus.Resource, error)
}

func (s *mockStore) Save(r faststatus.Resource) error {
	s.saveCalled++
	return s.saveFn(r)
}

func (s *mockStore) Get(id faststatus.ID) (faststatus.Resource, error) {
	s.getCalled++
	return s.getFn(id)
}
